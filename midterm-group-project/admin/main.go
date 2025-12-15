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

type ListSessionsRequest struct {
	Token string
}

type KillSessionRequest struct {
	ID    string
	Token string
}

func main() {
	var serverAddr = flag.String("server", "localhost:8080", "RPC server address")
	var token = flag.String("token", "", "Auth token (if server requires)")
	var killID = flag.String("kill", "", "Kill session by client ID")
	var listSessions = flag.Bool("sessions", false, "List sessions with details")
	flag.Parse()

	// Connect to server
	client, err := rpc.Dial("tcp", *serverAddr)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer client.Close()

	if *killID != "" {
		var resp string
		req := KillSessionRequest{ID: *killID, Token: *token}
		err := client.Call("RemoteShellService.KillSession", req, &resp)
		if err != nil {
			log.Fatal("Error killing session:", err)
		}
		fmt.Printf("Kill session %s: %s\n", *killID, resp)
		return
	}

	if *listSessions {
		var sessions []map[string]interface{}
		req := ListSessionsRequest{Token: *token}
		err := client.Call("RemoteShellService.ListSessions", req, &sessions)
		if err != nil {
			log.Fatal("Error listing sessions:", err)
		}
		fmt.Printf("Sessions (%d):\n", len(sessions))
		for i, s := range sessions {
			fmt.Printf("  %d. id=%v workdir=%v env=%v last=%v idle=%v\n",
				i+1,
				s["id"], s["work_dir"], s["env_count"], s["last_active"], s["idle"])
		}
		return
	}

	// Default: list active clients
	var clients []string
	req := ListRequest{Token: *token}
	if err := client.Call("RemoteShellService.ListClients", req, &clients); err != nil {
		log.Fatal("Error listing clients:", err)
	}
	fmt.Printf("Active clients (%d):\n", len(clients))
	for i, id := range clients {
		fmt.Printf("  %d. %s\n", i+1, id)
	}
}




