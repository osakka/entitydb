# EntityDB Scripts

This directory contains utility scripts for EntityDB operations.

## Available Scripts

### set_log_level.sh

Sets the EntityDB server log level for debugging purposes.

```bash
./set_log_level.sh debug  # Enable debug logging
./set_log_level.sh info   # Normal logging (default)
./set_log_level.sh warn   # Warnings only
./set_log_level.sh error  # Errors only
```

**Note**: This script starts the server directly with the specified log level,
bypassing the daemon script. Use `entitydbd.sh` for normal operations.

## Removed Scripts

The following scripts have been removed or reorganized:

- `enable_turbo.sh` - Removed (outdated, high-performance mode is now default)
- `entitydbd-info.sh` - Removed (use `set_log_level.sh` instead)
- `repair-index.sh` - Moved to `share/tools/repair_index.sh`

## Other Tools

See the `share/tools/` directory for additional utilities:
- `setup_ssl.sh` - SSL certificate setup
- `repair_index.sh` - Database index repair tool