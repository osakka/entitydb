# EntityDB Quick Start Guide

Welcome to EntityDB! This guide will help you get started quickly with EntityDB v2.29.0.

> **⚠️ Important**: v2.29.0 includes breaking changes to authentication. See the [migration guide](../api/auth.md#v229-migration) if upgrading.

## Installation

1. Clone the repository:
   ```bash
   git clone https://git.home.arpa/itdlabs/entitydb.git
   cd entitydb
   ```

2. Build the server:
   ```bash
   cd src
   make
   ```

## Starting the Server

```bash
# Start the server daemon
./bin/entitydbd.sh start

# Check status
./bin/entitydbd.sh status

# Stop the server
./bin/entitydbd.sh stop
```

The server runs on:
- HTTP: port 8085 (default)
- HTTPS: port 8443 (when SSL is enabled)

## Default Admin User

The server automatically creates a default admin user on first start:
- Username: `admin`
- Password: `admin`

## Using the API

1. Login with default credentials:
   ```bash
   TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' | jq -r '.token')
   ```

2. Create an entity:
   ```bash
   curl -k -X POST https://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "tags": ["type:issue", "status:pending", "priority:high"],
       "content": "My First Issue"
     }'
   ```

3. List entities:
   ```bash
   curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=type:issue" \
     -H "Authorization: Bearer $TOKEN"
   ```

4. Get a specific entity:
   ```bash
   curl -k -X GET "https://localhost:8085/api/v1/entities/get?id=entity_123" \
     -H "Authorization: Bearer $TOKEN"
   ```

## Web Interface

Open your browser to http://localhost:8085 to access the web dashboard.

## Next Steps

- Read the [Architecture Overview](../architecture/overview.md)
- Learn about [Tag Namespaces](../architecture/tags.md)
- Explore the [API Reference](../api/entities.md)
- Check out [API Examples](../api/examples.md)