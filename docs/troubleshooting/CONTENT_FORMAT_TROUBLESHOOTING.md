# EntityDB Content Format Troubleshooting

This document addresses content format issues in EntityDB and provides guidance on diagnosing and resolving them, including the recently discovered content wrapping issue.

## Background

EntityDB v2.13.0 uses a custom binary format (EBF) with a unified Entity model where every entity has:
- A unique ID
- Tags (with timestamps)
- Content (as byte array)

The content field's encoding is critical for proper functionality, especially for user entities that need to be processed by the authentication system.

## Known Issues

### Content Wrapping Issue (v2.14.0)

In EntityDB version 2.14.0, we've identified an issue where entity content can become wrapped in multiple layers of `application/octet-stream` JSON objects, particularly affecting user authentication. The problem creates a nested structure like:

```json
{
  "application/octet-stream": "{\"application/octet-stream\":\"{\\\"username\\\":\\\"admin\\\",\\\"password_hash\\\":\\\"$2a$10$...\\\"}\"}"
}
```

This multi-level wrapping causes failed logins, with server logs showing errors like `Password hash is empty for user <id>`.

#### Root Causes:

1. **Inconsistent Content Handling**: Different code paths handle content differently during create/update operations.
2. **Multiple Content Encoding Layers**: Content gets wrapped in additional layers during some API operations.
3. **Migration Scripts**: Some initialization scripts don't properly check current content format.
4. **Mixed Client/Server Processing**: Both client and server may add wrapping layers to content.

#### Resolution:

A patch has been created that adds robust content unwrapping to the login handler. To apply this fix:

1. Run the content unwrapping tool:
   ```bash
   cd /opt/entitydb/src
   go run ./tools/fixes/patch_login_handler.go -verbose
   ```

2. For specific admin user fixes, run:
   ```bash
   cd /opt/entitydb/src
   go run ./tools/fixes/fix_user_content.go -verbose
   ```

3. Restart the server to apply the patch:
   ```bash
   ./bin/entitydbd.sh restart
   ```

### Authentication Failures (500 Error)

If you encounter authentication failures with error "Invalid user data" and HTTP 500 status, this indicates a content format issue rather than incorrect credentials. A proper credential error would return HTTP 401 with "Invalid credentials".

#### Symptoms:
- Login attempts return `{"error":"Invalid user data"}` with 500 status
- User entities appear in database but cannot authenticate
- Error occurs regardless of password used

This problem occurs when the content format of user entities doesn't match what the authentication system expects.

## Content Format Requirements

User entities must have their content stored in the following format:

```
{
  "username": "admin",
  "password_hash": "$2a$10$hash_value_here",
  "display_name": "Administrator"
}
```

This content is then encoded in the database following these rules:
1. The JSON object is stored in a field called `application/octet-stream`
2. The resulting object is Base64 encoded

Properly encoded content for a user looks like:
```
eyJhcHBsaWNhdGlvbi9vY3RldC1zdHJlYW0iOiJ7XCJkaXNwbGF5X25hbWVcIjpcIkFkbWluaXN0cmF0b3JcIixcInBhc3N3b3JkX2hhc2hcIjpcIiQyYSQxMCQzcXlFMzNNWXgwc1FlbFJrMFIwWmFlbzMuajNETUdQY0lvMVRYdi9qRHUzMmY3MjN1SUxCeVwiLFwidXNlcm5hbWVcIjpcImFkbWluXCJ9In0=
```

## Resolution Steps

If you encounter content format issues:

1. **Reset the Database**:
   ```bash
   # Stop the service
   /opt/entitydb/bin/entitydbd.sh stop
   
   # Remove database files
   rm -f /opt/entitydb/var/entities.ebf /opt/entitydb/var/entitydb.wal
   
   # Start the service (will create a fresh admin user)
   /opt/entitydb/bin/entitydbd.sh start
   ```

2. **Verify Admin Login**:
   ```bash
   curl -sk -X POST "https://localhost:8085/api/v1/auth/login" \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "admin"}'
   ```

3. **Create New Users with Correct Format**:
   When creating users, ensure the content format is correct by using the standard API rather than direct database modifications.

## Prevention

To prevent content format issues:

1. Always use the standard API for entity creation, especially for user entities
2. Avoid direct manipulation of the database files
3. Run comprehensive API tests after system changes
4. Back up working database files before major changes

## Diagnostic Tools

The following scripts are available in `/opt/entitydb/share/tests/` to help diagnose content format issues:

- **test_all_endpoints.sh**: Tests all API endpoints to verify system functionality
- **debug_login_500.sh**: Specifically diagnoses login issues
- **try_multiple_passwords.sh**: Tests different password combinations
- **create_simple_format_admin.sh**: Creates an admin user with simplified content format

## Additional Information

For more details on the entity model and binary format, see:
- [CLAUDE.md](/opt/entitydb/CLAUDE.md) for an overview
- [CONTENT_V3_EXAMPLE.md](/opt/entitydb/docs/CONTENT_V3_EXAMPLE.md) for content format examples
- [BINARY_FORMAT_IMPLEMENTATION.md](/opt/entitydb/docs/BINARY_FORMAT_IMPLEMENTATION.md) for details on the binary storage