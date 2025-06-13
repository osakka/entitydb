package main

import (
	"entitydb/config"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	flag.Parse()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	fmt.Println("Clearing EntityDB cache by restarting the service...")
	
	// Build API endpoint URL
	var endpoint string
	if cfg.UseSSL {
		endpoint = fmt.Sprintf("https://localhost:%d/admin/shutdown", cfg.SSLPort)
	} else {
		endpoint = fmt.Sprintf("http://localhost:%d/admin/shutdown", cfg.Port)
	}
	
	// Stop the service
	fmt.Println("Stopping EntityDB service...")
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("Failed to stop service: %v", err)
	} else {
		resp.Body.Close()
		fmt.Println("Service stopped")
	}
	
	// Wait a moment
	time.Sleep(2 * time.Second)
	
	fmt.Println("Cache cleared. Please restart the service manually with:")
	fmt.Println("cd /opt/entitydb/src && ./entitydb")
}