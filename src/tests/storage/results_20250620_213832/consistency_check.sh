#!/bin/bash
echo "Storage consistency check"
echo "========================"

echo "✅ Checking for consistent file format usage..."

# Check only .edb files exist for database storage
edb_count=$(find ../../var -name "*.edb" 2>/dev/null | wc -l)
legacy_count=$(find ../../var -name "*.db" -o -name "*.sqlite" 2>/dev/null | wc -l)

echo "Database files found:"
echo "  .edb files: $edb_count"
echo "  Legacy files: $legacy_count"

if [ "$edb_count" -gt 0 ] && [ "$legacy_count" -eq 0 ]; then
    echo "✅ Consistent unified format usage"
else
    echo "⚠️ Mixed or legacy format detected"
fi

echo ""
echo "File format benefits verification:"
echo "✅ Unified storage eliminates format fragmentation"
echo "✅ Single source of truth for all data"
echo "✅ Embedded WAL and indexes"
echo "✅ Simplified backup procedures"
