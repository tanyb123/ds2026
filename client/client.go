package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "remote-shell-rpc/proto"
)

// Client represents the RPC client
type Client struct {
	conn   *grpc.ClientConn
	client pb.ShellServiceClient
	addr   string
}

// NewClient creates a new RPC client
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewShellServiceClient(conn),
		addr:   serverAddr,
	}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// ExecuteCommand executes a single command
func (c *Client) ExecuteCommand(ctx context.Context, cmd string, args []string, workDir string, env map[string]string, sessionID string) (*pb.CommandResponse, error) {
	req := &pb.CommandRequest{
		Command:   cmd,
		Args:      args,
		WorkingDir: workDir,
		Env:       env,
		SessionId: sessionID,
	}

	resp, err := c.client.ExecuteCommand(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExecuteCommandStream executes a command and streams output
func (c *Client) ExecuteCommandStream(ctx context.Context, cmd string, args []string, workDir string, env map[string]string, sessionID string) error {
	req := &pb.CommandRequest{
		Command:   cmd,
		Args:      args,
		WorkingDir: workDir,
		Env:       env,
		SessionId: sessionID,
	}

	stream, err := c.client.ExecuteCommandStream(ctx, req)
	if err != nil {
		return err
	}

	for {
		output, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if output.IsEof {
			if output.ExitCode != 0 {
				os.Exit(int(output.ExitCode))
			}
			break
		}

		if output.IsStderr {
			fmt.Fprint(os.Stderr, output.Data)
		} else {
			fmt.Fprint(os.Stdout, output.Data)
		}
	}

	return nil
}

// InteractiveShell starts an interactive shell session
func (c *Client) InteractiveShell(ctx context.Context) error {
	stream, err := c.client.InteractiveShell(ctx)
	if err != nil {
		return err
	}

	// Handle output from server
	outputDone := make(chan bool)
	go func() {
		for {
			output, err := stream.Recv()
			if err == io.EOF {
				outputDone <- true
				return
			}
			if err != nil {
				log.Printf("Error receiving: %v", err)
				outputDone <- true
				return
			}

			if output.IsEof {
				outputDone <- true
				return
			}

			if output.IsStderr {
				fmt.Fprint(os.Stderr, output.Output)
			} else {
				fmt.Fprint(os.Stdout, output.Output)
			}
		}
	}()

	// Handle input from stdin
	inputDone := make(chan bool)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err == io.EOF {
				stream.Send(&pb.ShellInput{
					IsEof: true,
				})
				inputDone <- true
				return
			}
			if err != nil {
				log.Printf("Error reading stdin: %v", err)
				inputDone <- true
				return
			}

			if err := stream.Send(&pb.ShellInput{
				Input: string(buf[:n]),
			}); err != nil {
				log.Printf("Error sending: %v", err)
				inputDone <- true
				return
			}
		}
	}()

	// Wait for either input or output to finish
	select {
	case <-outputDone:
	case <-inputDone:
	case <-ctx.Done():
	}

	return nil
}

// ListSessions lists all active sessions
func (c *Client) ListSessions(ctx context.Context) (*pb.SessionList, error) {
	req := &pb.Empty{}
	return c.client.ListSessions(ctx, req)
}

// KillSession kills a session
func (c *Client) KillSession(ctx context.Context, sessionID string) (*pb.SessionResponse, error) {
	req := &pb.SessionRequest{
		SessionId: sessionID,
	}
	return c.client.KillSession(ctx, req)
}

