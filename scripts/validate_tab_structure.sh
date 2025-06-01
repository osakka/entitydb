#!/bin/bash
#
# Tab Structure Validator
# Ensures index.html follows proper tab structure guidelines
#

echo "🔍 Validating tab structure in index.html..."

HTML_FILE="/opt/entitydb/share/htdocs/index.html"
ERRORS=0

# Check for x-show usage on tabs
echo -n "Checking for x-show on tab-content... "
XSHOW_COUNT=$(grep 'x-show="activeTab.*class="tab-content"' "$HTML_FILE" 2>/dev/null | wc -l || echo 0)
if [ "$XSHOW_COUNT" -gt 0 ]; then
    echo "❌ FAIL"
    echo "  Found $XSHOW_COUNT tabs using x-show instead of x-if templates!"
    echo "  Tabs must use: <template x-if=\"activeTab === 'name'\"><div class=\"tab-content\">..."
    ERRORS=$((ERRORS + 1))
else
    echo "✅ PASS"
fi

# Check for x-if templates
echo -n "Checking for proper x-if templates... "
XIF_COUNT=$(grep -c 'x-if="activeTab' "$HTML_FILE" 2>/dev/null || echo 0)
if [ "$XIF_COUNT" -eq 0 ]; then
    echo "❌ FAIL"
    echo "  No x-if templates found for tabs!"
    ERRORS=$((ERRORS + 1))
else
    echo "✅ PASS ($XIF_COUNT tab templates found)"
fi

# Check for nested tab-content
echo -n "Checking for nested tab-content... "
if grep -q 'tab-content.*tab-content' "$HTML_FILE"; then
    echo "❌ FAIL"
    echo "  Found nested tab-content elements!"
    ERRORS=$((ERRORS + 1))
else
    echo "✅ PASS"
fi

# Check CSS structure
echo -n "Checking main-content CSS... "
if ! grep -q 'overflow: hidden' "$HTML_FILE"; then
    echo "⚠️  WARNING"
    echo "  main-content might not have overflow: hidden"
else
    echo "✅ PASS"
fi

# Check for absolute positioning on tabs
echo -n "Checking for absolute positioned tabs... "
if grep -A5 'tab-content' "$HTML_FILE" | grep -q 'position: absolute'; then
    echo "⚠️  WARNING"
    echo "  Found absolute positioning near tab-content (might cause overlap)"
fi

# Summary
echo ""
if [ "$ERRORS" -eq 0 ]; then
    echo "✅ All tab structure validations passed!"
    exit 0
else
    echo "❌ Found $ERRORS tab structure errors!"
    echo ""
    echo "To fix:"
    echo "1. Replace all x-show=\"activeTab...\" with x-if templates"
    echo "2. See docs/development/TAB_STRUCTURE_GUIDELINES.md"
    exit 1
fi