//go:build tool
package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: test_single_header <server_url> <header_name> [header_value]")
		fmt.Println("Example: test_single_header https://localhost:8443 Accept-Language 'en-US,en;q=0.9'")
		fmt.Println("Example: test_single_header https://localhost:8443 TE trailers")
		os.Exit(1)
	}

	serverURL := os.Args[1]
	headerName := os.Args[2]
	headerValue := ""
	if len(os.Args) > 3 {
		headerValue = os.Args[3]
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableKeepAlives:     true,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	// Create request
	req, err := http.NewRequest("GET", serverURL+"/health", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Add the test header
	if headerValue != "" {
		req.Header.Set(headerName, headerValue)
		fmt.Printf("Testing with header: %s: %s\n", headerName, headerValue)
	} else {
		fmt.Printf("Testing without header: %s\n", headerName)
	}

	// Add basic headers
	req.Header.Set("User-Agent", "test_single_header/1.0")

	// Print all headers
	fmt.Println("\nRequest headers:")
	for k, v := range req.Header {
		fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
	}

	// Make request with timing
	fmt.Println("\nMaking request...")
	start := time.Now()
	
	resp, err := client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			fmt.Printf("\n❌ TIMEOUT after %v - possible hang!\n", duration)
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n❌ Request failed after %v\n", duration)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("\n❌ Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Request successful!\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response headers:\n")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
	}
	
	if len(body) > 200 {
		fmt.Printf("\nResponse body (truncated):\n%s...\n", string(body[:200]))
	} else {
		fmt.Printf("\nResponse body:\n%s\n", string(body))
	}
}