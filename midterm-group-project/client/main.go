package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

// CommandRequest and CommandResponse must match server definitions
type CommandRequest struct {
	Command string
	Args    []string
	ID      string
}

type CommandResponse struct {
	Output   string
	Error    string
	ExitCode int
	ID       string
}

type RemoteShellClient struct {
	client     *rpc.Client
	id         string
	serverAddr string
	connected  bool
}

func NewRemoteShellClient(serverAddr string, clientID string) (*RemoteShellClient, error) {
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	return &RemoteShellClient{
		client:     client,
		id:         clientID,
		serverAddr: serverAddr,
		connected:  true,
	}, nil
}

// Reconnect attempts to reconnect to the server
func (c *RemoteShellClient) Reconnect() error {
	if c.connected {
		c.client.Close()
	}
	
	client, err := rpc.Dial("tcp", c.serverAddr)
	if err != nil {
		c.connected = false
		return fmt.Errorf("failed to reconnect: %v", err)
	}
	
	c.client = client
	c.connected = true
	return nil
}

// SendHeartbeat sends a heartbeat to keep the session alive
func (c *RemoteShellClient) SendHeartbeat() error {
	var resp string
	err := c.client.Call("RemoteShellService.Heartbeat", c.id, &resp)
	if err != nil {
		c.connected = false
		return err
	}
	return nil
}

func (c *RemoteShellClient) Execute(command string) (*CommandResponse, error) {
	req := CommandRequest{
		Command: command,
		ID:      c.id,
	}
	var resp CommandResponse

	err := c.client.Call("RemoteShellService.Execute", req, &resp)
	if err != nil {
		// Try to reconnect once
		if reconnectErr := c.Reconnect(); reconnectErr == nil {
			// Retry the call
			err = c.client.Call("RemoteShellService.Execute", req, &resp)
			if err != nil {
				return nil, fmt.Errorf("execution failed after reconnect: %v", err)
			}
		} else {
			return nil, fmt.Errorf("execution failed: %v", err)
		}
	}

	return &resp, nil
}

func (c *RemoteShellClient) SetEnv(key, value string) error {
	req := map[string]string{
		"client_id": c.id,
		"key":       key,
		"value":     value,
	}
	var resp string
	return c.client.Call("RemoteShellService.SetEnv", req, &resp)
}

func (c *RemoteShellClient) ChangeDir(dir string) error {
	req := map[string]string{
		"client_id": c.id,
		"dir":       dir,
	}
	var resp string
	return c.client.Call("RemoteShellService.ChangeDir", req, &resp)
}

func (c *RemoteShellClient) Register() error {
	var resp string
	return c.client.Call("RemoteShellService.Register", c.id, &resp)
}

func (c *RemoteShellClient) Close() error {
	return c.client.Close()
}

func main() {
	var (
		serverAddr = flag.String("server", "localhost:8080", "RPC server address")
		clientID   = flag.String("id", "", "Client ID (required)")
		command    = flag.String("cmd", "", "Command to execute (optional, if not provided, enters interactive mode)")
	)
	flag.Parse()

	if *clientID == "" {
		// Generate a unique client ID
		*clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
		fmt.Printf("Generated client ID: %s\n", *clientID)
	}

	// Connect to server
	shellClient, err := NewRemoteShellClient(*serverAddr, *clientID)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer shellClient.Close()

	// Register client with server
	err = shellClient.Register()
	if err != nil {
		log.Printf("Warning: Failed to register client: %v", err)
	}

	fmt.Printf("Connected to server %s as %s\n", *serverAddr, *clientID)
	fmt.Println("Type 'exit' to quit, 'help' for commands")

	// If command provided, execute and exit
	if *command != "" {
		resp, err := shellClient.Execute(*command)
		if err != nil {
			log.Fatal("Error executing command:", err)
		}

		if resp.ExitCode != 0 {
			fmt.Fprintf(os.Stderr, "%s\n", resp.Error)
			os.Exit(resp.ExitCode)
		}
		fmt.Print(resp.Output)
		return
	}

	// Start heartbeat goroutine to keep session alive
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Send heartbeat every minute
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := shellClient.SendHeartbeat(); err != nil {
					log.Printf("Heartbeat failed: %v", err)
				}
			}
		}
	}()

	// Interactive mode
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("[%s@remote]$ ", *clientID)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Handle special commands
		if line == "exit" {
			break
		}

		if line == "help" {
			fmt.Println("Available commands:")
			fmt.Println("  exit              - Exit the client")
			fmt.Println("  help              - Show this help")
			fmt.Println("  cd <dir>          - Change directory")
			fmt.Println("  setenv <k> <v>    - Set environment variable")
			fmt.Println("  <command>         - Execute shell command")
			continue
		}

		// Handle cd command
		if strings.HasPrefix(line, "cd ") {
			dir := strings.TrimSpace(line[3:])
			err := shellClient.ChangeDir(dir)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Directory changed")
			}
			continue
		}

		// Handle setenv command
		if strings.HasPrefix(line, "setenv ") {
			parts := strings.Fields(line[7:])
			if len(parts) != 2 {
				fmt.Println("Usage: setenv <key> <value>")
				continue
			}
			err := shellClient.SetEnv(parts[0], parts[1])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Set %s=%s\n", parts[0], parts[1])
			}
			continue
		}

		// Execute command
		resp, err := shellClient.Execute(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if resp.ExitCode != 0 {
			fmt.Fprintf(os.Stderr, "Exit code: %d\n", resp.ExitCode)
			if resp.Error != "" {
				fmt.Fprintf(os.Stderr, "%s\n", resp.Error)
			}
		}

		if resp.Output != "" {
			fmt.Print(resp.Output)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}

	fmt.Println("Goodbye!")
}


