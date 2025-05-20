#!/bin/bash
# Initialize the database with compatibility tables for testing

# Color codes for better readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directory setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
DB_DIR="$ROOT_DIR/var/db"
SCHEMA_DIR="$ROOT_DIR/src/models/sqlite"

# Database file
DB_PATH="$DB_DIR/entitydb.db"

# Make sure the database directory exists
mkdir -p "$DB_DIR"

# Create a backup of the existing database if it exists
if [ -f "$DB_PATH" ]; then
    BACKUP_NAME="entitydb_db_backup_$(date +%Y%m%d_%H%M%S).db"
    BACKUP_DIR="$DB_DIR/backups"
    mkdir -p "$BACKUP_DIR"
    cp "$DB_PATH" "$BACKUP_DIR/$BACKUP_NAME"
    echo -e "${BLUE}Created backup of existing database at:${NC} $BACKUP_DIR/$BACKUP_NAME"
    
    # Remove existing database
    rm "$DB_PATH"
    echo -e "${YELLOW}Deleted existing database at${NC} $DB_PATH"
fi

# Check if SQLite is installed
if ! command -v sqlite3 &> /dev/null; then
    echo -e "${RED}Error: sqlite3 is not installed${NC}"
    exit 1
fi

# Enable foreign keys
echo -e "${BLUE}Foreign keys enabled:${NC} true"

# Apply the minimal schema
MINIMAL_SCHEMA="$SCHEMA_DIR/minimal_schema.sql"
echo -e "${BLUE}Using schema file:${NC} $MINIMAL_SCHEMA"
sqlite3 "$DB_PATH" < "$MINIMAL_SCHEMA"

# Apply the compatibility schema
COMPAT_SCHEMA="$SCHEMA_DIR/schema_update_test_compat.sql"
echo -e "${BLUE}Applying compatibility schema:${NC} $COMPAT_SCHEMA"
sqlite3 "$DB_PATH" < "$COMPAT_SCHEMA"

# Verify database creation
if [ -f "$DB_PATH" ]; then
    # Get the database size in bytes and convert to human-readable format
    DB_SIZE=$(stat -c%s "$DB_PATH" 2>/dev/null || stat -f%z "$DB_PATH")
    DB_SIZE_HUMAN=$(numfmt --to=iec-i --suffix=B --format="%.2f" $DB_SIZE 2>/dev/null || echo "$DB_SIZE bytes")
    
    echo -e "${GREEN}Database initialized successfully!${NC}"
    echo -e "${BLUE}Database path:${NC} $DB_PATH"
    echo -e "${BLUE}Database size:${NC} $DB_SIZE_HUMAN"
    
    # Show table count
    TABLE_COUNT=$(echo ".tables" | sqlite3 "$DB_PATH" | wc -w)
    echo -e "${BLUE}Tables created:${NC} $TABLE_COUNT"
    
    exit 0
else
    echo -e "${RED}Failed to create database!${NC}"
    exit 1
fi