#!/bin/bash
# Analyze metrics in the database

BASE_URL="https://localhost:8085/api/v1"
CURL_OPTS="-k -s"

echo "=== EntityDB Metrics Analysis ==="
echo

# Login
TOKEN=$(curl $CURL_OPTS -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "1. OLD APPROACH (Wrong - One entity per data point):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
OLD_COUNT=$(curl $CURL_OPTS -X GET "$BASE_URL/entities/query?tags=hub:metrics" \
  -H "Authorization: Bearer $TOKEN" | jq '.entities | length')
echo "   Found $OLD_COUNT entities using old format"
echo "   Pattern: hub:metrics, metrics:self:name:*, metrics:self:value:*"
echo "   Problem: Each metric value creates a NEW entity!"
echo

# Show sample old metrics
echo "   Sample old metrics (each is a separate entity):"
curl $CURL_OPTS -X GET "$BASE_URL/entities/query?tags=hub:metrics&limit=3" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.entities[] | "   • ID: \(.id[0:8])... Tags: \(.tags | join(", "))"'
echo

echo "2. NEW APPROACH (Correct - One entity per metric type):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
NEW_COUNT=$(curl $CURL_OPTS -X GET "$BASE_URL/entities/query?tags=type:metric" \
  -H "Authorization: Bearer $TOKEN" | jq '.entities | length')
echo "   Found $NEW_COUNT entities using new format"
echo "   Pattern: type:metric, metric:name:*, temporal value tags"
echo "   Benefit: ONE entity stores entire history!"
echo

# List new format metrics
echo "   Our temporal metrics:"
curl $CURL_OPTS -X GET "$BASE_URL/metrics/current" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.metrics[] | "   • \(.name) (\(.instance)): \(.value)\(.unit) - Entity ID: \(.id)"'
echo

echo "3. COMPARISON:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   Old approach: $OLD_COUNT entities (wasteful!)"
echo "   New approach: $NEW_COUNT entities (efficient!)"
echo "   Space savings: $(echo "scale=1; ($OLD_COUNT - $NEW_COUNT) / $OLD_COUNT * 100" | bc)%"
echo

echo "4. CLEANUP RECOMMENDATION:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   To remove old metrics (if desired):"
echo "   curl -X DELETE '$BASE_URL/entities/bulk?tags=hub:metrics'"
echo "   (Note: This would need to be implemented or done entity by entity)"
echo

echo "5. WHY THIS MATTERS:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   Old: 1 metric collected every 5 seconds = 17,280 entities/day"
echo "   New: 1 metric collected every 5 seconds = 1 entity (with 17,280 temporal tags)"
echo "   That's a 17,280x reduction in entities!"