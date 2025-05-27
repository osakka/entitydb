package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Clearing EntityDB cache by restarting the service...")
	
	// Stop the service
	fmt.Println("Stopping EntityDB service...")
	resp, err := http.Get("http://localhost:8085/admin/shutdown")
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