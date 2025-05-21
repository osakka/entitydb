#!/bin/bash
# Script to apply temporal tag fix to the EntityDB server

# Helper function to show a formatted message
show_message() {
  echo "--------------------------------------"
  echo "$1"
  echo "--------------------------------------"
}

show_message "Starting temporal tag fix process"

# Stop the server if it's running
if pgrep -f "entitydb" > /dev/null; then
  show_message "Stopping EntityDB server..."
  if [ -f "/opt/entitydb/bin/entitydbd.sh" ]; then
    /opt/entitydb/bin/entitydbd.sh stop
  else
    pkill -f "entitydb"
  fi
  sleep 2
fi

# Navigate to source directory
cd /opt/entitydb/src || { 
  echo "Error: Source directory not found"; 
  exit 1; 
}

# Compile the improved_temporal_fix.go
show_message "Compiling the improved temporal fix..."
go build -o /opt/entitydb/bin/entitydb-with-fix

if [ $? -ne 0 ]; then
  show_message "Failed to compile temporal fix"
  exit 1
fi

# Restart the server
show_message "Starting EntityDB server with fix..."
if [ -f "/opt/entitydb/bin/entitydbd.sh" ]; then
  /opt/entitydb/bin/entitydbd.sh start
else
  nohup /opt/entitydb/bin/entitydb-with-fix > /dev/null 2>&1 &
fi

# Wait for server to start
show_message "Waiting for server to start..."
sleep 5

# Test the fix
show_message "Testing the temporal tag fix..."
cd /opt/entitydb || exit 1
./improved_temporal_fix.sh

# Show completion message
show_message "Temporal tag fix process completed"