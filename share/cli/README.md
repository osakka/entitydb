# EntityDB Command Line Interface

This directory contains tools and documentation for interacting with the EntityDB (EntityDB) API via command line interfaces.

## Contents

- **README.md**: This file - comprehensive API reference
- **entitydb-api.sh**: Bash-based CLI tool for EntityDB API
- **entitydb_client.py**: Python client library for EntityDB API
- **example.py**: Example of using the Python client programmatically
- **test_api.sh**: Test script to verify EntityDB API functionality

## Getting Started

The EntityDB API uses JWT authentication for all requests. You'll need to authenticate first to get a token before making other API calls.

### Using the Bash CLI (entitydb-api.sh)

```bash
# Make the script executable
chmod +x entitydb-api.sh

# Login to get a token
./entitydb-api.sh login admin password

# List entities
./entitydb-api.sh entity list

# Get help
./entitydb-api.sh help
```

### Using the Python Client (entitydb_client.py)

```bash
# Make the script executable
chmod +x entitydb_client.py

# Login and list entities
./entitydb_client.py login admin password
./entitydb_client.py entity list

# Get help
./entitydb_client.py -h
```

### Run the Example

```bash
# Make the example executable
chmod +x example.py

# Run the example
./example.py
```

### Run API Tests

```bash
# Make the test script executable
chmod +x test_api.sh

# Run tests
./test_api.sh
```

## API Structure

The EntityDB API follows a pure entity-based architecture with these main endpoints:

- `/api/v1/entities`: Manage entities (create, read, update, delete)
- `/api/v1/entity-relationships`: Manage relationships between entities
- `/api/v1/auth/login`: Authentication endpoint

For detailed API reference, see the API reference section in this document.

## Integration with Other Tools

You can use the EntityDB API with standard tools like `curl` and `jq`:

```bash
# Login and get a token
TOKEN=$(curl -s -X POST "http://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.data.token')

# Use the token for API requests
curl -X GET "http://localhost:8085/api/v1/entities" \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Development

To extend the CLI tools or add new functionality:

1. The Bash CLI (`entitydb-api.sh`) is designed to be easily extended with new commands
2. The Python client (`entitydb_client.py`) can be imported as a module in other Python applications
3. New API endpoints can be added by updating the respective functions in each tool

## Troubleshooting

If you encounter issues:

1. Ensure the EntityDB server is running on the expected port (default: 8085)
2. Check your authentication credentials
3. Use the `--debug` flag with the Bash CLI or `--format=json` with the Python client for more detailed output
4. Run the `test_api.sh` script to verify API functionality