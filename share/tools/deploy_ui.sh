#!/bin/bash

# Stop the EntityDB server
echo "Stopping EntityDB server..."
/opt/entitydb/bin/entitydbd.sh stop

# Create a backup of the current htdocs directory
echo "Creating backup of current htdocs directory..."
BACKUP_DIR="/opt/entitydb/share/htdocs.backup.$(date +%Y%m%d%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -r /opt/entitydb/share/htdocs/* "$BACKUP_DIR/"

# Ensure dashboard.html is an exact copy of index.html (not a symlink)
echo "Creating dashboard.html as copy of index.html..."
rm -f /opt/entitydb/share/htdocs/dashboard.html
cp /opt/entitydb/share/htdocs/index.html /opt/entitydb/share/htdocs/dashboard.html

# Create a simple Python static file server to serve our UI
echo "Creating Python static file server..."
cat > /opt/entitydb/share/tools/static_server.py << 'EOF'
#!/usr/bin/env python3

import http.server
import socketserver
import os
import sys
from urllib.parse import urlparse, parse_qs

PORT = 8086
DIRECTORY = "/opt/entitydb/share/htdocs"

class EntityDBHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DIRECTORY, **kwargs)
        
    def do_GET(self):
        # Redirect root to dashboard.html
        if self.path == "/":
            self.send_response(302)
            self.send_header("Location", "/dashboard.html")
            self.end_headers()
            return
            
        # Forward API requests to the EntityDB server
        if self.path.startswith("/api/"):
            self.send_response(404)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(b"API requests are not handled by the static server")
            return
            
        return http.server.SimpleHTTPRequestHandler.do_GET(self)
    
    def log_message(self, format, *args):
        print(f"[Static Server] {self.address_string()} - {format % args}")

def run_server():
    os.chdir(DIRECTORY)
    handler = EntityDBHandler
    
    with socketserver.TCPServer(("", PORT), handler) as httpd:
        print(f"Serving EntityDB UI at http://localhost:{PORT}/")
        httpd.serve_forever()

if __name__ == "__main__":
    run_server()
EOF

chmod +x /opt/entitydb/share/tools/static_server.py

# Create a script to start both servers
echo "Creating startup script..."
cat > /opt/entitydb/bin/start_entitydb_ui.sh << 'EOF'
#!/bin/bash

# Kill any existing static server
pkill -f "python3 .*static_server.py" 2>/dev/null

# Start the EntityDB server
/opt/entitydb/bin/entitydbd.sh start

# Start the static file server
nohup /opt/entitydb/share/tools/static_server.py > /tmp/entitydb_static_server.log 2>&1 &

echo "EntityDB server started on port 8085 (API only)"
echo "EntityDB UI server started on port 8086"
echo "Access the UI at: http://localhost:8086/"
EOF

chmod +x /opt/entitydb/bin/start_entitydb_ui.sh

# Create a script to stop both servers
echo "Creating shutdown script..."
cat > /opt/entitydb/bin/stop_entitydb_ui.sh << 'EOF'
#!/bin/bash

# Stop the EntityDB server
/opt/entitydb/bin/entitydbd.sh stop

# Kill the static server
pkill -f "python3 .*static_server.py" 2>/dev/null

echo "EntityDB servers stopped"
EOF

chmod +x /opt/entitydb/bin/stop_entitydb_ui.sh

# Start both servers
echo "Starting servers..."
/opt/entitydb/bin/start_entitydb_ui.sh

echo "Deployment complete!"
echo "Your EntityDB UI is now available at: http://localhost:8086/"
echo "The EntityDB API is available at: http://localhost:8085/api/v1/"
echo ""
echo "To start both servers: /opt/entitydb/bin/start_entitydb_ui.sh"
echo "To stop both servers: /opt/entitydb/bin/stop_entitydb_ui.sh"