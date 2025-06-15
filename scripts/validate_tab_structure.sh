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

# Check for tab implementation (Vue/Alpine or vanilla JS)
echo -n "Checking for proper tab implementation... "
TAB_EXISTS=$(grep -c 'tab-content\|activeTab' "$HTML_FILE" 2>/dev/null || echo 0)
TAB_EXISTS=$(echo "$TAB_EXISTS" | tr -d '\n\r')
if [ "$TAB_EXISTS" -eq 0 ]; then
    echo "✅ PASS (No tab structure found)"
else
    # Check for Vue/Alpine conditional classes
    VUE_COUNT=$(grep -c ':class=.*active.*activeTab' "$HTML_FILE" 2>/dev/null || echo 0)
    VUE_COUNT=$(echo "$VUE_COUNT" | tr -d '\n\r')
    
    # Check for vanilla JS tab implementation
    VANILLA_COUNT=$(grep -c 'switchTab\|data-tab=' "$HTML_FILE" 2>/dev/null || echo 0)
    VANILLA_COUNT=$(echo "$VANILLA_COUNT" | tr -d '\n\r')
    
    if [ "$VUE_COUNT" -gt 0 ]; then
        echo "✅ PASS (Vue/Alpine tab implementation found)"
    elif [ "$VANILLA_COUNT" -gt 0 ]; then
        echo "✅ PASS (Vanilla JS tab implementation found)"
    else
        echo "❌ FAIL"
        echo "  Tab structure found but no proper implementation!"
        ERRORS=$((ERRORS + 1))
    fi
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