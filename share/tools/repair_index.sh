#!/bin/bash
# Repair EntityDB index

echo "=== EntityDB Index Repair ==="
echo "This will attempt to fix corrupted index entries"
echo

# Path configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENTITYDB_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Stop the server first
echo "Stopping EntityDB server..."
"$ENTITYDB_DIR/bin/entitydbd.sh" stop

sleep 2

# Create a repair tool
cat > /tmp/repair_index.go << 'EOF'
package main

import (
	"entitydb/storage/binary"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: repair_index <data_dir>")
		os.Exit(1)
	}

	dataDir := os.Args[1]
	
	// Open the repository
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	
	// Repair the index
	fmt.Println("Repairing index...")
	if err := repo.RepairIndex(); err != nil {
		log.Fatalf("Failed to repair index: %v", err)
	}
	
	fmt.Println("Index repair completed successfully")
}
EOF

# Build and run the repair tool
cd "$ENTITYDB_DIR/src"
go build -o /tmp/repair_index /tmp/repair_index.go

echo "Running index repair..."
/tmp/repair_index "$ENTITYDB_DIR/var"

# Clean up
rm -f /tmp/repair_index /tmp/repair_index.go

echo
echo "Index repair complete. You can now restart the server with:"
echo "$ENTITYDB_DIR/bin/entitydbd.sh start"