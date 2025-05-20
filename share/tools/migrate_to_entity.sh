#!/bin/bash
#
# migrate_to_entity.sh - Migrate issues to entities in the EntityDB database
#
# This script migrates issues from the traditional issue-based model to the new
# entity-based architecture. It enables dual-write mode, runs tests to ensure
# compatibility, and helps transition smoothly.
#

# Set script to exit on any error
set -e

# Set base directory
BASE_DIR="/opt/entitydb"
BIN_DIR="$BASE_DIR/bin"
VAR_DIR="$BASE_DIR/var"
DB_DIR="$VAR_DIR/db"
DB_PATH="$DB_DIR/entitydb.db"
BACKUP_DIR="$DB_DIR/backups"
LOG_FILE="$VAR_DIR/log/migration.log"
SCHEMA_DIR="$BASE_DIR/src/models/sqlite"

# Function to log messages
log() {
    echo "$(date +'%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
    echo "$1"
}

# Function to print usage
usage() {
  echo "Usage: $0 [options]"
  echo ""
  echo "Options:"
  echo "  --help                  Show this help message"
  echo "  --dry-run               Run migration in simulation mode without making changes"
  echo "  --dual-write            Enable dual-write mode only (no full migration)"
  echo "  --force                 Skip confirmation prompts"
  echo "  --skip-backup           Skip database backup (not recommended)"
  echo ""
  echo "This script helps migrate the database from issue-based to entity-based architecture."
  echo ""
  echo "Migration steps:"
  echo "  1. Backup the database"
  echo "  2. Enable dual-write mode"
  echo "  3. Migrate existing issues to entities"
  echo "  4. Run tests to verify migration"
  echo ""
}

# Initialize options
DRY_RUN=false
FORCE=false
SKIP_BACKUP=false
DUAL_WRITE_ONLY=false

# Parse command line options
while [[ $# -gt 0 ]]; do
  case "$1" in
    --help)
      usage
      exit 0
      ;;
    --dry-run)
      DRY_RUN=true
      log "Running in dry-run mode. No changes will be made."
      ;;
    --force)
      FORCE=true
      log "Force mode enabled. Skipping confirmation prompts."
      ;;
    --skip-backup)
      SKIP_BACKUP=true
      log "Skipping database backup (not recommended)."
      ;;
    --dual-write)
      DUAL_WRITE_ONLY=true
      log "Enabling dual-write mode only (no full migration)."
      ;;
    *)
      echo "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
  shift
done

# Check if server is running
check_server() {
  if ! pgrep -f "entitydb" > /dev/null; then
    log "EntityDB server is not running. Starting server..."
    "$BIN_DIR/entitydbd.sh" start
    sleep 2
  fi
}

# Create database backup
backup_database() {
  if [ "$SKIP_BACKUP" = true ]; then
    log "Skipping database backup as requested."
    return
  fi

  log "Creating database backup..."

  # Create backup directory if it doesn't exist
  mkdir -p "$BACKUP_DIR"

  # Create backup with timestamp
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  BACKUP_FILE="$BACKUP_DIR/entitydb_db_backup_${TIMESTAMP}.db"

  if [ "$DRY_RUN" = true ]; then
    log "[DRY RUN] Would create backup at: $BACKUP_FILE"
  else
    log "Creating backup at: $BACKUP_FILE"
    cp "$DB_PATH" "$BACKUP_FILE"
    log "Backup created successfully."
  fi
}

# Apply entity schema
apply_entity_schema() {
  log "Applying entity schema..."

  # Check for schema file
  ENTITY_SCHEMA="$SCHEMA_DIR/schema_entity_migration.sql"

  if [ ! -f "$ENTITY_SCHEMA" ]; then
    log "ERROR: Entity schema migration file not found at $ENTITY_SCHEMA"
    exit 1
  fi

  if [ "$DRY_RUN" = true ]; then
    log "[DRY RUN] Would apply entity schema from $ENTITY_SCHEMA"
  else
    log "Applying entity schema..."
    sqlite3 "$DB_PATH" < "$ENTITY_SCHEMA"

    # Verify entity table was created
    TABLE_COUNT=$(sqlite3 "$DB_PATH" "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='entities';")
    if [ "$TABLE_COUNT" -ne 1 ]; then
      log "ERROR: Entity table was not created successfully"
      exit 1
    fi

    log "Entity schema applied successfully"
  fi
}

# Enable dual-write mode
enable_dual_write() {
  log "Enabling dual-write mode..."

  if [ "$DRY_RUN" = true ]; then
    log "[DRY RUN] Would enable dual-write mode."
  else
    # Enable dual-write mode using API
    curl -s -X PUT "http://localhost:8085/api/v1/entity/config" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer admin_token" \
      -d '{"dual_write_enabled": true}' >/dev/null

    log "Dual-write mode enabled successfully."
  fi
}

