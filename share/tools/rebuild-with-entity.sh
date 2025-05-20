#!/bin/bash

# Script to rebuild EntityDB with the entity-based architecture

# Function to log messages
log() {
    echo "$(date +'%Y-%m-%d %H:%M:%S') - $1"
}

# Step 1: Stop any running server
log "Stopping running server..."
/opt/entitydb/bin/entitydbd.sh stop

# Step 2: Back up the database
log "Backing up database..."
DB_PATH="/opt/entitydb/var/db/entitydb.db"
BACKUP_DIR="/opt/entitydb/var/db/backups"
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/entitydb_db_backup_$(date +'%Y%m%d_%H%M%S').db"
cp "$DB_PATH" "$BACKUP_FILE"
log "Database backed up to $BACKUP_FILE"

# Step 3: Apply entity migration
log "Applying entity schema migration..."
/opt/entitydb/share/tools/migrate_to_entity.sh
if [ $? -ne 0 ]; then
    log "ERROR: Migration failed. Restoring from backup..."
    cp "$BACKUP_FILE" "$DB_PATH"
    exit 1
fi

# Step 4: Rebuild server
log "Rebuilding server..."
cd /opt/entitydb/src && make
if [ $? -ne 0 ]; then
    log "ERROR: Server build failed."
    exit 1
fi

# Step 5: Start server
log "Starting server..."
/opt/entitydb/bin/entitydbd.sh start

# Step 6: Test entity creation
log "Testing entity creation..."
sleep 2  # Wait for server to start
/opt/entitydb/bin/entity-client.sh create --tags="type=test,status=active" --content-value="This is a test entity"

log "Rebuild completed successfully"
echo "You can use /opt/entitydb/bin/entity-client.sh to interact with the entity-based system"