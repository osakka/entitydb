#!/bin/bash

# EntityDB Swagger Documentation Generator
# This script automatically generates Swagger/OpenAPI documentation from Go annotations

set -e

echo "🔧 Generating EntityDB API Documentation..."

# Check if swag is installed
if ! command -v ~/go/bin/swag &> /dev/null; then
    echo "📦 Installing swag tool..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate swagger documentation
echo "📝 Generating swagger files..."
~/go/bin/swag init

# Copy generated files to htdocs for serving
echo "📂 Copying swagger files to htdocs..."
cp docs/swagger.json ../share/htdocs/swagger/
cp docs/swagger.yaml ../share/htdocs/swagger/

# Update version in main.go if needed
echo "🔄 Documentation generation complete!"
echo ""
echo "Generated files:"
echo "  - docs/docs.go (Go package)"
echo "  - docs/swagger.json (JSON spec)"
echo "  - docs/swagger.yaml (YAML spec)"
echo "  - ../share/htdocs/swagger/swagger.json (Served via web)"
echo ""
echo "📊 Access documentation at: https://localhost:8085/swagger/"
echo "🌐 Or view JSON spec at: https://localhost:8085/swagger/swagger.json"