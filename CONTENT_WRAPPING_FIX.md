# Content Wrapping Fix for EntityDB

## Issue Summary

EntityDB version 2.14.0 had a critical authentication issue where entity content could become wrapped in multiple layers of `application/octet-stream` JSON objects. This particularly affected user entities, causing login failures with the server logs showing errors like `Password hash is empty for user <id>`.

The multi-level wrapping created nested structures like:

```json
{
  "application/octet-stream": "{\"application/octet-stream\":\"{\\\"username\\\":\\\"admin\\\",\\\"password_hash\\\":\\\"$2a$10$...\\\"}\"}"
}
```

## Root Cause Identified

The root cause was found in the binary storage writer (`/opt/entitydb/src/storage/binary/writer.go`):

**The binary writer was automatically wrapping ALL content in `application/octet-stream`** regardless of the actual content type. This meant:

1. API layer JSON-marshals content (correct)
2. Binary writer wraps it in `application/octet-stream` (incorrect for JSON)
3. Result: Double-wrapped content that breaks parsing

The fix was simple: **Respect the content type from entity tags** and only use `application/octet-stream` for binary data, not JSON content.

## Implemented Solutions

### Root Cause Fix (Integrated)

The **root cause has been permanently fixed** in the binary storage writer (`/opt/entitydb/src/storage/binary/writer.go`):

**Changes made:**
- Binary writer now respects content type from entity tags
- JSON content (marked with `content:type:application/json`) is stored directly without wrapping
- Only binary data uses `application/octet-stream` wrapper
- Admin user creation now properly tags content as `application/json`

### Backward Compatibility

The login handler in `/opt/entitydb/src/main.go` includes fallback logic to handle existing wrapped content while new content will be clean.

### Additional Tools

### 1. Enhanced Content Tracing

We implemented tracing to track entity content through its lifecycle and identify where wrapping occurs:

- `/opt/entitydb/src/tools/diagnostics/trace_content_wrapping.go`: Tool to trace content through operations
- `/opt/entitydb/src/tools/diagnostics/run_wrapping_test.sh`: Script to run and analyze tracing

### 2. Login Handler Patch

We created a robust content unwrapping solution for the login handler:

- `/opt/entitydb/src/api/auth_login_patch.go`: Enhanced extraction of user data from wrapped content
- Features:
  - Unwraps content recursively up to 5 levels deep
  - Handles multiple JSON formats and content structures
  - Uses fallback strategies when standard parsing fails

### 3. Content Fix Tools

We developed tools to fix affected content:

- `/opt/entitydb/src/tools/fixes/patch_login_handler.go`: Creates the login handler patch
- `/opt/entitydb/src/tools/fixes/fix_user_content.go`: Fixes wrapped content in user entities

### 4. Documentation

We updated documentation to describe the issue and solutions:

- `/opt/entitydb/docs/troubleshooting/CONTENT_FORMAT_TROUBLESHOOTING.md`: Added content wrapping issue section
- `/opt/entitydb/FIXED_LOGIN.md`: Created a marker file indicating the fix is applied

## How to Apply the Fix

**The fix is now permanently integrated into the codebase.** Simply rebuild and restart the server:

1. **Build the fixed server**:
   ```bash
   cd /opt/entitydb/src
   go build -o /opt/entitydb/bin/entitydb
   ```

2. **Restart the server**:
   ```bash
   ./bin/entitydbd.sh restart
   ```

### Optional: Fix existing wrapped content

If you have existing user entities with wrapped content, you can fix them using:

```bash
cd /opt/entitydb/src
go run ./tools/fixes/fix_user_content.go -verbose
```

## Additional Fix: Reader Format Compatibility

After implementing the writer fix, it was discovered that the binary reader was still expecting the old multi-content format while the writer was now using the new unified content format. This caused a reader/writer format mismatch.

**Fix Applied**: Updated `/opt/entitydb/src/storage/binary/reader.go` to handle the new unified content format:

1. **Single Content Parsing**: Changed from multi-content loop to single content type + blob parsing
2. **Format Alignment**: Reader now matches writer's format (content type → content data → timestamp)  
3. **Direct Storage**: Content stored directly without unnecessary JSON conversion

**Files Modified**:
- `/opt/entitydb/src/storage/binary/reader.go` - Updated parseEntity method
- Removed unused imports and added `strings` package

## Verification

After applying the complete fix:

1. Test the admin login using default credentials:
   ```bash
   curl -sk -X POST "https://localhost:8085/api/v1/auth/login" \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "admin"}'
   ```

2. Verify successful response with JWT token
3. Check server startup shows: `Default admin user already exists and is working.`
4. Check server logs show no "Password hash is empty" errors

## ReindexTags Method Added

As part of this work, the missing `ReindexTags()` method was added to EntityRepository:

- **Location**: `/opt/entitydb/src/storage/binary/entity_repository.go`
- **Function**: Rebuilds all tag indexes from scratch with proper locking
- **API Endpoint**: `POST /api/v1/patches/reindex-tags`
- **Usage**: For fixing corrupted or inconsistent tag indexes

## Future Prevention

To prevent content wrapping in future development:

1. Ensure reader and writer formats stay synchronized
2. Test binary format changes with integration tests
3. Add validation to detect format mismatches
4. Consider adding automated tests for content storage/retrieval
5. Document binary format changes in commit messages

---

Fix implemented: May 22, 2025  
Complete fix version: 2.14.1+
Commits: 6b19ebc, 82515da, 98ffbfb