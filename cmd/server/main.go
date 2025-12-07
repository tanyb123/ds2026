package main

import (
	"flag"
	"log"

	"remote-shell-rpc/server"
)

func main() {
	port := flag.Int("port", 50051, "Server port")
	flag.Parse()

	srv := server.NewServer(*port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

