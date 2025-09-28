package main

import (
	"log"
	"metrics-sidecard/internal/config"
	"metrics-sidecard/internal/server"
)

func main() {
	cfg := config.LoadFromCLI()

	srv := server.New(cfg)

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start metrics sidecard: %v", err)
	}
}
