#!/bin/bash
# Fix authentication for tests

# Set execution path
cd "$(dirname "$0")"

# Import common utilities
source "test_utils.sh"

echo "Attempting to fix authentication for tests..."

# Reset admin password 
echo "Resetting admin password..."
sqlite3 ../../../var/db/entitydb.db "UPDATE users SET password_hash = '\$2a\$10\$fDkwkDSBIcYYSW0Kb3XtBuWGK6PCN1zdTQn47IrktED.y.9QYIqGq' WHERE username = 'admin';"

# Verify update
ADMIN_COUNT=$(sqlite3 ../../../var/db/entitydb.db "SELECT COUNT(*) FROM users WHERE username = 'admin';")
if [ "$ADMIN_COUNT" -eq "1" ]; then
  echo "Admin user found and updated."
else
  echo "Admin user not found! Creating admin user..."
  sqlite3 ../../../var/db/entitydb.db "INSERT INTO users (id, username, password_hash, email, full_name, display_name, created_at, roles, status) VALUES ('user_admin', 'admin', '\$2a\$10\$fDkwkDSBIcYYSW0Kb3XtBuWGK6PCN1zdTQn47IrktED.y.9QYIqGq', '', 'Administrator', 'Administrator', datetime('now'), 'admin', 'active');"
fi

echo "Testing admin login..."
AUTH_TOKEN=$(get_auth_token "$ADMIN_USERNAME" "$ADMIN_PASSWORD")
if [ -z "$AUTH_TOKEN" ]; then
  echo "Authentication still failing. Check server logs for details."
  exit 1
else
  echo "Authentication successful!"
  echo "Token: $AUTH_TOKEN"
fi

echo "Authentication fix completed."
exit 0