package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// CommandRequest represents a command execution request
type CommandRequest struct {
	Command string
	Args    []string
	ID      string // Client ID for tracking
}

// CommandResponse represents the result of command execution
type CommandResponse struct {
	Output   string
	Error    string
	ExitCode int
	ID       string
}

// RemoteShellService is the RPC service for remote shell execution
type RemoteShellService struct {
	mu            sync.RWMutex
	sessions      map[string]*Session // Track active sessions by client ID
	sessionTimeout time.Duration      // Timeout for inactive sessions
	stopCleanup   chan bool           // Channel to stop cleanup goroutine
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
func NewRemoteShellService() *RemoteShellService {
	service := &RemoteShellService{
		sessions:       make(map[string]*Session),
		sessionTimeout: 30 * time.Minute, // 30 minutes timeout
		stopCleanup:    make(chan bool),
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
func (r *RemoteShellService) Heartbeat(clientID string, resp *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[clientID]
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
func (r *RemoteShellService) Register(clientID string, resp *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	session, exists := r.sessions[clientID]
	if !exists {
		session = &Session{
			ID:          clientID,
			Env:         make(map[string]string),
			WorkDir:     getDefaultWorkDir(),
			ConnectedAt: now,
			LastActive:  now,
		}
		r.sessions[clientID] = session
		log.Printf("[Client %s] Registered (new session)", clientID)
		*resp = fmt.Sprintf("Client %s registered successfully", clientID)
	} else {
		session.LastActive = now
		log.Printf("[Client %s] Re-registered (existing session)", clientID)
		*resp = fmt.Sprintf("Client %s re-registered", clientID)
	}

	return nil
}

// SetEnv sets an environment variable for a client session
func (r *RemoteShellService) SetEnv(req map[string]string, resp *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clientID := req["client_id"]
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

	key := req["key"]
	value := req["value"]
	if key != "" && value != "" {
		session.Env[key] = value
		*resp = fmt.Sprintf("Set %s=%s for client %s", key, value, clientID)
	} else {
		*resp = "Error: key and value required"
	}

	return nil
}

// ChangeDir changes the working directory for a client session
func (r *RemoteShellService) ChangeDir(req map[string]string, resp *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clientID := req["client_id"]
	dir := req["dir"]

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
func (r *RemoteShellService) ListClients(req string, resp *[]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clients := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		clients = append(clients, id)
	}

	*resp = clients
	return nil
}

func getDefaultWorkDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func main() {
	// Create and register RPC service
	service := NewRemoteShellService()
	rpc.Register(service)

	// Start RPC server (bind to all interfaces: 0.0.0.0:8080)
	// This allows connections from both local network and internet
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error starting server:", err)
	}

	// Get server IP addresses for display
	log.Println("Remote Shell RPC Server started on :8080")
	log.Println("Server is listening on all network interfaces (0.0.0.0:8080)")
	
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
	log.Println("Clients can connect using: <server-ip>:8080")

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
			
			// Serve RPC with error handling
			if err := rpc.ServeConn(conn); err != nil {
				log.Printf("RPC error for client %s: %v", clientAddr, err)
			}
			log.Printf("Client disconnected: %s", clientAddr)
		}(conn)
	}
}


