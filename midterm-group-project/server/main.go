package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// CommandRequest represents a command execution request
type CommandRequest struct {
	Command string
	Args    []string
	ID      string // Client ID for tracking
	Token   string // Auth token
}

// CommandResponse represents the result of command execution
type CommandResponse struct {
	Output   string
	Error    string
	ExitCode int
	ID       string
}

// HeartbeatRequest for keepalive
type HeartbeatRequest struct {
	ID    string
	Token string
}

// RegisterRequest for registering client
type RegisterRequest struct {
	ID    string
	Token string
}

// EnvRequest for set env
type EnvRequest struct {
	ID    string
	Token string
	Key   string
	Value string
}

// DirRequest for change directory
type DirRequest struct {
	ID    string
	Token string
	Dir   string
}

// ListRequest for listing clients
type ListRequest struct {
	Token string
}

type ListSessionsRequest struct {
	Token string
}

type KillSessionRequest struct {
	ID    string
	Token string
}

// UpdateWhitelistRequest for dynamic whitelist changes
type UpdateWhitelistRequest struct {
	Token    string
	Commands []string
}

// RemoteShellService is the RPC service for remote shell execution
type RemoteShellService struct {
	mu             sync.RWMutex
	sessions       map[string]*Session // Track active sessions by client ID
	sessionTimeout time.Duration       // Timeout for inactive sessions
	stopCleanup    chan bool           // Channel to stop cleanup goroutine

	// Security / limits
	authToken     string
	allowedCmds   map[string]struct{}
	rateLimit     int
	rateWindow    time.Duration
	rateCounters  map[string]*rateInfo
	maxRuntime    time.Duration
	maxOutput     int
	blockChaining bool
	banned        map[string]struct{} // Banned client IDs
}

type rateInfo struct {
	count      int
	windowFrom time.Time
}

// Session tracks a client session
type Session struct {
	ID          string
	Env         map[string]string
	WorkDir     string
	ConnectedAt time.Time
	LastActive  time.Time
}

// NewRemoteShellService creates a new remote shell service
func NewRemoteShellService(authToken string, allowedCmds map[string]struct{}, rateLimit int, rateWindow time.Duration, maxRuntime time.Duration, maxOutput int, blockChaining bool) *RemoteShellService {
	service := &RemoteShellService{
		sessions:       make(map[string]*Session),
		sessionTimeout: 30 * time.Minute, // 30 minutes timeout
		stopCleanup:    make(chan bool),
		authToken:      authToken,
		allowedCmds:    allowedCmds,
		rateLimit:      rateLimit,
		rateWindow:     rateWindow,
		rateCounters:   make(map[string]*rateInfo),
		maxRuntime:     maxRuntime,
		maxOutput:      maxOutput,
		blockChaining:  blockChaining,
		banned:         make(map[string]struct{}),
	}
	// Start background cleanup goroutine
	go service.cleanupInactiveSessions()
	return service
}

// cleanupInactiveSessions periodically removes inactive sessions
func (r *RemoteShellService) cleanupInactiveSessions() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			now := time.Now()
			for id, session := range r.sessions {
				if now.Sub(session.LastActive) > r.sessionTimeout {
					log.Printf("[Cleanup] Removing inactive session: %s (inactive for %v)", id, now.Sub(session.LastActive))
					delete(r.sessions, id)
				}
			}
			r.mu.Unlock()
		case <-r.stopCleanup:
			return
		}
	}
}

