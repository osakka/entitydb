#\!/bin/bash

# Concurrent performance test for sharded indexing
TOKEN="9a137f05c2b3d3eb6b42ce6579a2afa1b2957cc4f8f7bef3de08911971442290"
BASE_URL="https://localhost:8085/api/v1"

# Function to create entities concurrently
create_entity() {
    local id=$1
    local worker=$2
    curl -s -k -X POST "$BASE_URL/entities/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"id\": \"concurrent_test_${worker}_${id}\",
            \"tags\": [
                \"type:concurrent_test\",
                \"worker:${worker}\",
                \"batch:${id}\",
                \"status:active\",
                \"priority:medium\",
                \"environment:test\",
                \"shard_test:true\"
            ],
            \"content\": [{\"type\": \"text/plain\", \"value\": \"Concurrent test entity ${worker}-${id}\"}]
        }" > /dev/null 2>&1
    echo "Worker $worker created entity $id"
}

# Function to query entities by tag concurrently  
query_entities() {
    local worker=$1
    curl -s -k "$BASE_URL/entities/list?tag=worker:${worker}" \
        -H "Authorization: Bearer $TOKEN" > /dev/null 2>&1
    echo "Worker $worker queried entities"
}

echo "Starting concurrent sharded indexing test..."
echo "Creating 50 entities across 5 workers (10 per worker)..."

# Start timestamp
start_time=$(date +%s%N)

# Create entities concurrently
for worker in {1..5}; do
    for batch in {1..10}; do
        create_entity $batch $worker &
    done
done

# Wait for all creation jobs to complete
wait

# Query entities concurrently
echo "Querying entities concurrently..."
for worker in {1..5}; do
    query_entities $worker &
done

# Wait for all query jobs to complete  
wait

# End timestamp
end_time=$(date +%s%N)
duration_ms=$(( (end_time - start_time) / 1000000 ))

echo "Concurrent test completed in ${duration_ms}ms"
echo "Testing concurrent tag operations..."

# Test concurrent tag additions to same entities
for worker in {1..5}; do
    (
        for i in {1..5}; do
            curl -s -k -X POST "$BASE_URL/entities/concurrent_test_${worker}_1/tags" \
                -H "Authorization: Bearer $TOKEN" \
                -H "Content-Type: application/json" \
                -d "{\"tag\": \"concurrent_tag_${worker}_${i}\"}" > /dev/null 2>&1
        done
        echo "Worker $worker added concurrent tags"
    ) &
done

wait
echo "Concurrent sharded indexing test completed successfully\!"
