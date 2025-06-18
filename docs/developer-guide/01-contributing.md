# Contributing to EntityDB

Thank you for your interest in contributing to the EntityDB (EntityDB) platform. This document provides guidelines and best practices for contributing to the project.

## Development Setup

### Prerequisites
- Go 1.23 or higher (current: Go 1.23.0 with toolchain 1.24.2)
- Git

### Setting Up the Development Environment
1. Clone the repository:
   ```bash
   git clone https://git.home.arpa/itdlabs/entitydb.git
   cd entitydb
   ```

2. Build and start the server:
   ```bash
   cd src
   make
   make install
   ../bin/entitydbd.sh start
   ```

3. Verify the server is running:
   ```bash
   ./bin/entitydbd.sh status
   ```

## Project Structure

- `/bin/` - System binaries and daemon scripts
- `/docs/` - Comprehensive documentation library
- `/share/` - Static web assets, configuration templates, and test cases
- `/src/` - Go source code
  - `/src/api/` - HTTP handlers and middleware
  - `/src/models/` - Entity models and business logic
  - `/src/storage/` - Binary storage implementation with WAL
  - `/src/tools/` - Administrative and maintenance utilities
- `/var/` - Runtime data (binary database, WAL, logs, PID files)

## Coding Standards

### Go Code Style
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format code before committing
- Add appropriate comments for public functions and packages
- Use meaningful variable and function names
- Keep functions small and focused on a single task

### Shell Scripts
- Include a shebang line (`#!/bin/bash`)
- Add a brief description of what the script does
- Use meaningful variable names
- Add error handling where appropriate


## Git Workflow

### Branches
- `main` - Stable release branch
- Feature branches should be named with a descriptive prefix: `feature/feature-name`
- Bug fix branches should be named with the prefix: `fix/bug-name`

### Commit Messages
- Use present tense ("Add feature" not "Added feature")
- First line should be a summary (50 chars or less)
- Include a more detailed description if necessary
- Reference issue numbers if applicable

### Pull Requests
1. Create a new branch from `main`
2. Make your changes with appropriate commits
3. Push your branch to the repository
4. Create a pull request back to `main`
5. Include a clear description of the changes and any related issues

### Code Review Process
- All code must be reviewed before merging
- Address all comments and feedback
- All tests must pass

## Testing

### Running Tests
- Run comprehensive test suite:
  ```bash
  cd src
  make test
  ```
- Run specific test categories:
  ```bash
  # API integration tests
  ../tests/run_all_tests.sh
  
  # Performance tests
  ../tests/performance/test_performance.sh
  
  # Temporal functionality tests
  ../tests/temporal/test_temporal.sh
  ```

### Adding New Tests
- **Unit Tests**: Add Go unit tests alongside source files with `_test.go` suffix
- **Integration Tests**: Add shell-based tests in `/tests/` directory structure
- **API Tests**: Add test cases to `/share/tests/cases/` with `.test` extension
- **Performance Tests**: Add benchmarks to `/tests/performance/` directory
- Follow existing patterns in test directory structure

## Documentation

- Update documentation in `/docs/` following the professional taxonomy
- Document new API endpoints in `/docs/api-reference/`
- Add inline Go documentation for public functions and types
- Update `/docs/developer-guide/` for development-related changes
- Follow documentation standards in [Documentation Architecture](./09-documentation-architecture.md)

## Issue Reporting

When reporting issues, please include:
- A clear description of the issue
- Steps to reproduce
- Expected behavior
- Actual behavior
- EntityDB version or commit hash
- Any relevant logs or error messages

## License

By contributing to EntityDB, you agree that your contributions will be licensed under the project's license.

---

Thank you for contributing to the EntityDB platform!