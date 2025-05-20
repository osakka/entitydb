#!/bin/bash

# Create a backup of the current htdocs directory for safety
echo "Creating backup of current htdocs directory"
BACKUP_DIR="/opt/entitydb/share/htdocs.backup.$(date +%Y%m%d%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -r /opt/entitydb/share/htdocs/* "$BACKUP_DIR/"

# Ensure dashboard.html is a copy (not symlink) of index.html
echo "Creating dashboard.html from index.html"
rm -f /opt/entitydb/share/htdocs/dashboard.html
cp /opt/entitydb/share/htdocs/index.html /opt/entitydb/share/htdocs/dashboard.html

# Create a simple index page that redirects to dashboard.html
echo "Creating redirecting index.html"
cat > /opt/entitydb/share/htdocs/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0;url=dashboard.html">
    <title>Redirecting to Dashboard</title>
</head>
<body>
    <p>Redirecting to <a href="dashboard.html">Dashboard</a>...</p>
</body>
</html>
EOF

# Restart the EntityDB server
echo "Restarting EntityDB server"
/opt/entitydb/bin/entitydbd.sh restart

echo "Fix applied. Access the UI at http://localhost:8085/"