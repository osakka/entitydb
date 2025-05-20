#!/bin/bash

# Install Nginx if not already installed
echo "Checking if Nginx is installed..."
if ! command -v nginx &> /dev/null; then
    echo "Installing Nginx..."
    apt-get update
    apt-get install -y nginx
else
    echo "Nginx is already installed."
fi

# Create Nginx configuration for EntityDB
echo "Creating Nginx configuration for EntityDB..."
cat > /etc/nginx/sites-available/entitydb << 'EOF'
server {
    listen 8087;
    
    location / {
        root /opt/entitydb/share/htdocs;
        index index.html dashboard.html;
        try_files $uri $uri/ /index.html;
    }
    
    location /api/ {
        proxy_pass http://localhost:8085;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Enable the site
echo "Enabling the site..."
ln -sf /etc/nginx/sites-available/entitydb /etc/nginx/sites-enabled/

# Test Nginx configuration
echo "Testing Nginx configuration..."
nginx -t

# Restart Nginx
echo "Restarting Nginx..."
systemctl restart nginx

echo "Nginx has been configured. Access the UI at http://localhost:8087/"