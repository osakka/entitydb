# EntityDB v2.13.0 Release Notes

## Release Date: 2025-05-19

## 🚀 Major Features

### SSL/TLS Support
- Added full SSL/TLS support on all ports
- Configurable via environment variables
- Proper certificate management
- Seamless HTTP/HTTPS switching
- Simple configuration via `ENTITYDB_USE_SSL` flag

### Configuration System Overhaul
- Complete environment-based configuration implementation
- All hardcoded values eliminated from source code
- Clear configuration hierarchy with multiple levels
- Comprehensive documentation for all settings

### Configuration Files
- **Default Configuration**: `share/config/entitydb_server.env`
  - Contains all available settings with sensible defaults
  - Well-documented with descriptions for each variable
  - Reference for all configuration options
  
- **Instance Configuration**: `var/entitydb.env`
  - Override specific settings per instance
  - Only include values that differ from defaults
  - Keeps sensitive settings separate

### Configuration Hierarchy
1. Command Line Flags (highest precedence)
2. Environment Variables
3. Instance Configuration (`var/entitydb.env`)
4. Default Configuration (`share/config/entitydb_server.env`)
5. Hardcoded Defaults (lowest precedence)

### Content Encoding Improvements
- Fixed double-encoding issues in entity content
- Added proper MIME type detection and tagging system
- Standardized content handling across system
- Auto-detection of content types (string vs JSON)
- Seamless handling of both text and binary content
- Proper base64 encoding with consistent format
- Enhanced debug logging for entity operations
- Fixed entity persistence issues
- Backward compatibility with existing entities

## 🔧 Technical Improvements

- Removed unused `--config` flag that was never implemented
- Updated startup script to automatically source configuration files
- Added helper functions for environment variable handling
- Improved configuration loading order and precedence
- Better separation of concerns for configuration management

## 📦 Project Structure Cleanup

- Moved temporary scripts to `tmp/` directory
- Reorganized configuration files to `share/config/`
- Updated all documentation to reflect new structure
- Cleaned up project root directory

## 🔄 Migration Guide

### From v2.12.0
1. Configuration is backward compatible - no immediate action required
2. To use new system:
   - Copy needed values from command line to `var/entitydb.env`
   - Remove command line flags from startup scripts
   - Let environment variables handle configuration

### Environment Variables
All configuration now available as environment variables:
- `ENTITYDB_PORT` - HTTP server port (default: 8085)
- `ENTITYDB_SSL_PORT` - HTTPS server port (default: 8443)
- `ENTITYDB_USE_SSL` - Enable SSL/TLS (default: false)
- `ENTITYDB_DATA_PATH` - Data storage directory
- `ENTITYDB_LOG_LEVEL` - Logging level
- And more...

## 📊 Default Changes

- SSL disabled by default for development (`ENTITYDB_USE_SSL=false`)
- Default HTTP port set to 8085
- Default HTTPS port set to 8443
- All ports consistently documented as 8085/8443

## 🐛 Bug Fixes

- Fixed SSL certificate verification to only run when SSL is enabled
- Corrected port checking in startup script
- Fixed environment variable export in daemon script
- Resolved configuration loading order issues
- Fixed critical content encoding issues with binary data
- Resolved JSON content double-encoding problem
- Fixed content type detection and preservation
- Corrected content storage format issues
- Fixed inconsistencies in base64 encoding/decoding
- Addressed data corruption risks with binary content

## 📚 Documentation

- New `docs/CONFIG_SYSTEM.md` with comprehensive configuration guide
- Updated `README.md` with new configuration instructions
- Added migration notes for existing installations
- Documented all available environment variables

## 🙏 Acknowledgments

This release continues the collaboration with Claude AI, implementing a clean configuration methodology that eliminates hardcoded values and provides flexible deployment options.

---

For questions or issues, please visit:
https://github.com/osakka/entitydb