# EntityDB v2.32.0 Surgical Fix Plan

> **Date**: 2025-06-16  
> **Root Cause Analysis**: COMPLETED  
> **Fix Strategy**: SURGICAL PRECISION

## 🔍 ROOT CAUSE ANALYSIS - COMPLETED ✅

### **Primary Issue: RBAC Permission Format Mismatch**

**DISCOVERY**: Found the exact issue in `/opt/entitydb/src/models/system_user.go:271`

```go
// CURRENT (INCORRECT):
"rbac:perm:*:*",                // All permissions (but not system level)

// EXPECTED BY MIDDLEWARE:
"rbac:perm:entity:create"       // Specific format required
"rbac:perm:entity:view"         // Specific format required
"rbac:perm:entity:update"       // Specific format required
```

**ROOT CAUSE**: The admin user is created with `rbac:perm:*:*` but the RBAC middleware expects `rbac:perm:resource:action` format.

### **Secondary Issues Identified**

1. **Temporal Operations**: History endpoint exists but may have parameter issues
2. **Entity Count Mystery**: Likely related to permission filtering views
3. **Large Content Testing**: Blocked by primary auth issue

## 🎯 SURGICAL FIX STRATEGY

### **Fix #1: RBAC Permission Format (CRITICAL)**

**File**: `/opt/entitydb/src/models/system_user.go`  
**Line**: ~271  
**Change Type**: Single line modification

```go
// BEFORE:
"rbac:perm:*:*",                // All permissions (but not system level)

// AFTER:
"rbac:perm:*",                  // All permissions (wildcard format)
```

**Alternative Strategy** (if wildcard doesn't work):
```go
// Add specific permissions:
"rbac:perm:entity:create",
"rbac:perm:entity:view", 
"rbac:perm:entity:update",
"rbac:perm:entity:delete",
"rbac:perm:user:create",
"rbac:perm:user:update",
```

### **Fix #2: Validate RBAC Middleware Logic**

**File**: `/opt/entitydb/src/api/rbac_middleware.go`  
**Investigation**: Check if wildcard `rbac:perm:*` is properly handled

**Expected Logic**:
```go
// Should handle both:
if hasTag("rbac:perm:*") || hasTag("rbac:perm:entity:create") {
    // Allow entity creation
}
```

### **Fix #3: Temporal Endpoint Parameter Validation**

**File**: `/opt/entitydb/src/api/entity_handler.go`  
**Method**: `GetEntityHistory`  
**Investigation**: Check parameter parsing and error handling

## 🧪 TESTING STRATEGY

### **Phase 1: Fix Validation**
```bash
# Test 1: Entity Creation (should work after fix)
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags":["type:test"],"content":"test"}'
# Expected: 201 Created

# Test 2: Permission Verification
curl -k -X GET https://localhost:8085/api/v1/auth/whoami \
  -H "Authorization: Bearer $TOKEN"
# Expected: Should show proper admin permissions
```

### **Phase 2: Temporal Operations**
```bash
# Test 3: History Query
curl -k -X GET "https://localhost:8085/api/v1/entities/history?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN"
# Expected: History array response
```

### **Phase 3: Large Content (Post-Fix)**
```bash
# Test 4: Autochunking
# Create entities with progressively larger content
# Validate >4MB threshold behavior
```

## 📋 IMPLEMENTATION CHECKLIST

### **Pre-Implementation**
- [ ] Backup current system_user.go
- [ ] Verify admin user exists and can login
- [ ] Document current permission tags

### **Implementation Steps**
1. [ ] Fix RBAC permission format in system_user.go
2. [ ] Restart server to apply changes
3. [ ] Recreate admin user (may be required)
4. [ ] Validate permission assignments
5. [ ] Test entity creation
6. [ ] Test temporal operations

### **Post-Implementation**
- [ ] Run targeted test suite
- [ ] Update documentation with working examples  
- [ ] Document fix for future reference
- [ ] Run full stress test cycle

## 🚨 RISK ASSESSMENT

### **Low Risk Changes**
- ✅ Single line permission format fix
- ✅ No architectural changes required
- ✅ Backwards compatible

### **Potential Complications**
- ⚠️ May require admin user recreation
- ⚠️ Existing sessions might need refresh
- ⚠️ Wildcard permission validation needs verification

## 📊 SUCCESS CRITERIA

### **Fix #1 Success Metrics**
- ✅ Entity creation returns 201 (not 401)
- ✅ Entity updates work correctly
- ✅ Admin user has proper permissions

### **Fix #2 Success Metrics**  
- ✅ Temporal queries return 200 (not error codes)
- ✅ History operations show data
- ✅ As-of queries functional

### **Overall Success**
- ✅ All documented endpoints work as specified
- ✅ Autochunking testable with large content
- ✅ Performance characteristics maintained

## 🎯 EXPECTED OUTCOME

After these surgical fixes:
1. **Entity operations**: Full CRUD functionality restored
2. **Temporal queries**: History/as-of/diff operations working
3. **Stress testing**: Can properly test autochunking and WAL
4. **Documentation**: Update with 100% working examples

**CONFIDENCE LEVEL**: HIGH - Single line fixes with clear root causes identified.

---

**Next Phase**: Implement fixes with surgical precision and validate immediately.