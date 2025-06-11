# Contributing to EntityDB

Thank you for your interest in contributing to the EntityDB (EntityDB) platform. This document provides guidelines and best practices for contributing to the project.

## Development Setup

### Prerequisites
- Go 1.18 or higher
- SQLite 3
- Git

### Setting Up the Development Environment
1. Clone the repository:
   ```bash
   git clone https://git.home.arpa/osakka/entitydb.git
   cd entitydb
   ```

2. Start the server:
   ```bash
   ./bin/entitydbd.sh start
   ```

3. Verify the server is running:
   ```bash
   ./bin/entitydbd.sh status
   ```

## Project Structure

- `/bin` - System binaries and scripts
- `/docs` - Documentation
- `/share` - Shared resources and test scripts
- `/src` - Source code
  - `/src/api` - API endpoints and handlers
  - `/src/models` - Data models and repository implementations
  - `/src/tools` - Utility tools and test implementations
- `/var` - Variable data (database, logs, PID files)

## Coding Standards

### Go Code Style
- Follow the [Go Code Review Comments](https://gitdataset.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format code before committing
- Add appropriate comments for public functions and packages
- Use meaningful variable and function names
- Keep functions small and focused on a single task

### Shell Scripts
- Include a shebang line (`#!/bin/bash`)
- Add a brief description of what the script does
- Use meaningful variable names
- Add error handling where appropriate

### SQL
- Use uppercase for SQL keywords
- Use snake_case for table and column names
- Include comments for complex queries
- Organize SQL files in the src/models/sqlite directory

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
- Use the test runner to run API tests:
  ```bash
  ./src/tools/run_api_tests.sh.improved --verbose
  ```

### Adding New Tests
- Add unit tests for new code in the appropriate test directory
- For API endpoints, add API tests in the share/tests/api directory
- Ensure your tests are included in the appropriate test runner

## Documentation

- Update the README.md file with any changes to the system's functionality
- Document new API endpoints
- Add inline documentation for complex code
- Update the STATE.md and WORKFLOW.md files as needed

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