# Secure Password Implementation in EntityDB

## Overview

This document describes the secure password handling implementation for the EntityDB platform's entity-based architecture. The implementation uses bcrypt for password hashing and provides backward compatibility with the legacy authentication system.

## Implementation Details

### Password Hashing

The system uses the [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) algorithm for password hashing, which is an industry-standard approach for secure password storage. Bcrypt automatically handles salting and has built-in protection against brute force attacks through its configurable work factor.

```go
// hashPassword securely hashes a password using bcrypt
func hashPassword(password string) (string, error) {
    // Generate a bcrypt hash with cost factor 12 (recommended default)
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return "", err
    }
    return string(hashedBytes), nil
}

// validatePassword checks if the provided password matches the stored hash
func validatePassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Entity-Based User Storage

User passwords are now stored as hashed values in the user entity's properties:

```json
{
  "id": "entity_user_admin",
  "type": "user",
  "title": "admin",
  "description": "Admin user account for admin",
  "status": "active",
  "tags": ["user", "admin"],
  "properties": {
    "username": "admin",
    "roles": ["admin"],
    "password_hash": "$2a$12$AbCdEfGhIjKlMnOpQrStUvWxYz01234567890AbCdEfGhIj"
  },
  "created_at": "2023-08-15T10:30:00Z",
  "created_by": "system"
}
```

### Authentication Flow

The new authentication flow first attempts to use the entity-based authentication with secure password hashing, then falls back to the legacy authentication if needed:

1. When a user attempts to log in, the system looks up the user in the traditional user store
2. If the user exists, the system tries to find a matching user entity
3. If a user entity is found, the system extracts the password hash from the entity properties and validates the provided password against it using bcrypt
4. If entity-based authentication fails or isn't available, the system falls back to the legacy direct password comparison for backward compatibility
5. The authentication succeeds only if either the entity-based or legacy authentication validates the credentials

```go
// Authentication code snippet
authenticated := false
var user *User

// Check if user exists in traditional user store
var exists bool
user, exists = s.users[req.Username]

if exists {
    // Try to find matching user entity
    var userEntity map[string]interface{}
    entityID := "entity_user_" + user.Username
    
    // Check if entity exists in entity storage
    if entity, ok := s.entities[entityID]; ok {
        userEntity = entity
        
        // Extract password hash from entity properties
        if properties, ok := userEntity["properties"].(map[string]interface{}); ok {
            if passwordHash, ok := properties["password_hash"].(string); ok {
                // Validate using bcrypt password hash
                if validatePassword(req.Password, passwordHash) {
                    authenticated = true
                }
            }
        }
    }
    
    // Fall back to direct password comparison if entity-based auth failed
    if !authenticated {
        if user.Password == req.Password {
            authenticated = true
        }
    }
}
```

### User Creation

When creating a new user, the system:

1. Validates the password (minimum length of 8 characters)
2. Hashes the password using bcrypt
3. Creates a traditional user for backward compatibility
4. Creates a user entity with the hashed password
5. Stores both representations for dual compatibility

```go
// Hash the password
passwordHash, err := hashPassword(req.Password)
if err != nil {
    // Handle error
    return
}

// Create user with original password (for legacy support)
newUser := &User{
    ID:       "usr_" + req.Username,
    Username: req.Username,
    Password: req.Password, // Keep original for backward compatibility
    Roles:    req.Roles,
}

// Store user
s.users[req.Username] = newUser

// Create user entity with hashed password
entityID := "entity_user_" + req.Username
userEntity := map[string]interface{}{
    "id":          entityID,
    "type":        "user",
    "title":       req.Username,
    "description": fmt.Sprintf("User account for %s", req.Username),
    "status":      "active",
    "tags":        append([]string{"user"}, req.Roles...),
    "properties": map[string]interface{}{
        "username":      req.Username,
        "roles":         req.Roles,
        "password_hash": passwordHash,
    },
    "created_at": time.Now().Format(time.RFC3339),
    "created_by": "system",
}

// Store the user entity
s.entities[entityID] = userEntity
```

## Security Considerations

1. **Password Strength**: The system enforces a minimum password length of 8 characters. For production environments, consider implementing additional password complexity requirements.

2. **Work Factor**: The bcrypt implementation uses a cost factor of 12, which provides a good balance between security and performance. This can be adjusted based on your specific needs.

3. **Legacy Compatibility**: The system maintains backward compatibility with plaintext passwords in the legacy user store. In a production environment, consider migrating all users to the entity-based system with secure password hashing and retiring the legacy system.

4. **Password Updates**: When a user changes their password, the system updates both the legacy user record and the user entity with the new password (hashed in the entity).

## Testing

A test script is available at `/opt/entitydb/share/tests/entity/test_secure_password.sh` to verify the secure password implementation. The script:

1. Creates a test user through the entity API
2. Verifies that the password is stored as a bcrypt hash
3. Tests successful authentication with the correct password
4. Tests failed authentication with an incorrect password
5. Cleans up by deleting the test user

Run the test script to verify that the secure password implementation is working correctly:

```bash
/opt/entitydb/share/tests/entity/test_secure_password.sh
```

## Future Improvements

1. **Password Rotation**: Implement policies for password expiration and rotation
2. **Failed Login Tracking**: Add tracking for failed login attempts to prevent brute force attacks
3. **Multi-factor Authentication**: Add support for additional authentication factors
4. **Password Complexity Rules**: Add more sophisticated password complexity requirements
5. **Complete Legacy Retirement**: Once all users have been migrated to the entity-based system, remove the legacy authentication system completely