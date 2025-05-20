# EntityDB Project Structure

The EntityDB project follows a clean, organized structure:

```
/opt/entitydb/
├── bin/                    # Core binaries only
│   ├── entitydb           # Main server binary
│   └── entitydbd.sh       # Server daemon control script
│
├── share/                  # Shared resources
│   ├── cli/               # Command-line tools
│   │   ├── entitydb-cli   # Main CLI interface
│   │   └── entitydb-api.sh
│   │
│   ├── tests/             # Test scripts
│   │   ├── api/          # API tests
│   │   └── entity/       # Entity tests
│   │
│   ├── utilities/         # Utility programs
│   │   ├── generate_hash.go
│   │   └── ...           # Other utilities
│   │
│   ├── tools/            # Operation tools
│   │   └── ...           # Setup and migration scripts
│   │
│   └── htdocs/           # Web UI files
│
├── src/                   # Source code
│   ├── api/              # API handlers
│   ├── models/           # Data models
│   └── storage/          # Storage implementations
│
├── docs/                  # Documentation
└── var/                   # Runtime data
    ├── entities.ebf      # Binary database
    └── entitydb.wal      # Write-ahead log
```

## Key Principles

1. **Binary Directory**: Contains only essential binaries (server and daemon script)
2. **Share Directory**: Contains all shared resources organized by type
3. **Source Directory**: Clean source code without test files or utilities mixed in
4. **Documentation**: Comprehensive documentation in `/docs`
5. **Runtime Data**: All runtime data in `/var`