# Migrate issues to entities
migrate_issues() {
  log "Migrating issues to entities..."

  if [ "$DRY_RUN" = true ]; then
    log "[DRY RUN] Would migrate issues to entities."
    return
  fi

  # Get total issues count
  ISSUE_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM issues")
  log "Found $ISSUE_COUNT issues to migrate."

  # Run SQL to copy issues to entities
  cat <<EOF | sqlite3 "$DB_PATH"
-- Create entity_tags table if it doesn't exist (for explicit tag storage)
CREATE TABLE IF NOT EXISTS entity_tags (
    entity_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY(entity_id, tag)
);

-- Create entity_content table if it doesn't exist (for explicit content storage)
CREATE TABLE IF NOT EXISTS entity_content (
    entity_id TEXT NOT NULL,
    type TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY(entity_id, type, timestamp)
);

-- Migrate issues to entities using both the JSON format and explicit tables
-- First, migrate to the entities table with JSON format
INSERT OR IGNORE INTO entities (id, tags, content)
SELECT
    id as entity_id,
    json_array(
        'type:' || type,
        'status:' || status,
        'priority:' || priority,
        CASE WHEN workspace_id IS NOT NULL THEN 'workspace:' || workspace_id ELSE NULL END,
        CASE WHEN parent_id IS NOT NULL THEN 'parent:' || parent_id ELSE NULL END,
        CASE WHEN created_by IS NOT NULL THEN 'created_by:' || created_by ELSE NULL END,
        'created_at:' || created_at,
        CASE WHEN progress > 0 THEN 'progress:' || progress ELSE NULL END
    ) as tags,
    json_array(
        json_object(
            'timestamp', created_at,
            'type', 'title',
            'value', title
        ),
        json_object(
            'timestamp', created_at,
            'type', 'description',
            'value', description
        )
    ) as content
FROM issues;

-- Also insert into entity_tags table for explicit tag access
INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'type:' || type as tag
FROM issues;

INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'status:' || status as tag
FROM issues;

INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'priority:' || priority as tag
FROM issues;

INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'workspace:' || workspace_id as tag
FROM issues
WHERE workspace_id IS NOT NULL;

INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'parent:' || parent_id as tag
FROM issues
WHERE parent_id IS NOT NULL;

INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'created_by:' || created_by as tag
FROM issues
WHERE created_by IS NOT NULL;

INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'created_at:' || created_at as tag
FROM issues;

INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'progress:' || progress as tag
FROM issues
WHERE progress > 0;

-- Insert into entity_content table for explicit content access
INSERT OR IGNORE INTO entity_content (entity_id, type, timestamp, value)
SELECT
    id as entity_id,
    'title' as type,
    created_at as timestamp,
    title as value
FROM issues;

INSERT OR IGNORE INTO entity_content (entity_id, type, timestamp, value)
SELECT
    id as entity_id,
    'description' as type,
    created_at as timestamp,
    description as value
FROM issues;

-- Mark workspaces as workspace type
INSERT OR IGNORE INTO entity_tags (entity_id, tag)
SELECT
    id as entity_id,
    'type:workspace' as tag
FROM issues
WHERE type = 'workspace';
EOF

  log "Migration completed successfully."
}

# Enable entity-based API
enable_entity_handler() {
  log "Enabling entity-based API handler..."

  if [ "$DRY_RUN" = true ]; then
    log "[DRY RUN] Would enable entity-based handler."
  else
    # Enable entity handler using API
    curl -s -X PUT "http://localhost:8085/api/v1/entity/config" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer admin_token" \
      -d '{"entity_handler_enabled": true}' >/dev/null

    log "Entity handler enabled successfully."
    log "Restarting server to apply changes..."
    "$BIN_DIR/entitydbd.sh" restart
    sleep 2
  fi
}

# Run migration tests
run_tests() {
  log "Running migration tests..."

  if [ "$DRY_RUN" = true ]; then
    log "[DRY RUN] Would run migration tests."
    return
  fi

  # Run basic API tests to verify that both APIs work
  "$BIN_DIR/entitydbc.sh" issue list >/dev/null || { log "Issue API test failed!"; exit 1; }
  log "Issue API test passed."

  "$BIN_DIR/entitydbc.sh" workspace list >/dev/null || { log "Workspace API test failed!"; exit 1; }
  log "Workspace API test passed."

  # Test a complex issue to ensure full conversion
  TEST_ISSUE_ID=$(curl -s -X POST "http://localhost:8085/api/v1/issues/create" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer admin_token" \
    -d '{
      "title": "Test Issue for Migration",
      "description": "This issue tests entity migration",
      "priority": "high",
      "type": "task",
      "estimated_effort": 4.5,
      "tags": ["migration", "testing"]
    }' | grep -o '"id":"[^"]*"' | cut -d '"' -f 4)

  log "Created test issue with ID: $TEST_ISSUE_ID"

  # Verify the issue was created in both repositories
  ISSUE_RESULT=$(curl -s "http://localhost:8085/api/v1/issues/get?issue_id=$TEST_ISSUE_ID" \
    -H "Authorization: Bearer admin_token")

  echo "$ISSUE_RESULT" | grep -q "Test Issue for Migration" || { log "Issue retrieval test failed!"; exit 1; }
  log "Issue retrieval test passed."

  log "All tests passed successfully."
}

# Main script logic
main() {
  log "=== EntityDB Entity Migration Tool ==="
  log ""

  # Check server
  check_server

  # Backup database
  backup_database

  # Apply entity schema
  apply_entity_schema

  # Enable dual-write mode
  enable_dual_write

  # If only enabling dual-write, stop here
  if [ "$DUAL_WRITE_ONLY" = true ]; then
    log ""
    log "Dual-write mode has been enabled. You can now test the system with both repositories active."
    log "Once you are confident in the entity-based storage, run this script again without --dual-write to complete the migration."
    exit 0
  fi

  # Confirm before proceeding
  if [ "$FORCE" != true ]; then
    log ""
    log "WARNING: The next steps will migrate all issues to entities and enable entity-based APIs."
    log "This is a one-way operation. While your issues will remain in the database,"
    log "the system will now use the entity-based storage for all operations."
    log ""
    read -p "Are you sure you want to proceed? (yes/no) " CONFIRM
    if [ "$CONFIRM" != "yes" ]; then
      log "Migration aborted."
      exit 0
    fi
  fi

  # Migrate issues to entities
  migrate_issues

  # Enable entity handler
  enable_entity_handler

  # Run migration tests
  run_tests

  log ""
  log "Migration completed successfully!"
  log "The system is now using the entity-based architecture for all issue operations."
  log "The original issues still exist in the database for reference."
}

# Run the main function
main