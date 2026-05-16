package main

import (
	"log"

	"cloud-agent-runtime/backend"
)

func main() {
	manager, err := backend.NewSessionManager()
	if err != nil {
		log.Fatalf("failed to init session manager: %v", err)
	}
	server := backend.NewServer(manager)
	if err := server.Router().Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
