package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "practical-work-02-rpc-file-transfer/proto"
)

const (
	DEFAULT_PORT = 50051
	STORAGE_DIR = "./storage"
	CHUNK_SIZE  = 4096
)

type server struct {
	pb.UnimplementedFileTransferServiceServer
	storageDir string
	files      map[string]*fileInfo
	mu         sync.RWMutex
}

type fileInfo struct {
	path    string
	size    int64
	modTime time.Time
}

func main() {
	port := flag.Int("port", DEFAULT_PORT, "Server port")
	storageDir := flag.String("dir", STORAGE_DIR, "Storage directory")
	flag.Parse()

	// Create storage directory
	if err := os.MkdirAll(*storageDir, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Initialize server
	s := &server{
		storageDir: *storageDir,
		files:      make(map[string]*fileInfo),
	}

	// Scan existing files
	s.scanFiles()

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterFileTransferServiceServer(grpcServer, s)

	// Listen
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("RPC File Transfer Server listening on port %d", *port)
	log.Printf("Storage directory: %s", *storageDir)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// scanFiles scans the storage directory for existing files
func (s *server) scanFiles() {
	entries, err := os.ReadDir(s.storageDir)
	if err != nil {
		log.Printf("Error scanning directory: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			s.files[entry.Name()] = &fileInfo{
				path:    filepath.Join(s.storageDir, entry.Name()),
				size:    info.Size(),
				modTime: info.ModTime(),
			}
		}
	}
}

// SendFile receives a file from client (streaming)
func (s *server) SendFile(stream pb.FileTransferService_SendFileServer) error {
	var filename string
	var file *os.File
	var totalReceived int64
	var totalSize int64

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			// File transfer complete
			if file != nil {
				file.Close()
				s.mu.Lock()
				s.files[filename] = &fileInfo{
					path:    filepath.Join(s.storageDir, filename),
					size:    totalReceived,
					modTime: time.Now(),
				}
				s.mu.Unlock()
			}
			return stream.SendAndClose(&pb.FileResponse{
				Success:  true,
				Message:  fmt.Sprintf("File %s received successfully (%d bytes)", filename, totalReceived),
				FileSize: totalReceived,
			})
		}
		if err != nil {
			if file != nil {
				file.Close()
				os.Remove(filepath.Join(s.storageDir, filename))
			}
			return err
		}

		// First chunk: create file
		if file == nil {
			filename = chunk.Filename
			if filename == "" {
				return fmt.Errorf("filename is required")
			}

			// Sanitize filename
			filename = filepath.Base(filename)
			if filename == "" || filename == "." || filename == ".." {
				return fmt.Errorf("invalid filename")
			}

			filePath := filepath.Join(s.storageDir, filename)
			var err error
			file, err = os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			totalSize = chunk.TotalSize
			log.Printf("Receiving file: %s (expected size: %d bytes)", filename, totalSize)
		}

		// Write chunk to file
		if len(chunk.Data) > 0 {
			n, err := file.Write(chunk.Data)
			if err != nil {
				file.Close()
				os.Remove(filepath.Join(s.storageDir, filename))
				return fmt.Errorf("failed to write chunk: %v", err)
			}
			totalReceived += int64(n)
		}

		// Check if last chunk
		if chunk.IsLast {
			file.Close()
			s.mu.Lock()
			s.files[filename] = &fileInfo{
				path:    filepath.Join(s.storageDir, filename),
				size:    totalReceived,
				modTime: time.Now(),
			}
			s.mu.Unlock()
			log.Printf("File received: %s (%d bytes)", filename, totalReceived)
			return stream.SendAndClose(&pb.FileResponse{
				Success:  true,
				Message:  fmt.Sprintf("File %s received successfully", filename),
				FileSize: totalReceived,
			})
		}
	}
}

// ReceiveFile sends a file to client (streaming)
func (s *server) ReceiveFile(req *pb.FileRequest, stream pb.FileTransferService_ReceiveFileServer) error {
	filename := filepath.Base(req.Filename)
	if filename == "" || filename == "." || filename == ".." {
		return fmt.Errorf("invalid filename")
	}

	s.mu.RLock()
	info, exists := s.files[filename]
	s.mu.RUnlock()

	if !exists {
		// Try to find file in storage
		filePath := filepath.Join(s.storageDir, filename)
		stat, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("file not found: %s", filename)
		}
		info = &fileInfo{
			path:    filePath,
			size:    stat.Size(),
			modTime: stat.ModTime(),
		}
	}

	// Open file
	file, err := os.Open(info.path)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Seek to offset if specified
	if req.Offset > 0 {
		if _, err := file.Seek(req.Offset, 0); err != nil {
			return fmt.Errorf("failed to seek: %v", err)
		}
	}

	log.Printf("Sending file: %s (%d bytes)", filename, info.size)

	// Send file in chunks
	buffer := make([]byte, CHUNK_SIZE)
	var offset int64 = req.Offset

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		chunk := &pb.FileChunk{
			Filename:  filename,
			Data:      buffer[:n],
			Offset:    offset,
			TotalSize: info.size,
			IsLast:    offset+int64(n) >= info.size,
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %v", err)
		}

		offset += int64(n)
	}

	return nil
}

// ListFiles lists all available files
func (s *server) ListFiles(ctx context.Context, req *pb.Empty) (*pb.FileList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files := make([]*pb.FileInfo, 0, len(s.files))
	for name, info := range s.files {
		// Verify file still exists
		stat, err := os.Stat(info.path)
		if err != nil {
			continue
		}

		files = append(files, &pb.FileInfo{
			Filename:    name,
			Size:        stat.Size(),
			ModifiedTime: stat.ModTime().Unix(),
		})
	}

	return &pb.FileList{Files: files}, nil
}

// DeleteFile deletes a file
func (s *server) DeleteFile(ctx context.Context, req *pb.FileRequest) (*pb.FileResponse, error) {
	filename := filepath.Base(req.Filename)
	if filename == "" || filename == "." || filename == ".." {
		return &pb.FileResponse{
			Success: false,
			Message: "invalid filename",
		}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.files[filename]
	if !exists {
		return &pb.FileResponse{
			Success: false,
			Message: fmt.Sprintf("file not found: %s", filename),
		}, nil
	}

	if err := os.Remove(info.path); err != nil {
		return &pb.FileResponse{
			Success: false,
			Message: fmt.Sprintf("failed to delete: %v", err),
		}, nil
	}

	delete(s.files, filename)
	log.Printf("File deleted: %s", filename)

	return &pb.FileResponse{
		Success: true,
		Message: fmt.Sprintf("File %s deleted successfully", filename),
	}, nil
}