// Heartbeat updates the last active time for a client (for keepalive)
func (r *RemoteShellService) Heartbeat(req HeartbeatRequest, resp *string) error {
	if !r.validateToken(req.Token) {
		*resp = "Error: unauthorized"
		return nil
	}
	if r.isBanned(req.ID) {
		*resp = "Error: banned"
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[req.ID]
	if !exists {
		*resp = "Error: client not registered"
		return nil
	}

	session.LastActive = time.Now()
	*resp = "OK"
	return nil
}

// GetSessionInfo returns information about a client session
func (r *RemoteShellService) GetSessionInfo(clientID string, resp *map[string]interface{}) error {
	// Kept without token for backward compatibility; can be secured similarly if needed
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[clientID]
	if !exists {
		*resp = map[string]interface{}{
			"error": "Session not found",
		}
		return nil
	}

	*resp = map[string]interface{}{
		"id":           session.ID,
		"work_dir":     session.WorkDir,
		"connected_at": session.ConnectedAt.Format(time.RFC3339),
		"last_active":  session.LastActive.Format(time.RFC3339),
		"env_count":    len(session.Env),
		"is_active":    time.Since(session.LastActive) < r.sessionTimeout,
	}
	return nil
}

// Execute executes a shell command remotely
func (r *RemoteShellService) Execute(req CommandRequest, resp *CommandResponse) error {
	if !r.validateToken(req.Token) {
		resp.Error = "unauthorized"
		resp.ExitCode = -1
		return nil
	}
	if r.isBanned(req.ID) {
		resp.Error = "banned"
		resp.ExitCode = -1
		return nil
	}

	if !r.allowCommand(req.Command) {
		resp.Error = "command not allowed"
		resp.ExitCode = -1
		return nil
	}

	if !r.consumeRate(req.ID) {
		resp.Error = "rate limit exceeded"
		resp.ExitCode = -1
		return nil
	}

	if r.blockChaining && containsChaining(req.Command) {
		resp.Error = "chaining/piping is blocked"
		resp.ExitCode = -1
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Get or create session for this client
	session, exists := r.sessions[req.ID]
	if !exists {
		now := time.Now()
		session = &Session{
			ID:          req.ID,
			Env:         make(map[string]string),
			WorkDir:     getDefaultWorkDir(),
			ConnectedAt: now,
			LastActive:  now,
		}
		r.sessions[req.ID] = session
		log.Printf("[Client %s] Auto-registered on first command", req.ID)
	}

	// Prepare command with timeout context
	limit := r.maxRuntime
	if limit <= 0 {
		limit = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", req.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", req.Command)
	}

	// Set working directory
	if session.WorkDir != "" {
		cmd.Dir = session.WorkDir
	}

	// Set environment variables
	cmd.Env = os.Environ()
	for k, v := range session.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	if r.maxOutput > 0 && len(output) > r.maxOutput {
		output = output[:r.maxOutput]
	}
	
	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		resp.ID = req.ID
		resp.ExitCode = -1
		resp.Error = "Command execution timeout (5 minutes)"
		resp.Output = string(output)
		log.Printf("[Client %s] Command timeout: %s", req.ID, req.Command)
		return nil
	}
	
	resp.ID = req.ID
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitError.ExitCode()
		} else {
			resp.ExitCode = -1
		}
		resp.Error = err.Error()
		resp.Output = string(output)
	} else {
		resp.ExitCode = 0
		resp.Output = string(output)
	}

	log.Printf("[Client %s] Executed: %s (Exit: %d)", req.ID, req.Command, resp.ExitCode)
	
	// Update last active time
	if session, exists := r.sessions[req.ID]; exists {
		session.LastActive = time.Now()
	}
	
	return nil
}

