# Audit Logging Implementation

## Overview

This document describes the security audit logging system implemented for the EntityDB platform. The audit logging system captures security-relevant events and stores them in a structured format for compliance, intrusion detection, and forensic analysis.

## Audit Log Structure

Each audit log entry is a JSON object with the following structure:

```json
{
  "timestamp": "2023-08-15T10:30:00Z",
  "event_type": "authentication",
  "user_id": "usr_admin",
  "username": "admin",
  "entity_id": "entity_001",
  "entity_type": "issue",
  "action": "login",
  "status": "success",
  "ip": "192.168.1.1",
  "request_path": "/api/v1/auth/login",
  "details": {
    "additional_info": "Any additional context-specific information"
  }
}
```

### Field Descriptions

| Field | Description | Required |
|-------|-------------|----------|
| timestamp | ISO 8601 formatted timestamp of the event | Yes |
| event_type | Type of event (authentication, access_control, entity, administrative, system) | Yes |
| user_id | ID of the user performing the action (if available) | No |
| username | Username of the user performing the action (if available) | No |
| entity_id | ID of the entity being acted upon (if applicable) | No |
| entity_type | Type of entity being acted upon (if applicable) | No |
| action | Specific action being performed (e.g., "login", "create", "delete") | Yes |
| status | Outcome of the action (e.g., "success", "failure", "denied") | Yes |
| ip | IP address of the client (if available) | No |
| request_path | API endpoint path (if applicable) | No |
| details | Additional context-specific information | No |

## Event Types

The audit logging system categorizes events into the following types:

1. **Authentication Events**
   - User login attempts (success and failure)
   - Password changes
   - Token generation and validation

2. **Access Control Events**
   - Permission checks
   - Authorization decisions
   - Role-based access control enforcement

3. **Entity Events**
   - Creation, modification, or deletion of entities
   - Entity relationship changes
   - Entity state transitions

4. **Administrative Events**
   - User management operations
   - System configuration changes
   - Security policy modifications

5. **System Events**
   - Server startup and shutdown
   - Audit log initialization and rotation
   - Critical system events

## Implementation Details

The audit logging system is implemented in the `AuditLogger` struct in `/opt/entitydb/src/audit_logger.go`. It provides the following functionality:

### Logger Initialization

```go
// Create new audit logger
auditLogger, err := NewAuditLogger("/opt/entitydb/var/log/audit", server.entities)
if err != nil {
    log.Fatalf("Failed to initialize audit logger: %v", err)
}
```

### Logging Events

The audit logger provides specialized methods for different event types:

```go
// Log authentication event
auditLogger.LogAuthEvent(user.ID, user.Username, "login", "success", clientIP, nil)

// Log access control event
auditLogger.LogAccessEvent(user.ID, user.Username, "read", "success", r.URL.Path, nil)

// Log entity event
auditLogger.LogEntityEvent(user.ID, user.Username, entity.ID, entity.Type, "create", "success", nil)

// Log administrative event
auditLogger.LogAdminEvent(user.ID, user.Username, "config_change", "success", configDetails)
```

### Log File Management

The audit logger writes to daily log files with automatic rotation:

- Log files are stored in `/opt/entitydb/var/log/audit/`
- Log files are named with the pattern `entitydb_audit_YYYY-MM-DD.log`
- Rotation occurs automatically at midnight or can be triggered manually

### Integration Points

The audit logging system is integrated at the following points in the codebase:

1. **Authentication Handlers**
   - `handleLogin` - Log login attempts
   - `validateToken` - Log token validation events

2. **RBAC Enforcement**
   - `checkAuth` - Log authentication checks
   - `checkAdminRole` - Log admin role checks

3. **Entity API**
   - `handleEntityAPI` - Log entity operations
   - `handleEntityRelationshipAPI` - Log relationship operations

4. **Administrative Actions**
   - All administrative endpoints log actions

## Security Considerations

1. **Log Integrity**: Audit logs are append-only to prevent tampering
2. **Sensitive Data**: Passwords and sensitive data are never logged
3. **Log Persistence**: Logs are stored on persistent storage
4. **Log Rotation**: Automatic log rotation prevents logs from growing too large
5. **Error Handling**: Logging failures are properly handled to prevent information loss

## Testing

A test script is available at `/opt/entitydb/share/tests/entity/test_audit_logging.sh` to verify the audit logging implementation. The script tests:

1. Entity creation audit logging
2. User creation audit logging
3. Authentication event logging (successful and failed logins)
4. Access control event logging
5. Administrative action logging

Run the test script to verify that the audit logging implementation is working correctly:

```bash
/opt/entitydb/share/tests/entity/test_audit_logging.sh
```

## Log Analysis

The audit logs can be analyzed using standard JSON processing tools:

```bash
# Count events by type
jq -r '.event_type' /opt/entitydb/var/log/audit/entitydb_audit_2023-08-15.log | sort | uniq -c

# Find failed login attempts
jq -r 'select(.event_type=="authentication" and .status=="failure")' /opt/entitydb/var/log/audit/entitydb_audit_*.log

# Find administrative actions by a specific user
jq -r 'select(.event_type=="administrative" and .username=="admin")' /opt/entitydb/var/log/audit/entitydb_audit_*.log
```

## Future Enhancements

1. **Centralized Logging**: Integration with centralized logging systems (ELK, Splunk, etc.)
2. **Real-time Alerting**: Implement real-time alerts for suspicious events
3. **Log Encryption**: Encrypt sensitive log entries
4. **Digital Signatures**: Add digital signatures to log entries for tamper detection
5. **Advanced Querying**: Implement advanced querying capabilities for log analysis