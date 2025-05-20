# EntityDB Source Code

This directory contains the source code for the EntityDB platform. This document provides an overview of the source code structure and development guidelines.

## Directory Structure

```
/src/
├── api/                      # API handlers and routing
│   ├── entity_handler.go     # Entity API implementation
│   ├── entity_handler_rbac.go # RBAC wrapper for entity API
│   ├── entity_relationship_handler.go # Relationship API
│   ├── relationship_handler_rbac.go # RBAC wrapper for relationships
│   ├── user_handler.go       # User management API
│   ├── user_handler_rbac.go  # RBAC wrapper for user API
│   ├── auth_middleware.go    # Authentication middleware
│   ├── rbac_middleware.go    # RBAC enforcement middleware
│   ├── router.go             # HTTP router setup
│   └── response_helpers.go   # Response formatting utilities
├── models/                   # Data models
│   ├── entity.go             # Core entity model
│   ├── entity_relationship.go # Relationship model
│   ├── entity_query.go       # Query builder pattern
│   ├── session.go            # Session management
│   ├── tag_namespace.go      # Tag namespace utilities
│   └── errors.go             # Error definitions
├── storage/                  # Storage implementations
│   └── binary/               # Binary format storage
│       ├── entity_repository.go      # Entity storage
│       ├── relationship_repository.go # Relationship storage
│       ├── temporal_repository.go    # Temporal features
│       ├── high_performance_repository.go # Optimized implementation
│       ├── writer.go                 # Binary format writing
│       ├── reader.go                 # Binary format reading
│       ├── format.go                 # Binary format specification
│       ├── wal.go                    # Write-ahead logging
│       └── mmap_reader.go            # Memory-mapped access
├── cache/                    # Caching implementations
│   └── query_cache.go        # Query result caching
├── logger/                   # Logging system
│   └── logger.go             # Structured logging
├── tools/                    # Development tools
│   ├── fix_index.go          # Index repair utility
│   └── check_admin_user.go   # Admin user verification
├── main.go                   # Server entry point
└── Makefile                  # Build system
```

## Code Organization Principles

1. **Package Structure**
   - Each package should have a single, well-defined responsibility
   - Avoid circular dependencies between packages
   - Prefer shallow hierarchies over deep nesting

2. **File Naming**
   - Use lowercase with underscores for filenames
   - Group related functionality in the same file
   - Use descriptive, specific names

3. **Interface-Based Design**
   - Define interfaces before implementations
   - Program to interfaces, not concrete types
   - Keep interfaces focused and minimal

## Git Workflow and Protocol

### Branch Strategy

We use a trunk-based development model:

1. **Main Branch (`main`)**
   - Always deployable
   - Protected from direct pushes
   - All features, fixes, and changes go through PRs

2. **Feature Branches**
   - Short-lived branches for specific features or fixes
   - Format: `feature/short-description` or `fix/issue-description`
   - Rebase frequently with main
   - Delete after merging

### Commit Guidelines

1. **Commit Frequency**
   - Commit FREQUENTLY (multiple times per day)
   - Small, focused commits are better than large, monolithic ones
   - Each commit should compile and pass tests

2. **Commit Message Format**
   ```
   type: Short summary (50 chars or less)

   Detailed explanation if necessary. Wrap at 72 characters.
   Explain what and why, not how (the code shows that).

   Fixes #123
   ```

3. **Commit Types**
   - `feat:` - New features
   - `fix:` - Bug fixes
   - `docs:` - Documentation changes
   - `style:` - Formatting, missing semicolons, etc (no code change)
   - `refactor:` - Code refactoring (no feature or bug fix)
   - `perf:` - Performance improvements
   - `test:` - Adding or fixing tests
   - `chore:` - Build process or auxiliary tool changes

4. **Sign Your Commits**
   - All commits must be signed by default
   - Use `git config --global commit.gpgsign true`
   - Include the AI co-author line when generated with Claude:
     ```
     🤖 Generated with Claude Code

     Co-Authored-By: Claude <noreply@anthropic.com>
     ```

### Pull Request Protocol

1. **Before Creating a PR**
   - Ensure all tests pass
   - Rebase on latest main
   - No unfinished work in the PR

2. **PR Description Template**
   ```
   ## Summary
   Brief explanation of the changes

   ## Test Plan
   How you tested the changes

   ## Related Issues
   Fixes #123
   ```

3. **PR Review Process**
   - PRs require at least one review before merging
   - Address ALL feedback before merging
   - Use the "request changes" feature for blocking issues
   - Respond to all comments

4. **After PR Approval**
   - Squash commits if necessary for a clean history
   - Merge using "Merge commit" (not squash merge or rebase)
   - Delete the branch after merging

### Git Hygiene Rules

1. **NEVER** rewrite public history (no force push to main)
2. **NEVER** commit directly to main branch
3. **NEVER** commit temporary or debug code
4. **NEVER** commit large binary files (use Git LFS if necessary)
5. **NEVER** commit sensitive information (tokens, passwords, keys)
6. **ALWAYS** verify what you're committing before pushing
7. **ALWAYS** keep commits focused on a single logical change
8. **ALWAYS** write meaningful commit messages
9. **ALWAYS** keep your local copy updated with remote

### Git Repository Configuration

- **Repository URL**: https://git.home.arpa/itdlabs/entitydb.git
- **Credentials Management**: Use Git credential helper

## Code Quality Standards

1. **Testing Requirements**
   - All new code must have tests
   - Unit tests for core functionality
   - Integration tests for API endpoints
   - Maintain >80% test coverage

2. **Code Style**
   - Follow Go style conventions (gofmt)
   - Use consistent naming conventions
   - Document all exported functions, types, and constants
   - Keep functions small and focused

3. **Code Review Checklist**
   - Does the code follow our architecture principles?
   - Is the code well-tested?
   - Is there proper error handling?
   - Is the code efficient? Any performance concerns?
   - Is the code secure? Any vulnerability concerns?
   - Is the code maintainable? Clear and readable?

## Build and Test

```bash
# Build the server
make

# Run all tests
make test

# Run only unit tests
make unit-tests

# Run only API tests
make api-tests

# Build and install
make install
```

## Architecture Policy

1. **Pure Entity Model**: All data must be stored as entities with tags
2. **No Direct Database Access**: All operations through the API
3. **Proper Authentication**: All endpoints must use auth middleware
4. **RBAC Enforcement**: All operations must check permissions
5. **Clean Architecture**: Clear separation of concerns