// Register registers a new client session
func (r *RemoteShellService) Register(req RegisterRequest, resp *string) error {
	if !r.validateToken(req.Token) {
		*resp = "Error: unauthorized"
		return nil
	}
	if r.isBanned(req.ID) {
		*resp = "Error: banned"
		return nil
	}

	if !r.consumeRate(req.ID) {
		*resp = "Error: rate limit exceeded"
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	session, exists := r.sessions[req.ID]
	if !exists {
		session = &Session{
			ID:          req.ID,
			Env:         make(map[string]string),
			WorkDir:     getDefaultWorkDir(),
			ConnectedAt: now,
			LastActive:  now,
		}
		r.sessions[req.ID] = session
		log.Printf("[Client %s] Registered (new session)", req.ID)
		*resp = fmt.Sprintf("Client %s registered successfully", req.ID)
	} else {
		session.LastActive = now
		log.Printf("[Client %s] Re-registered (existing session)", req.ID)
		*resp = fmt.Sprintf("Client %s re-registered", req.ID)
	}

	return nil
}

// SetEnv sets an environment variable for a client session
func (r *RemoteShellService) SetEnv(req EnvRequest, resp *string) error {
	if !r.validateToken(req.Token) {
		*resp = "Error: unauthorized"
		return nil
	}
	if r.isBanned(req.ID) {
		*resp = "Error: banned"
		return nil
	}

	if !r.consumeRate(req.ID) {
		*resp = "Error: rate limit exceeded"
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	clientID := req.ID
	if clientID == "" {
		*resp = "Error: client_id required"
		return nil
	}

	session, exists := r.sessions[clientID]
	if !exists {
		now := time.Now()
		session = &Session{
			ID:          clientID,
			Env:         make(map[string]string),
			WorkDir:     getDefaultWorkDir(),
			ConnectedAt: now,
			LastActive:  now,
		}
		r.sessions[clientID] = session
		log.Printf("[Client %s] Auto-registered on SetEnv", clientID)
	}
	session.LastActive = time.Now()

	key := req.Key
	value := req.Value
	if key != "" && value != "" {
		session.Env[key] = value
		*resp = fmt.Sprintf("Set %s=%s for client %s", key, value, clientID)
	} else {
		*resp = "Error: key and value required"
	}

	return nil
}

// ChangeDir changes the working directory for a client session
func (r *RemoteShellService) ChangeDir(req DirRequest, resp *string) error {
	if !r.validateToken(req.Token) {
		*resp = "Error: unauthorized"
		return nil
	}
	if r.isBanned(req.ID) {
		*resp = "Error: banned"
		return nil
	}

	if !r.consumeRate(req.ID) {
		*resp = "Error: rate limit exceeded"
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	clientID := req.ID
	dir := req.Dir

	if clientID == "" {
		*resp = "Error: client_id required"
		return nil
	}

	session, exists := r.sessions[clientID]
	if !exists {
		now := time.Now()
		session = &Session{
			ID:          clientID,
			Env:         make(map[string]string),
			WorkDir:     getDefaultWorkDir(),
			ConnectedAt: now,
			LastActive:  now,
		}
		r.sessions[clientID] = session
		log.Printf("[Client %s] Auto-registered on ChangeDir", clientID)
	}
	session.LastActive = time.Now()

	if dir != "" {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			*resp = fmt.Sprintf("Error: directory %s does not exist", dir)
			return nil
		}
		session.WorkDir = dir
		*resp = fmt.Sprintf("Changed directory to %s for client %s", dir, clientID)
	} else {
		*resp = "Error: dir required"
	}

	return nil
}

// ListClients returns list of active client sessions
func (r *RemoteShellService) ListClients(req ListRequest, resp *[]string) error {
	if !r.validateToken(req.Token) {
		return fmt.Errorf("unauthorized")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	clients := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		clients = append(clients, id)
	}

	*resp = clients
	return nil
}

// ListSessions returns detail of sessions
func (r *RemoteShellService) ListSessions(req ListSessionsRequest, resp *[]map[string]interface{}) error {
	if !r.validateToken(req.Token) {
		return fmt.Errorf("unauthorized")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]map[string]interface{}, 0, len(r.sessions))
	now := time.Now()
	for id, s := range r.sessions {
		out = append(out, map[string]interface{}{
			"id":           id,
			"work_dir":     s.WorkDir,
			"env_count":    len(s.Env),
			"connected_at": s.ConnectedAt.Format(time.RFC3339),
			"last_active":  s.LastActive.Format(time.RFC3339),
			"age":          now.Sub(s.ConnectedAt).String(),
			"idle":         now.Sub(s.LastActive).String(),
			"is_active":    now.Sub(s.LastActive) < r.sessionTimeout,
		})
	}
	*resp = out
	return nil
}

// KillSession removes a session by ID
func (r *RemoteShellService) KillSession(req KillSessionRequest, resp *string) error {
	if !r.validateToken(req.Token) {
		*resp = "unauthorized"
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sessions[req.ID]; !ok {
		*resp = "not found"
		return nil
	}
	delete(r.sessions, req.ID)
	r.banned[req.ID] = struct{}{}
	*resp = "killed and banned"
	log.Printf("[Admin] Killed and banned session %s", req.ID)
	return nil
}

// AddToWhitelist adds commands to the allowed command whitelist
func (r *RemoteShellService) AddToWhitelist(req UpdateWhitelistRequest, resp *[]string) error {
	if !r.validateToken(req.Token) {
		return fmt.Errorf("unauthorized")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.allowedCmds == nil {
		r.allowedCmds = make(map[string]struct{})
	}

	for _, c := range req.Commands {
		trimmed := strings.TrimSpace(c)
		if trimmed == "" {
			continue
		}
		first := strings.Fields(trimmed)[0]
		if first == "" {
			continue
		}
		r.allowedCmds[first] = struct{}{}
		log.Printf("[Admin] Added to whitelist: %s", first)
	}

	// Return current whitelist for convenience
	*resp = keys(r.allowedCmds)
	return nil
}

func getDefaultWorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// validateToken checks auth token if configured
func (r *RemoteShellService) validateToken(token string) bool {
	if r.authToken == "" {
		return true // no auth configured
	}
	return token == r.authToken
}

// allowCommand checks whitelist; if empty allow all
func (r *RemoteShellService) allowCommand(cmd string) bool {
	if len(r.allowedCmds) == 0 {
		return true
	}
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return false
	}
	first := strings.Fields(trimmed)[0]
	_, ok := r.allowedCmds[first]
	return ok
}

// consumeRate applies simple fixed window rate limiting per client ID
func (r *RemoteShellService) consumeRate(id string) bool {
	if r.rateLimit <= 0 {
		return true
	}
	now := time.Now()
	info, ok := r.rateCounters[id]
	if !ok || now.Sub(info.windowFrom) > r.rateWindow {
		r.rateCounters[id] = &rateInfo{count: 1, windowFrom: now}
		return true
	}
	if info.count >= r.rateLimit {
		return false
	}
	info.count++
	return true
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func (r *RemoteShellService) isBanned(id string) bool {
	_, ok := r.banned[id]
	return ok
}

// containsChaining blocks common shell chaining tokens
func containsChaining(cmd string) bool {
	l := strings.ToLower(cmd)
	bad := []string{"|", "&&", "||", ";"}
	for _, b := range bad {
		if strings.Contains(l, b) {
			return true
		}
	}
	return false
}

func main() {
	var (
		port          = flag.Int("port", 8080, "Port to listen on")
		authToken     = flag.String("auth-token", "", "Auth token required from clients (optional)")
		allowCmdsStr  = flag.String("allow-commands", "", "Comma-separated whitelist of allowed commands (empty = allow all)")
		rateLimit     = flag.Int("rate-limit", 60, "Max requests per window per client (0 = disable)")
		rateWindowSec = flag.Int("rate-window-sec", 60, "Rate limit window in seconds")
		tlsCert       = flag.String("tls-cert", "", "Path to TLS certificate (optional)")
		tlsKey        = flag.String("tls-key", "", "Path to TLS key (optional)")
	)
	flag.Parse()

	allowed := make(map[string]struct{})
	if strings.TrimSpace(*allowCmdsStr) != "" {
		parts := strings.Split(*allowCmdsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				allowed[p] = struct{}{}
			}
		}
	}

	limit := time.Duration(*rateWindowSec) * time.Second
	service := NewRemoteShellService(*authToken, allowed, *rateLimit, time.Duration(*rateWindowSec)*time.Second, limit, 256*1024, true)
	rpc.Register(service)

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}

	// Wrap with TLS if cert/key provided
	if *tlsCert != "" && *tlsKey != "" {
		cer, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("Failed to load TLS cert/key: %v", err)
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		listener = tls.NewListener(listener, config)
		log.Printf("TLS enabled with cert %s", *tlsCert)
	}

	// Get server IP addresses for display
	log.Printf("Remote Shell RPC Server started on %s", addr)
	if *authToken != "" {
		log.Println("Auth token required for all calls")
	}
	if len(allowed) > 0 {
		log.Printf("Command whitelist enabled: %v", keys(allowed))
	}
	log.Printf("Rate limit: %d requests / %ds per client", *rateLimit, *rateWindowSec)
	log.Printf("Max runtime: %ds, Max output: %d bytes, Block chaining: %v", int(service.maxRuntime.Seconds()), service.maxOutput, service.blockChaining)
	
	// Display local IP addresses
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		log.Println("Local IP addresses:")
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					log.Printf("  - %s:8080", ipNet.IP.String())
				}
			}
		}
	}
	
	log.Println("Waiting for clients...")
	log.Printf("Clients can connect using: <server-ip>:%d", *port)

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// Set connection timeout
		conn.SetDeadline(time.Now().Add(10 * time.Minute))
		
		// Handle each client in a separate goroutine
		go func(conn net.Conn) {
			clientAddr := conn.RemoteAddr()
			log.Printf("New client connected: %s", clientAddr)
			defer conn.Close()
			
			// Serve RPC (net/rpc ServeConn does not return an error)
			rpc.ServeConn(conn)
			log.Printf("Client disconnected: %s", clientAddr)
		}(conn)
	}
}


