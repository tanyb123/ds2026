package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
)

type ListRequest struct {
	Token string
}

func main() {
	var serverAddr = flag.String("server", "localhost:8080", "RPC server address")
	var token = flag.String("token", "", "Auth token (if server requires)")
	flag.Parse()

	// Connect to server
	client, err := rpc.Dial("tcp", *serverAddr)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer client.Close()

	// List active clients
	var clients []string
	req := ListRequest{Token: *token}
	err = client.Call("RemoteShellService.ListClients", req, &clients)
	if err != nil {
		log.Fatal("Error listing clients:", err)
	}

	fmt.Printf("Active clients (%d):\n", len(clients))
	for i, id := range clients {
		fmt.Printf("  %d. %s\n", i+1, id)
	}
}




