package main

import (
	"log"
	"net/http"

	"cloud-agent-runtime/backend"
)

func main() {
	manager, err := backend.NewSessionManager()
	if err != nil {
		log.Fatalf("failed to init session manager: %v", err)
	}
	server := backend.NewServer(manager)
	if err := http.ListenAndServe(":8080", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
