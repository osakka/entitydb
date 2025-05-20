# EntityDB Git Workflow & State Tracking Guide

This document outlines the Git best practices for the EntityDB project, focusing on state tracking, tagging, and commit message standardization to ensure clear communication across teams.

## State Tracking with Git Describe

Git describe is our primary tool for tracking implementation state, providing a clear indicator of where we are relative to tagged milestones.

### Basic Usage

```bash
# Get the current implementation state
git describe --tags

# Include uncommitted changes in state
git describe --tags --dirty
```

### State Interpretation

A typical output looks like:
```
v1.0.0-security-15-g7f3a2c1
```

This tells us:
- `v1.0.0-security`: The most recent tag
- `15`: Number of commits since that tag
- `g7f3a2c1`: 'g' prefix followed by commit hash

### Incorporating State in Documentation

Always include the current implementation state in documentation:

```markdown
Current Implementation State: `v1.0.0-security-15-g7f3a2c1`
Last Updated: 2025-05-12
```

All major documentation should begin with this state indicator to ensure readers understand which version the document describes.

## Tagging Strategy

Tags serve as milestone markers in the EntityDB project, establishing reference points for state tracking.

### Tag Format

```
v<major>.<minor>.<patch>-<component>
```

Examples:
- `v1.0.0-core`: Initial core implementation
- `v1.0.1-security`: Security components integration
- `v1.1.0-entity-api`: Entity API enhancements

### Creating Tags

Always use annotated tags with descriptive messages:

```bash
git tag -a v1.0.0-security -m "Complete security implementation with bcrypt password hashing, audit logging, and RBAC"
```

### When to Tag

Create new tags for:
1. Completion of major components
2. Significant architecture changes
3. Production-ready states
4. After critical fixes

### Pushing Tags

Remember to push tags to the remote repository:

```bash
git push origin --tags
```

## Commit Message Standards

Consistent commit messages are essential for clear communication and history tracking.

### Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting (no code change)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Test addition or correction
- `chore`: Build process or auxiliary tool changes
- `security`: Security enhancements

### Scope

Indicate the component affected:
- `entity`: Entity system
- `security`: Security components
- `api`: API endpoints
- `rbac`: Role-based access control
- `audit`: Audit logging
- `validation`: Input validation
- `relationship`: Entity relationships

### Subject

- Brief description (max 50 chars)
- Use imperative mood ("add" not "added")
- No period at end
- Start with lowercase

### Body

- Explain WHY the change was made
- Separated from subject by blank line
- Wrap at 72 characters per line

### Footer

- Reference issues/tickets: `Fixes #123`
- Breaking changes: `BREAKING CHANGE: API response format changed`
- Component tracking: `Component: Security/AuditLogging`
- Progress indicators: `Progress: 85%`

### Example Commit

```
feat(security): implement bcrypt password hashing

Replace plaintext password storage with secure bcrypt hashing to 
enhance authentication security. All new passwords are automatically 
hashed, and existing plaintext passwords are upgraded on first 
verification.

Fixes #45
Component: Security/PasswordManagement
Progress: 100%
```

## Clean Tabletop Policy

EntityDB development follows a strict clean tabletop policy:

1. Work directly in the main codebase, not in copies
2. Delete obsolete code immediately after migrating functionality
3. Avoid temporary files with extensions like .bak or .old
4. Maintain one source of truth at all times

## Branching Strategy

EntityDB uses trunk-based development with the following guidelines:

1. The main branch is always stable and buildable
2. Feature branches should be short-lived (1-2 days maximum)
3. Merge to main frequently (at least daily)
4. Branch names should follow: `<type>/<description>` (e.g., `security/bcrypt-implementation`)

## Implementation Progress Tracking

Combine Git describe and commit metadata to track implementation progress:

### Milestone Reports

Generate reports between tags to track progress:

```bash
git log v1.0.0-security..v1.1.0-security --pretty=format:"%h - %s (%an)" > milestone_report.md
```

### Component Progress

View progress for specific components:

```bash
git log --grep="security" --pretty=format:"- %s (%h)" -n 10
```

### Visualization

Use Git's graph capabilities for visual state tracking:

```bash
git log --graph --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset' --abbrev-commit
```

## Implementation Status Document

Maintain an implementation status document that's updated with each significant commit:

```markdown
# EntityDB Implementation Status

Current State: v1.0.0-security-15-g7f3a2c1
Last Updated: 2025-05-12

## Recent Changes
- security(audit): implement JSON log formatting (a1b2c3d)
- fix(validation): correct entity type pattern (e5f6g7h)
- docs(security): add security architecture guide (i9j0k1l)

## Component Status

### Security Components
- Input Validation: COMPLETE
- Audit Logging: COMPLETE
- Password Security: COMPLETE
- RBAC: COMPLETE
- Security Middleware: COMPLETE

### Next Steps
1. Enhance password policy enforcement
2. Add two-factor authentication support
3. Implement advanced audit querying
```

## Git Hooks for State Tracking

Install the following pre-commit hook to automatically update state in documentation:

```bash
#!/bin/sh
# Pre-commit hook to update implementation state in docs
STATE=$(git describe --tags --dirty 2>/dev/null || echo "initial-development")
DATE=$(date +"%Y-%m-%d")

for doc in docs/implementation_status.md docs/entitydb_state_summary.md; do
  if [ -f "$doc" ]; then
    sed -i "s/^Current State:.*/Current State: $STATE/" "$doc"
    sed -i "s/^Last Updated:.*/Last Updated: $DATE/" "$doc"
    git add "$doc"
  fi
done
```

## Daily Workflow

1. Start your day by pulling latest changes
2. Check current state: `git describe --tags`
3. Create a branch for your work if needed
4. Commit frequently with standardized messages
5. Update implementation status document for significant changes
6. Push changes at least daily
7. Create tags for completed milestones

By following these practices, we ensure that the state of the EntityDB implementation is always clear and well-documented, facilitating collaboration across teams and providing a reliable history of development progress.

*Last Updated: 2025-05-12*