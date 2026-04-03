package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/pepperonas/go-sling/internal/config"
	"github.com/pepperonas/go-sling/internal/server"
	"github.com/pepperonas/go-sling/internal/storage"
	"github.com/pepperonas/go-sling/internal/ws"
)

//go:embed web/*
var webContent embed.FS

var version = "1.0.0"

func main() {
	var (
		port       int
		pin        string
		dataDir    string
		retention  int
		configFile string
		showVer    bool
	)

	flag.IntVar(&port, "port", 0, "server port (default: 8420)")
	flag.StringVar(&pin, "pin", "", "authentication PIN")
	flag.StringVar(&dataDir, "data-dir", "", "data storage directory (default: ./data)")
	flag.IntVar(&retention, "retention", 0, "file retention in hours (default: 24)")
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
	flag.BoolVar(&showVer, "version", false, "show version")
	flag.Parse()

	if showVer {
		fmt.Printf("go-sling v%s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// CLI flag overrides
	if port > 0 {
		cfg.Server.Port = port
	}
	if pin != "" {
		cfg.Auth.Pin = pin
	}
	if dataDir != "" {
		cfg.Storage.DataDir = dataDir
	}
	if retention > 0 {
		cfg.Storage.RetentionHours = retention
	}

	// Create default config if it doesn't exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg.WriteDefault(configFile)
	}

	// Initialize storage
	store, err := storage.New(cfg.Storage.DataDir, cfg.RetentionDuration(), cfg.Storage.MaxUploadSize)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Start auto-cleanup
	if cfg.Storage.AutoCleanup {
		store.StartCleanup(10 * time.Minute)
	}

	// WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Embedded web filesystem
	webFS, err := fs.Sub(webContent, "web")
	if err != nil {
		log.Fatalf("Failed to get web filesystem: %v", err)
	}

	// HTTP server
	srv := server.New(cfg, webFS)
	srv.RegisterRoutes(store, hub)

	log.Fatal(srv.Start())
}
