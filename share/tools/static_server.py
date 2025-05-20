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
