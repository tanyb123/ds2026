package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "practical-work-02-rpc-file-transfer/proto"
)

const (
	CHUNK_SIZE = 4096
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "Server address")
	sendFile := flag.String("send-file", "", "File to send")
	receiveFile := flag.String("receive-file", "", "File to receive")
	listFiles := flag.Bool("list", false, "List files on server")
	deleteFile := flag.String("delete", "", "File to delete")
	flag.Parse()

	// Connect to server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileTransferServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Handle different operations
	if *listFiles {
		if err := listFilesOnServer(ctx, client); err != nil {
			log.Fatalf("Failed to list files: %v", err)
		}
	} else if *deleteFile != "" {
		if err := deleteFileOnServer(ctx, client, *deleteFile); err != nil {
			log.Fatalf("Failed to delete file: %v", err)
		}
	} else if *sendFile != "" {
		if err := sendFileToServer(ctx, client, *sendFile); err != nil {
			log.Fatalf("Failed to send file: %v", err)
		}
	} else if *receiveFile != "" {
		if err := receiveFileFromServer(ctx, client, *receiveFile); err != nil {
			log.Fatalf("Failed to receive file: %v", err)
		}
	} else {
		fmt.Println("Usage:")
		fmt.Println("  Send file:    --send-file <filepath>")
		fmt.Println("  Receive file: --receive-file <filename>")
		fmt.Println("  List files:   --list")
		fmt.Println("  Delete file:  --delete <filename>")
		os.Exit(1)
	}
}

func sendFileToServer(ctx context.Context, client pb.FileTransferServiceClient, filepath string) error {
	// Open file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	filename := filepath.Base(filepath)
	filesize := fileInfo.Size()

	log.Printf("Sending file: %s (%d bytes)", filename, filesize)

	// Create stream
	stream, err := client.SendFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %v", err)
	}

	// Send file in chunks
	buffer := make([]byte, CHUNK_SIZE)
	var offset int64

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
			TotalSize: filesize,
			IsLast:    offset+int64(n) >= filesize,
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %v", err)
		}

		offset += int64(n)
	}

	// Close and receive response
	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close stream: %v", err)
	}

	if response.Success {
		log.Printf("File sent successfully: %s", response.Message)
	} else {
		return fmt.Errorf("server error: %s", response.Message)
	}

	return nil
}

func receiveFileFromServer(ctx context.Context, client pb.FileTransferServiceClient, filename string) error {
	req := &pb.FileRequest{
		Filename: filename,
	}

	log.Printf("Receiving file: %s", filename)

	// Create stream
	stream, err := client.ReceiveFile(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create stream: %v", err)
	}

	// Create output file
	outputFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outputFile.Close()

	// Receive chunks
	var totalReceived int64
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %v", err)
		}

		// Write chunk to file
		if len(chunk.Data) > 0 {
			n, err := outputFile.Write(chunk.Data)
			if err != nil {
				return fmt.Errorf("failed to write chunk: %v", err)
			}
			totalReceived += int64(n)
		}

		if chunk.IsLast {
			break
		}
	}

	log.Printf("File received: %s (%d bytes)", filename, totalReceived)
	return nil
}

func listFilesOnServer(ctx context.Context, client pb.FileTransferServiceClient) error {
	req := &pb.Empty{}
	list, err := client.ListFiles(ctx, req)
	if err != nil {
		return err
	}

	fmt.Println("Files on server:")
	fmt.Println("===============")
	for _, file := range list.Files {
		fmt.Printf("  %s - %d bytes - %s\n",
			file.Filename,
			file.Size,
			time.Unix(file.ModifiedTime, 0).Format(time.RFC3339))
	}

	return nil
}

func deleteFileOnServer(ctx context.Context, client pb.FileTransferServiceClient, filename string) error {
	req := &pb.FileRequest{
		Filename: filename,
	}

	response, err := client.DeleteFile(ctx, req)
	if err != nil {
		return err
	}

	if response.Success {
		log.Printf("File deleted: %s", response.Message)
	} else {
		return fmt.Errorf("failed to delete: %s", response.Message)
	}

	return nil
}

