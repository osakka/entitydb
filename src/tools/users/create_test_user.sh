#!/bin/bash
# Create a test user with admin privileges

# Generate a hashed password using bcrypt algorithm
# We'll use a precomputed hash for 'testpass'
HASHED_PASSWORD='$2a$10$HT3X8GlUY1RXdIKbS1Lw4.q1qCHNvYjSfhvUTwKZFOBRMxZFmYGPG'

# Create a user entity with admin privileges
cat > /tmp/test_user.json << EOF
{
  "id": "user_tester",
  "tags": [
    "type:user",
    "id:username:tester",
    "rbac:role:admin",
    "rbac:perm:*",
    "status:active"
  ],
  "content": {
    "username": "tester",
    "password_hash": "$HASHED_PASSWORD",
    "display_name": "Test User"
  }
}
EOF

# Restart the server to reset the db
echo "Restarting server..."
cd /opt/entitydb && ./bin/entitydbd.sh restart

echo "Creating test user..."
cd /opt/entitydb && ./bin/entitydb --debug-create-user=/tmp/test_user.json

# Clean up
rm /tmp/test_user.json