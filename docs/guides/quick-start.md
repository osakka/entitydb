# EntityDB Quick Start Guide

Welcome to EntityDB! This guide will help you get started quickly.

## Installation

1. Clone the repository:
   ```bash
   git clone https://git.home.arpa/osakka/entitydb.git
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

The server runs on port 8085 by default.

## Default Admin User

The server automatically creates a default admin user on first start:
- Username: `admin`
- Password: `admin`

## Using the CLI

1. Login with default credentials:
   ```bash
   ./share/cli/entitydb-cli login admin admin
   ```

2. Create an entity:
   ```bash
   ./share/cli/entitydb-cli entity create \
     --type=issue \
     --title="My First Issue" \
     --tags="priority:high,status:pending"
   ```

3. List entities:
   ```bash
   ./share/cli/entitydb-cli entity list --type=issue
   ```

4. Get a specific entity:
   ```bash
   ./share/cli/entitydb-cli entity get --id=entity_123
   ```

## Web Interface

Open your browser to http://localhost:8085 to access the web dashboard.

## Next Steps

- Read the [Architecture Overview](../architecture/overview.md)
- Learn about [Tag Namespaces](../architecture/tags.md)
- Explore the [API Reference](../api/entities.md)
- Check out [API Examples](../api/examples.md)