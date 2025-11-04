package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/internal/server"
	"server/internal/server/clients"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	DatabaseURL string
	Port        int
}

var (
	defaultConfig = &config{Port: 8080}
	configPath    = flag.String("config", ".env", "Path to the config file")
)

func loadConfig() *config {
	cfg := defaultConfig
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Printf("Error parsing PORT, using %d", cfg.Port)
		return cfg
	}
	cfg.Port = port
	return cfg
}

func main() {
	flag.Parse()
	err := godotenv.Load(*configPath)
	cfg := defaultConfig
	if err != nil {
		log.Printf("Error loading config file, defaulting to %+v", defaultConfig)
	} else {
		cfg = loadConfig()
	}

	// Validate DATABASE_URL is set
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	hub := server.NewHub(cfg.DatabaseURL)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	go hub.Run()
	addr := fmt.Sprintf(":%d", cfg.Port)

	log.Printf("Starting server on %s", addr)

	err = http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
