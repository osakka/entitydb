# EntityDB v2.12.0 Release Notes

## Release Date: 2025-05-19

## 🚀 Major Features

### Autochunking Implementation
- Automatic chunking for files larger than 4MB
- Support for unlimited file sizes (only limited by filesystem)
- Memory-efficient streaming - never loads full files into RAM
- SHA256 checksums for data integrity

### Simplified Entity Model
- Replaced complex multi-content array with single byte array
- Clean architectural approach with no backward compatibility
- Improved performance and reduced complexity

### Database Location Standardization
- All data now consistently stored in `/opt/entitydb/var/`
- Single directory for all database files
- No more scattered data locations

## 🔧 Technical Improvements

- Fixed authentication double-encoding issue
- Improved error handling for large files
- Enhanced repository interface implementations
- Better temporal data handling

## 📦 New Test Suite

- `test_entity_autochunking.sh` - Comprehensive chunking tests
- `test_simple_entity.sh` - Basic entity operations
- `test_entity_json_content.sh` - JSON content handling
- `test_data_inspection.sh` - Database inspection tools
- Multiple authentication and workflow tests

## 💥 Breaking Changes

- Entity model completely replaced - no migration path
- Previous content format incompatible
- Database format changed
- Clean installation required

## 🔄 Migration Guide

1. Stop existing EntityDB server
2. Backup existing data if needed
3. Install v2.12.0
4. Start fresh with new database

## 📊 Performance

- 4MB default chunk size
- Streaming support for large files
- Reduced memory footprint
- Improved write performance

## 🐛 Bug Fixes

- Fixed content double encoding issue - content bytes were being base64 encoded twice in API responses
- Fixed login authentication issues
- Resolved content encoding problems
- Corrected repository interface mismatches
- Fixed temporal repository implementations

## 📚 Documentation

- Comprehensive autochunking documentation
- Entity model migration guide
- Updated API documentation
- New test suite documentation

## 🙏 Acknowledgments

This release was developed with assistance from Claude AI, implementing a clean architectural approach to handle large files efficiently.

---

For questions or issues, please visit:
https://github.com/osakka/entitydb