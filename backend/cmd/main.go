package main

import (
	"log"

	"github.com/lekeOmotayo/runtir/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize AfriOps server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Fatal: AfriOps server exited with error: %v", err)
	}
}