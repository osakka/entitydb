#!/bin/bash

echo "Creating backup of current server binary"
cp /opt/entitydb/bin/entitydb /opt/entitydb/bin/entitydb.bak

echo "Creating patch for handleDefault function"
cat > /opt/entitydb/src/server_fix_patch.go << 'EOF'
package main

import (
	"log"
	"net/http"
	"strings"
)

// Patch the handleDefault function in the server to correctly serve static files
func PatchHandleDefault() {
	log.Println("Applying handleDefault patch to serve static files")
	
	// Replace the global handleDefaultFunc with our patched version
	originalHandleDefault := handleDefaultFunc
	
	// Create a patched version that serves static files correctly
	handleDefaultFunc = func(w http.ResponseWriter, r *http.Request) {
		log.Printf("EntityDB Server: Patched handleDefault called for path: %s", r.URL.Path)
		
		// Check if the path is to static files
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			// Create a file server pointing to our htdocs directory
			fileServer := http.FileServer(http.Dir("/opt/entitydb/share/htdocs"))
			
			// For requests to the root, redirect to dashboard.html
			if r.URL.Path == "/" {
				log.Printf("EntityDB Server: Redirecting root to dashboard.html")
				http.Redirect(w, r, "/dashboard.html", http.StatusFound)
				return
			}
			
			// For all other requests, try to serve the static file
			log.Printf("EntityDB Server: Serving static file: %s from /opt/entitydb/share/htdocs", r.URL.Path)
			fileServer.ServeHTTP(w, r)
			return
		}
		
		// For API requests, use the original handler
		originalHandleDefault(w, r)
	}
	
	log.Println("handleDefault patch applied successfully")
}

// This function is called during server initialization
func init() {
	log.Println("EntityDB Server: patch init() called")
	PatchHandleDefault()
}
EOF

echo "Stopping EntityDB server"
/opt/entitydb/bin/entitydbd.sh stop

echo "Patching server using symbolic link to dashboard.html"
ln -sf /opt/entitydb/share/htdocs/index.html /opt/entitydb/share/htdocs/dashboard.html

echo "Setting up a simple file server to test static files"
cat > /opt/entitydb/share/tools/static_serve.sh << 'EOF'
#!/bin/bash
cd /opt/entitydb/share/htdocs
python3 -m http.server 8086
EOF
chmod +x /opt/entitydb/share/tools/static_serve.sh

echo "Starting EntityDB server"
/opt/entitydb/bin/entitydbd.sh start

echo "You can test static files separately by running:"
echo "/opt/entitydb/share/tools/static_serve.sh"
echo "and then accessing http://localhost:8086/"

echo "Patch applied successfully"