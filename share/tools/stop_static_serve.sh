#!/bin/bash

# Default port
PORT=8086

# Check if port is provided as argument
if [ ! -z "$1" ]; then
  PORT=$1
fi

# Kill any existing static server
pkill -f "python3 -m http.server $PORT" 2>/dev/null

if [ $? -eq 0 ]; then
  echo "Static server on port $PORT stopped successfully."
else
  echo "No static server found running on port $PORT."
fi