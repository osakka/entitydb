#!/bin/bash

# Populate Worca with realistic metrics data
# This script creates temporal metrics for the Worca dashboard

echo "=== Populating Worca Metrics ==="

# Get auth token
TOKEN=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Failed to authenticate"
    exit 1
fi

echo "Authenticated successfully"

# Function to add a metric value
add_metric() {
    local metric_name=$1
    local instance=$2
    local value=$3
    local unit=$4
    local labels=$5
    
    curl -k -s -X POST https://localhost:8085/api/v1/metrics/collect \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"metric_name\": \"$metric_name\",
        \"metric\": \"$metric_name\",
        \"instance\": \"$instance\",
        \"value\": $value,
        \"unit\": \"$unit\",
        \"labels\": $labels
      }" > /dev/null
    
    echo -n "."
}

echo -n "Adding task completion metrics"
# Task completion rate over time
add_metric "worca_tasks_completed" "worca_hub" 15 "tasks" '{"project":"ocean-platform","team":"engineering"}'
add_metric "worca_tasks_completed" "worca_hub" 23 "tasks" '{"project":"ocean-platform","team":"engineering"}'
add_metric "worca_tasks_completed" "worca_hub" 28 "tasks" '{"project":"ocean-platform","team":"engineering"}'
add_metric "worca_tasks_completed" "worca_hub" 35 "tasks" '{"project":"ocean-platform","team":"engineering"}'
add_metric "worca_tasks_completed" "worca_hub" 42 "tasks" '{"project":"ocean-platform","team":"engineering"}'
echo " ✓"

echo -n "Adding team productivity metrics"
# Team productivity metrics
add_metric "worca_team_productivity" "worca_hub" 78.5 "percent" '{"team":"engineering","sprint":"current"}'
add_metric "worca_team_productivity" "worca_hub" 82.3 "percent" '{"team":"engineering","sprint":"current"}'
add_metric "worca_team_productivity" "worca_hub" 85.7 "percent" '{"team":"engineering","sprint":"current"}'
add_metric "worca_team_productivity" "worca_hub" 89.2 "percent" '{"team":"engineering","sprint":"current"}'
echo " ✓"

echo -n "Adding sprint velocity metrics"
# Sprint velocity
add_metric "worca_sprint_velocity" "worca_hub" 32 "story_points" '{"team":"engineering","sprint":"1"}'
add_metric "worca_sprint_velocity" "worca_hub" 38 "story_points" '{"team":"engineering","sprint":"2"}'
add_metric "worca_sprint_velocity" "worca_hub" 41 "story_points" '{"team":"engineering","sprint":"3"}'
add_metric "worca_sprint_velocity" "worca_hub" 45 "story_points" '{"team":"engineering","sprint":"4"}'
echo " ✓"

echo -n "Adding task status distribution"
# Task status distribution
add_metric "worca_task_status" "worca_hub" 25 "count" '{"status":"todo","project":"ocean-platform"}'
add_metric "worca_task_status" "worca_hub" 18 "count" '{"status":"in_progress","project":"ocean-platform"}'
add_metric "worca_task_status" "worca_hub" 42 "count" '{"status":"done","project":"ocean-platform"}'
add_metric "worca_task_status" "worca_hub" 5 "count" '{"status":"blocked","project":"ocean-platform"}'
echo " ✓"

echo -n "Adding team member utilization"
# Team member utilization
for member in "alice" "bob" "charlie" "diana" "eve"; do
    utilization=$((60 + RANDOM % 40))
    add_metric "worca_member_utilization" "worca_hub" $utilization "percent" "{\"member\":\"$member\",\"team\":\"engineering\"}"
done
echo " ✓"

echo -n "Adding epic progress metrics"
# Epic progress tracking
add_metric "worca_epic_progress" "worca_hub" 65 "percent" '{"epic":"ocean-navigation","project":"ocean-platform"}'
add_metric "worca_epic_progress" "worca_hub" 82 "percent" '{"epic":"whale-communication","project":"ocean-platform"}'
add_metric "worca_epic_progress" "worca_hub" 45 "percent" '{"epic":"deep-sea-exploration","project":"ocean-platform"}'
echo " ✓"

echo -n "Adding system health metrics"
# System health metrics
add_metric "worca_api_response_time" "worca_hub" 125 "ms" '{"endpoint":"tasks","method":"GET"}'
add_metric "worca_api_response_time" "worca_hub" 85 "ms" '{"endpoint":"projects","method":"GET"}'
add_metric "worca_api_response_time" "worca_hub" 210 "ms" '{"endpoint":"analytics","method":"POST"}'

add_metric "worca_active_users" "worca_hub" 12 "users" '{"timeframe":"realtime"}'
add_metric "worca_active_sessions" "worca_hub" 18 "sessions" '{"timeframe":"realtime"}'
echo " ✓"

echo -n "Adding burndown chart data"
# Burndown chart data (story points remaining)
for day in {1..10}; do
    remaining=$((100 - day * 8 + RANDOM % 5))
    add_metric "worca_sprint_burndown" "worca_hub" $remaining "story_points" "{\"sprint\":\"current\",\"day\":$day}"
done
echo " ✓"

echo ""
echo "✅ Metrics population complete!"

# Verify metrics were created
echo ""
echo "Verifying metrics..."
METRIC_COUNT=$(curl -k -s -H "Authorization: Bearer $TOKEN" \
  "https://localhost:8085/api/v1/entities/list?tags=type:metric" | jq '.entities | length')

echo "Total metric entities created: $METRIC_COUNT"

# Show sample metric values
echo ""
echo "Sample metric values:"
curl -k -s -H "Authorization: Bearer $TOKEN" \
  "https://localhost:8085/api/v1/entities/list?tags=type:metric&limit=3" | \
  jq -r '.entities[] | "- \(.tags | map(select(startswith("metric:name:") or startswith("metric:value:"))) | join(", "))"'