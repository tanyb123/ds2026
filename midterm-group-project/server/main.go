package main

import (
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
	mu       sync.Mutex
	sessions map[string]*Session // Track active sessions by client ID
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
	return &RemoteShellService{
		sessions: make(map[string]*Session),
	}
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

	// Prepare command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", req.Command)
	} else {
		cmd = exec.Command("sh", "-c", req.Command)
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

		// Handle each client in a separate goroutine
		go func(conn net.Conn) {
			log.Printf("New client connected: %s", conn.RemoteAddr())
			rpc.ServeConn(conn)
			log.Printf("Client disconnected: %s", conn.RemoteAddr())
		}(conn)
	}
}


