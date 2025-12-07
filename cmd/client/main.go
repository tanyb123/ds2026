package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"remote-shell-rpc/client"
	pb "remote-shell-rpc/proto"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "Server address")
	command := flag.String("command", "", "Command to execute")
	interactive := flag.Bool("interactive", false, "Start interactive shell")
	listSessions := flag.Bool("list-sessions", false, "List active sessions")
	killSession := flag.String("kill-session", "", "Kill a session by ID")
	flag.Parse()

	// Create client
	cl, err := client.NewClient(*serverAddr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Handle different modes
	if *listSessions {
		sessions, err := cl.ListSessions(ctx)
		if err != nil {
			log.Fatalf("Failed to list sessions: %v", err)
		}

		fmt.Println("Active Sessions:")
		fmt.Println("===============")
		for _, sess := range sessions.Sessions {
			fmt.Printf("ID: %s\n", sess.SessionId)
			fmt.Printf("  Client: %s\n", sess.ClientAddress)
			fmt.Printf("  Created: %s\n", time.Unix(sess.CreatedAt, 0).Format(time.RFC3339))
			fmt.Printf("  Active: %v\n", sess.IsActive)
			fmt.Println()
		}
		return
	}

	if *killSession != "" {
		resp, err := cl.KillSession(ctx, *killSession)
		if err != nil {
			log.Fatalf("Failed to kill session: %v", err)
		}

		if resp.Success {
			fmt.Printf("Session %s killed successfully\n", *killSession)
		} else {
			fmt.Printf("Failed to kill session: %s\n", resp.Message)
		}
		return
	}

	if *interactive {
		fmt.Println("Starting interactive shell...")
		fmt.Println("Type 'exit' to quit")
		fmt.Println()

		ctx := context.Background()
		if err := cl.InteractiveShell(ctx); err != nil {
			log.Fatalf("Interactive shell error: %v", err)
		}
		return
	}

	if *command == "" {
		fmt.Println("Usage:")
		fmt.Println("  Execute command:  --command 'ls -la'")
		fmt.Println("  Interactive mode: --interactive")
		fmt.Println("  List sessions:    --list-sessions")
		fmt.Println("  Kill session:    --kill-session <id>")
		os.Exit(1)
	}

	// Parse command
	parts := strings.Fields(*command)
	if len(parts) == 0 {
		log.Fatalf("Invalid command")
	}

	cmd := parts[0]
	args := parts[1:]

	// Execute command with streaming
	if err := cl.ExecuteCommandStream(ctx, cmd, args, "", nil, ""); err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
}

