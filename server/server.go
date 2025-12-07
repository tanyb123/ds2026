package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"

	"remote-shell-rpc/internal/executor"
	"remote-shell-rpc/internal/session"
	pb "remote-shell-rpc/proto"
)

// Server represents the RPC server
type Server struct {
	pb.UnimplementedShellServiceServer
	grpcServer  *grpc.Server
	sessionMgr  *session.Manager
	executor     *executor.Executor
	port         int
}

// NewServer creates a new RPC server
func NewServer(port int) *Server {
	return &Server{
		sessionMgr: session.NewManager(),
		executor:   executor.NewExecutor(),
		port:       port,
	}
}

// Start starts the RPC server
func (s *Server) Start() error {
	// Create gRPC server
	s.grpcServer = grpc.NewServer()

	// Register service
	pb.RegisterShellServiceServer(s.grpcServer, s)

	// Listen on port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Printf("Server starting on port %d...", s.port)

	// Handle graceful shutdown
	go s.handleShutdown()

	// Serve
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

// handleShutdown handles graceful shutdown
func (s *Server) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	s.grpcServer.GracefulStop()
	log.Println("Server stopped")
}

// ExecuteCommand implements the ExecuteCommand RPC
func (s *Server) ExecuteCommand(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	// Get client address
	clientAddr := getClientAddr(ctx)

	// Create or get session
	var sess *session.Session
	if req.SessionId != "" {
		var exists bool
		sess, exists = s.sessionMgr.GetSession(req.SessionId)
		if !exists {
			return nil, fmt.Errorf("session not found: %s", req.SessionId)
		}
	} else {
		sess = s.sessionMgr.CreateSession(clientAddr)
	}

	s.sessionMgr.UpdateActivity(sess.ID)

	// Parse command
	cmd, args := parseCommand(req.Command, req.Args)

	// Execute command
	result, err := s.executor.Execute(ctx, cmd, args, req.WorkingDir, req.Env)
	if err != nil {
		return &pb.CommandResponse{
			ExitCode:  -1,
			Stderr:    err.Error(),
			SessionId: sess.ID,
		}, nil
	}

	return &pb.CommandResponse{
		ExitCode:        int32(result.ExitCode),
		Stdout:          result.Stdout,
		Stderr:          result.Stderr,
		SessionId:       sess.ID,
		ExecutionTimeMs: result.ExecutionTime.Milliseconds(),
	}, nil
}

// ExecuteCommandStream implements the ExecuteCommandStream RPC
func (s *Server) ExecuteCommandStream(req *pb.CommandRequest, stream pb.ShellService_ExecuteCommandStreamServer) error {
	ctx := stream.Context()
	clientAddr := getClientAddr(ctx)

	// Create or get session
	var sess *session.Session
	if req.SessionId != "" {
		var exists bool
		sess, exists = s.sessionMgr.GetSession(req.SessionId)
		if !exists {
			return fmt.Errorf("session not found: %s", req.SessionId)
		}
	} else {
		sess = s.sessionMgr.CreateSession(clientAddr)
	}

	s.sessionMgr.UpdateActivity(sess.ID)

	// Parse command
	cmd, args := parseCommand(req.Command, req.Args)

	// Create output channel
	outputChan := make(chan *executor.Output, 100)

	// Execute command in goroutine
	go func() {
		if err := s.executor.ExecuteStream(ctx, cmd, args, req.WorkingDir, req.Env, outputChan); err != nil {
			stream.Send(&pb.CommandOutput{
				Data:    fmt.Sprintf("Error: %v\n", err),
				IsStderr: true,
				IsEof:    true,
			})
		}
	}()

	// Stream output
	for output := range outputChan {
		if err := stream.Send(&pb.CommandOutput{
			Data:     output.Data,
			IsStderr: output.IsStderr,
			IsEof:    output.IsEOF,
			ExitCode: int32(output.ExitCode),
		}); err != nil {
			return err
		}

		if output.IsEOF {
			break
		}
	}

	return nil
}

// InteractiveShell implements the InteractiveShell RPC
func (s *Server) InteractiveShell(stream pb.ShellService_InteractiveShellServer) error {
	ctx := stream.Context()
	clientAddr := getClientAddr(ctx)

	// Create session
	sess := s.sessionMgr.CreateSession(clientAddr)
	defer s.sessionMgr.KillSession(sess.ID)

	// Create input/output channels
	inputChan := make(chan string, 100)
	outputChan := make(chan *executor.Output, 100)

	// Handle input from client
	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				close(inputChan)
				return
			}

			if req.IsEof {
				close(inputChan)
				return
			}

			inputChan <- req.Input
		}
	}()

	// Handle output to client
	go func() {
		for output := range outputChan {
			if err := stream.Send(&pb.ShellOutput{
				SessionId: sess.ID,
				Output:    output.Data,
				IsStderr:  output.IsStderr,
				IsEof:     output.IsEOF,
			}); err != nil {
				return
			}

			if output.IsEOF {
				return
			}
		}
	}()

	// TODO: Implement interactive shell loop
	// For now, just echo back
	for input := range inputChan {
		outputChan <- &executor.Output{
			Data:    fmt.Sprintf("Echo: %s", input),
			IsStderr: false,
			IsEOF:   false,
		}
	}

	return nil
}

// ListSessions implements the ListSessions RPC
func (s *Server) ListSessions(ctx context.Context, req *pb.Empty) (*pb.SessionList, error) {
	sessions := s.sessionMgr.ListSessions()

	sessionInfos := make([]*pb.SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		sess.mu.RLock()
		sessionInfos = append(sessionInfos, &pb.SessionInfo{
			SessionId:     sess.ID,
			ClientAddress: sess.ClientAddr,
			CreatedAt:     sess.CreatedAt.Unix(),
			IsActive:      sess.IsActive,
		})
		sess.mu.RUnlock()
	}

	return &pb.SessionList{
		Sessions: sessionInfos,
	}, nil
}

// KillSession implements the KillSession RPC
func (s *Server) KillSession(ctx context.Context, req *pb.SessionRequest) (*pb.SessionResponse, error) {
	success := s.sessionMgr.KillSession(req.SessionId)
	if !success {
		return &pb.SessionResponse{
			Success: false,
			Message: fmt.Sprintf("Session not found: %s", req.SessionId),
		}, nil
	}

	return &pb.SessionResponse{
		Success: true,
		Message: fmt.Sprintf("Session %s killed", req.SessionId),
	}, nil
}

// Helper functions

func getClientAddr(ctx context.Context) string {
	// Extract client address from context
	// This is a simplified version
	return "unknown"
}

func parseCommand(cmd string, args []string) (string, []string) {
	if len(args) == 0 {
		// Try to parse command string
		parts := strings.Fields(cmd)
		if len(parts) > 0 {
			return parts[0], parts[1:]
		}
		return cmd, nil
	}
	return cmd, args
}

