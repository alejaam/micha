package main

import (
	"log"

	httpadapter "micha/backend/internal/adapters/http"
	"micha/backend/internal/infrastructure/config"
)

func main() {
	cfg := config.Load()
	server := httpadapter.NewServer(cfg.HTTPPort)

	log.Printf("api listening on :%s", cfg.HTTPPort)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
