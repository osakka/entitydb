#!/bin/bash
# Demo: Temporal Metrics Collection in EntityDB

BASE_URL="https://localhost:8085/api/v1"
CURL_OPTS="-k -s"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║        EntityDB Temporal Metrics Collection Demo             ║"
echo "║                                                              ║"
echo "║  Concept: ONE entity per metric, infinite temporal values   ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo

# Login
TOKEN=$(curl $CURL_OPTS -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to login"
    exit 1
fi

echo "🔐 Authenticated successfully"
echo

# First, let's create a fresh metric entity
echo "📊 Creating CPU Usage Metric Entity"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Send first metric value
CPU_VALUE=45.5
echo "→ Sending CPU usage: ${CPU_VALUE}%"
RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/metrics/collect \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"metric_name\": \"demo_cpu_usage\",
        \"value\": $CPU_VALUE,
        \"unit\": \"percent\",
        \"instance\": \"demo_server\",
        \"labels\": {
            \"region\": \"us-east\",
            \"env\": \"demo\"
        }
    }")
echo "$RESPONSE" | jq .
METRIC_ID=$(echo "$RESPONSE" | jq -r '.metric_id')
echo

# Wait a bit and send more values
echo "⏱️  Collecting more values over time..."
sleep 2

for i in {1..4}; do
    CPU_VALUE=$(echo "40 + $RANDOM % 20" | bc)
    echo "→ CPU usage update #$i: ${CPU_VALUE}%"
    curl $CURL_OPTS -X POST $BASE_URL/metrics/collect \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"metric_name\": \"demo_cpu_usage\",
            \"value\": $CPU_VALUE,
            \"unit\": \"percent\",
            \"instance\": \"demo_server\"
        }" | jq -r '.message'
    sleep 1
done
echo

# Now let's look at the entity
echo "🔍 Examining the Metric Entity Structure"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Entity ID: $METRIC_ID"
echo

ENTITY=$(curl $CURL_OPTS -X GET "$BASE_URL/entities/get?id=$METRIC_ID&include_timestamps=true" \
    -H "Authorization: Bearer $TOKEN")

echo "📌 Entity Tags (showing temporal nature):"
echo "$ENTITY" | jq -r '.entity.tags[]' | grep -E "(type:|metric:name:|metric:value:)" | head -10
echo

echo "📈 Metric Value History:"
echo "$ENTITY" | jq -r '.entity.tags[]' | grep "metric:value:" | while read -r tag; do
    if [[ $tag =~ ^([^|]+)\|metric:value:([0-9.]+):(.+)$ ]]; then
        timestamp="${BASH_REMATCH[1]}"
        value="${BASH_REMATCH[2]}"
        unit="${BASH_REMATCH[3]}"
        echo "  • $timestamp → ${value}${unit}"
    fi
done
echo

# Query the history through the API
echo "📊 Querying Metric History via API"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
HISTORY=$(curl $CURL_OPTS -X GET "$BASE_URL/metrics/history?metric=demo_cpu_usage&instance=demo_server" \
    -H "Authorization: Bearer $TOKEN")
echo "$HISTORY" | jq .
echo

# Show current metrics
echo "📋 Current Metrics Snapshot"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━"
curl $CURL_OPTS -X GET $BASE_URL/metrics/current \
    -H "Authorization: Bearer $TOKEN" | jq '.metrics[] | select(.name == "demo_cpu_usage")'
echo

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    💡 Key Takeaways                          ║"
echo "╟──────────────────────────────────────────────────────────────╢"
echo "║ • ONE entity stores entire metric history                    ║"
echo "║ • Each update adds a temporal tag (timestamp|value)         ║" 
echo "║ • Full time-series data preserved automatically             ║"
echo "║ • Can query by time range using temporal features           ║"
echo "║ • Perfect for monitoring, analytics, and observability      ║"
echo "╚══════════════════════════════════════════════════════════════╝"