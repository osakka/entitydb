#!/bin/bash

# Script to apply schema updates to the EntityDB database
# Author: Claude
# Date: 2025-05-07

# Database location
DB_PATH="/opt/entitydb/var/db/entitydb.db"
SCHEMA_DIR="/opt/entitydb/src/models/sqlite"
CONSOLIDATED_SCHEMA="${SCHEMA_DIR}/schema_update_consolidated.sql"
BACKUP_DIR="/opt/entitydb/var/db/backups"
LOG_FILE="/opt/entitydb/var/log/schema_fix.log"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Log function
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Ensure DB exists
if [ ! -f "$DB_PATH" ]; then
  log "Error: Database not found at $DB_PATH"
  log "Starting server to create database..."
  /opt/entitydb/bin/entitydbd.sh start
  sleep 3
  /opt/entitydb/bin/entitydbd.sh stop
  
  if [ ! -f "$DB_PATH" ]; then
    log "Error: Failed to create database"
    exit 1
  fi
fi

# Create a backup before making changes
BACKUP_FILE="$BACKUP_DIR/entitydb_db_backup_$(date '+%Y%m%d_%H%M%S').db"
cp "$DB_PATH" "$BACKUP_FILE"
log "Database backup created at $BACKUP_FILE"

# Fix column renaming separately if needed
log "Checking for column rename requirements..."
HAS_OLD_COLUMN=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM pragma_table_info('agents') WHERE name='last_active_at'")
HAS_NEW_COLUMN=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM pragma_table_info('agents') WHERE name='last_active'")

if [ "$HAS_OLD_COLUMN" -eq "1" ] && [ "$HAS_NEW_COLUMN" -eq "0" ]; then
  log "Renaming column last_active_at to last_active..."
  sqlite3 "$DB_PATH" <<EOF
BEGIN TRANSACTION;
CREATE TABLE agents_new (
    id TEXT PRIMARY KEY,
    handle TEXT UNIQUE NOT NULL,
    name TEXT,
    display_name TEXT NOT NULL,
    type TEXT NOT NULL,
    specialization TEXT,
    personality_profile TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP,
    worker_pool_id TEXT,
    expertise TEXT DEFAULT '',
    capability_score INTEGER DEFAULT 0
);
INSERT INTO agents_new SELECT id, handle, name, display_name, type, specialization, personality_profile, status, created_at, last_active_at, '', '', 0 FROM agents;
DROP TABLE agents;
ALTER TABLE agents_new RENAME TO agents;
COMMIT;
EOF
  if [ $? -ne 0 ]; then
    log "Error: Failed to rename column"
    exit 1
  fi
  log "Column renamed successfully"
else
  log "No column rename needed"
fi

# Check if server is running and stop it if needed
if pgrep -f "/opt/entitydb/bin/entitydb" > /dev/null; then
  log "Stopping EntityDB server before applying schema updates..."
  /opt/entitydb/bin/entitydbd.sh stop
  sleep 2
fi

# Apply consolidated schema updates
log "Applying consolidated schema updates..."
sqlite3 "$DB_PATH" < "$CONSOLIDATED_SCHEMA"

if [ $? -ne 0 ]; then
  log "Warning: Some schema updates may have been skipped (likely because columns already exist)"
fi

log "Schema updates completed"

# Verify table structure
log "Verifying updated schema..."

# Check auth_tokens table
sqlite3 "$DB_PATH" "PRAGMA table_info(auth_tokens)" | grep -q "token_type"
if [ $? -ne 0 ]; then
  log "Error: auth_tokens table missing token_type column"
  exit 1
fi

# Check agents table
sqlite3 "$DB_PATH" "PRAGMA table_info(agents)" | grep -q "worker_pool_id"
if [ $? -ne 0 ]; then
  log "Error: agents table missing worker_pool_id column"
  exit 1
fi

sqlite3 "$DB_PATH" "PRAGMA table_info(agents)" | grep -q "expertise"
if [ $? -ne 0 ]; then
  log "Error: agents table missing expertise column"
  exit 1
fi

sqlite3 "$DB_PATH" "PRAGMA table_info(agents)" | grep -q "capability_score"
if [ $? -ne 0 ]; then
  log "Error: agents table missing capability_score column"
  exit 1
fi

# Check agent_capabilities table existence
sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='agent_capabilities'" | grep -q "agent_capabilities"
if [ $? -ne 0 ]; then
  log "Error: agent_capabilities table not created"
  exit 1
fi

# Check agent_performance table existence
sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='agent_performance'" | grep -q "agent_performance"
if [ $? -ne 0 ]; then
  log "Error: agent_performance table not created"
  exit 1
fi

# Create admin users if they don't exist
log "Ensuring admin users exist..."

# Check for default admin user (password: password)
EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE username='admin'")
if [ "$EXISTS" -eq "0" ]; then
  log "Creating default admin user..."
  # bcrypt hash for 'password'
  HASH='$2a$10$fDkwkDSBIcYYSW0Kb3XtBuWGK6PCN1zdTQn47IrktED.y.9QYIqGq'
  sqlite3 "$DB_PATH" "INSERT INTO users (id, username, password_hash, display_name, full_name, roles, active, status) VALUES ('user_admin', 'admin', '$HASH', 'Administrator', 'Administrator', 'admin', 1, 'active')"
  
  # Grant all permissions to admin
  sqlite3 "$DB_PATH" "INSERT OR IGNORE INTO user_permissions SELECT 'user_admin', id FROM permissions"
fi

# Check for easier admin2 user (password: admin)
EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE username='admin2'")
if [ "$EXISTS" -eq "0" ]; then
  log "Creating admin2 user with simplified credentials..."
  # Generate user ID
  USER_ID="usr_$(date +%s)_$(shuf -i 1000-9999 -n 1)"
  # bcrypt hash for 'admin'
  HASH='$2a$10$hVWGiIjcrHrxItYdSMIAD.mUkW4PK64QTY7NVzUyGr1.iyLOW5/nG'
  sqlite3 "$DB_PATH" "INSERT INTO users (id, username, password_hash, display_name, full_name, roles, active, status) VALUES ('$USER_ID', 'admin2', '$HASH', 'Admin User', 'Admin User', 'admin', 1, 'active')"
  
  # Grant all permissions to admin2
  sqlite3 "$DB_PATH" "INSERT OR IGNORE INTO user_permissions SELECT '$USER_ID', id FROM permissions"
  log "Created admin2 user with password 'admin'"
fi

# Also create a test user for development
log "Creating test user with admin privileges..."
TEST_EXISTS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users WHERE username='testuser'")
if [ "$TEST_EXISTS" -eq "0" ]; then
  # bcrypt hash for 'testpass'
  TEST_HASH='$2a$10$NQy2gUfI7HaD2c4ZwexC2ub1onLR1aMK.23D7IroKRnQh2nBUSCT2'
  sqlite3 "$DB_PATH" "INSERT INTO users (id, username, password_hash, display_name, full_name, roles, active, status) VALUES ('user_test', 'testuser', '$TEST_HASH', 'Test User', 'Test User', 'admin', 1, 'active')"
  
  # Grant all permissions to test user
  sqlite3 "$DB_PATH" "INSERT OR IGNORE INTO user_permissions SELECT 'user_test', id FROM permissions"
fi

# Restart the server
log "Restarting EntityDB server..."
/opt/entitydb/bin/entitydbd.sh restart

log "Database schema verified and fixed"
log "Server has been restarted with the updated schema"

exit 0