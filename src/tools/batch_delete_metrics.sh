#!/bin/bash

# Batch delete old metrics more efficiently
# This script identifies and deletes old metric entities in batches

DB_PATH="/opt/entitydb/var"
BATCH_SIZE=50
TOTAL_DELETED=0

echo "=== EntityDB Batch Metric Cleanup ==="
echo "Finding old metric entities..."

# Get the entity IDs of old metrics
METRIC_IDS=$(./bin/entitydb-query -db "$DB_PATH" -tag "hub:metrics" -format id 2>/dev/null | grep -E '^[a-f0-9]{32}$')

if [ -z "$METRIC_IDS" ]; then
    echo "No old metrics found."
    exit 0
fi

TOTAL=$(echo "$METRIC_IDS" | wc -l)
echo "Found $TOTAL old metric entities to delete."

# Ask for confirmation
read -p "Are you sure you want to delete $TOTAL old metric entities? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Cleanup cancelled."
    exit 1
fi

echo "Starting batch deletion..."

# Process in batches
BATCH=()
COUNT=0

for ID in $METRIC_IDS; do
    BATCH+=("$ID")
    COUNT=$((COUNT + 1))
    
    # When batch is full, delete them
    if [ ${#BATCH[@]} -eq $BATCH_SIZE ]; then
        echo -n "Deleting batch of $BATCH_SIZE entities... "
        
        # Use curl to delete each entity
        for ENTITY_ID in "${BATCH[@]}"; do
            curl -s -X DELETE "https://localhost:8085/api/v1/entities/$ENTITY_ID" \
                -H "Authorization: Bearer $(cat /tmp/entitydb_token 2>/dev/null)" \
                --insecure > /dev/null 2>&1
        done
        
        TOTAL_DELETED=$((TOTAL_DELETED + ${#BATCH[@]}))
        echo "Done. ($TOTAL_DELETED/$TOTAL deleted)"
        
        # Clear batch
        BATCH=()
        
        # Small delay to avoid overwhelming the server
        sleep 0.1
    fi
done

# Delete remaining entities in last batch
if [ ${#BATCH[@]} -gt 0 ]; then
    echo -n "Deleting final batch of ${#BATCH[@]} entities... "
    
    for ENTITY_ID in "${BATCH[@]}"; do
        curl -s -X DELETE "https://localhost:8085/api/v1/entities/$ENTITY_ID" \
            -H "Authorization: Bearer $(cat /tmp/entitydb_token 2>/dev/null)" \
            --insecure > /dev/null 2>&1
    done
    
    TOTAL_DELETED=$((TOTAL_DELETED + ${#BATCH[@]}))
    echo "Done. ($TOTAL_DELETED/$TOTAL deleted)"
fi

echo ""
echo "✅ Successfully deleted $TOTAL_DELETED old metric entities."
echo ""

# Show remaining metrics
echo "Checking remaining metrics..."
NEW_COUNT=$(./bin/entitydb-query -db "$DB_PATH" -tag "type:metric" -format count 2>/dev/null || echo "0")
echo "Remaining temporal metrics: $NEW_COUNT"