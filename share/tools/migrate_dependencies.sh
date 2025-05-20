#!/bin/bash
set -e

# Script to build and run the dependency migration tool

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
TOOL_SRC="$ROOT_DIR/share/tools/migrate_issue_dependencies.go"
TOOL_BIN="$ROOT_DIR/var/migrate_issue_dependencies"
DB_PATH="$ROOT_DIR/var/db/entitydb.db"
LOG_DIR="$ROOT_DIR/var/log"
LOG_FILE="$LOG_DIR/dependency_migration.log"

# Make sure log directory exists
mkdir -p "$LOG_DIR"

# Ensure the database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Database file not found at $DB_PATH"
    exit 1
fi

# Create a backup of the database
BACKUP_DIR="$ROOT_DIR/var/db/backups"
mkdir -p "$BACKUP_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/entitydb_db_backup_${TIMESTAMP}.db"
echo "Creating database backup at $BACKUP_FILE"
cp "$DB_PATH" "$BACKUP_FILE"

# Build the migration tool
echo "Building migration tool..."
go build -o "$TOOL_BIN" "$TOOL_SRC"

# Display usage
function show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  --dry-run          Perform a dry run without making changes"
    echo "  --batch-size N     Process N dependencies at a time (default: 100)"
    echo "  --force            Force migration even if already migrated"
    echo "  --non-interactive  Run without prompting for confirmation"
    echo "  --help             Show this help message"
    echo
}

# Parse arguments
DRY_RUN=""
BATCH_SIZE="100"
FORCE=""
INTERACTIVE="--interactive"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --dry-run) DRY_RUN="--dry-run"; shift ;;
        --batch-size) BATCH_SIZE="$2"; shift 2 ;;
        --force) FORCE="--force"; shift ;;
        --non-interactive) INTERACTIVE=""; shift ;;
        --help) show_usage; exit 0 ;;
        *) echo "Unknown parameter: $1"; show_usage; exit 1 ;;
    esac
done

# Run the migration tool
echo "Running dependency migration tool..."
"$TOOL_BIN" --db="$DB_PATH" --log="$LOG_FILE" --batch-size="$BATCH_SIZE" $DRY_RUN $FORCE $INTERACTIVE

# Exit with the migration tool's exit code
EXIT_CODE=$?
if [ $EXIT_CODE -eq 0 ]; then
    echo "Migration completed successfully"
else
    echo "Migration completed with errors, check $LOG_FILE for details"
fi

exit $EXIT_CODE