# Tab Structure Guidelines

## Critical Rule: NEVER Use x-show for Tab Content

### The Problem
Using `x-show` with flex layouts causes tabs to not display properly. The Storage tab would show, but other tabs would be invisible even when selected.

### The Solution
Always use `x-if` with template tags for tab content:

```html
<!-- ❌ WRONG - DO NOT USE -->
<div x-show="activeTab === 'tabname'" class="tab-content">
    <!-- content -->
</div>

<!-- ✅ CORRECT - ALWAYS USE THIS -->
<template x-if="activeTab === 'tabname'">
    <div class="tab-content">
        <!-- content -->
    </div>
</template>
```

## Why This Matters

1. **x-show Issues with Flex**:
   - `x-show` only toggles `display: none`
   - Hidden elements can still affect flex layout calculations
   - Multiple tabs with `flex: 1` conflict even when hidden

2. **x-if Benefits**:
   - Completely removes elements from DOM when false
   - Only active tab exists in DOM
   - No layout conflicts
   - Better performance

## Required CSS Structure

```css
/* Main container must be flex with overflow hidden */
.main-content {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

/* Tab content takes remaining space */
.tab-content {
    flex: 1;
    overflow: auto;
    padding: 24px;
}
```

## Validation

A validator script is included: `/js/tab-validator.js`

It checks for:
- Any tabs using x-show instead of x-if
- Nested tab-content elements
- Proper main-content CSS
- Overall structure integrity

## Common Mistakes to Avoid

1. **Using x-show for tabs** - Always use x-if
2. **Absolute positioning tabs** - Causes header overlap
3. **Forgetting overflow: hidden on main-content** - Causes layout issues
4. **Not using templates** - x-if requires template tags
5. **Nesting tab-content divs** - Each tab should be independent

## Testing New Tabs

When adding a new tab:
1. Use the template pattern above
2. Ensure it's a direct child of main-content
3. Test all tabs still work
4. Run the validator in console: `new TabValidator().validateTabs()`

## Emergency Fix

If tabs stop showing:
1. Check browser console for validator errors
2. Verify all tabs use x-if templates
3. Check no CSS changes to main-content or tab-content
4. Ensure no nested tab structures