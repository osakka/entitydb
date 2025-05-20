# Repository Cleanup Plan

## Current Repository Status

Based on our analysis, the repository has several areas that need cleanup:

1. A remote branch `fix_makefile_and_tests` that contains deletions of our entity implementation (now deleted locally)
2. Several untracked files related to RBAC, feature flags, and schema updates
3. A deprecated directory with legacy code
4. Various backup and temporary files (.bak, .new) scattered throughout the repo
5. Some database backup files that shouldn't be in version control

## Cleanup Tasks

### 1. Branch Cleanup

- [x] Delete local `fix_makefile_and_tests` branch (already completed)
- [ ] Remove remote `fix_makefile_and_tests` branch:
  ```bash
  git push origin --delete fix_makefile_and_tests
  ```

### 2. Add Relevant Untracked Files

Several important files are currently untracked. We should review and add them to git:

- [ ] RBAC implementation files:
  ```bash
  git add src/api/fix_rbac_tests.go
  git add src/api/rbac_absolute_mock.go
  git add src/api/rbac_mock_handler.go
  git add src/api/rbac_test_handler.go
  ```

- [ ] Feature flag implementation:
  ```bash
  git add src/api/feature_flags_handler.go
  ```

- [ ] Schema update files (after review):
  ```bash
  # Add after reviewing which schema files are needed
  git add src/models/sqlite/schema_cleanup.sql
  git add src/models/sqlite/schema_fix_issue_history.sql
  git add src/models/sqlite/schema_update_test_compat.sql
  ```

### 3. Clean Up Backup and Temporary Files

- [x] Remove `src/tools/README.md.bak` (already completed)
- [ ] Remove or add to .gitignore other backup files:
  ```bash
  # Remove from git but keep locally (if needed)
  git rm --cached src/models/sqlite/schema.sql.bak.20250507205059
  git rm --cached var/db/entitydb.db.bak.20250507205059
  ```

- [ ] Add pattern to .gitignore to prevent future backup files:
  ```
  # Add to .gitignore
  *.bak
  *.tmp
  *.new
  *.old
  ```

### 4. Database Files Cleanup

- [ ] Remove database backup files from git:
  ```bash
  git rm --cached var/db/backups/entitydb_db_backup_*.db
  ```

- [ ] Add database backup pattern to .gitignore:
  ```
  # Add to .gitignore
  var/db/backups/
  var/db/*.bak
  var/db/*.backup
  ```

### 5. Review Deprecated Directory

- [ ] Review the deprecated directory to ensure no valuable code is lost
- [ ] Consider creating an archive branch before removing:
  ```bash
  git checkout -b archive/deprecated
  git add -A
  git commit -m "Archive full repository state before cleaning deprecated code"
  git push origin archive/deprecated
  git checkout main
  ```

- [ ] Remove deprecated directory from active development if no longer needed:
  ```bash
  git rm -r deprecated/
  git rm -r src/deprecated/
  git rm -r src/models/sqlite/deprecated/
  ```

### 6. Organize Untracked SQL Files

- [ ] Review all SQL files and categorize them:
  - Schema creation files
  - Schema migration files
  - Schema fix files
  - Permission setup files

- [ ] Create an organized structure:
  ```
  src/models/sqlite/
  ├── schema/
  │   └── base.sql             # Base schema
  ├── migrations/
  │   ├── 001_initial.sql
  │   ├── 002_entity.sql
  │   └── 003_relationships.sql
  ├── fixes/
  │   ├── fix_issue_history.sql
  │   └── fix_tags.sql
  └── setup/
      ├── create_admin.sql
      └── grant_permissions.sql
  ```

- [ ] Move files to appropriate directories

### 7. Create a .gitattributes File

- [ ] Create a .gitattributes file to ensure consistent line endings and handling of binary files:
  ```
  # Set default behavior to automatically normalize line endings
  * text=auto

  # Explicitly declare text files
  *.go text
  *.js text
  *.html text
  *.css text
  *.md text
  *.sql text
  *.sh text eol=lf

  # Declare binary files
  *.db binary
  *.png binary
  *.jpg binary
  *.gif binary
  *.ico binary
  ```

## Implementation Sequence

1. Create archive branch as a safety measure
2. Clean up branches (local and remote)
3. Add relevant untracked files
4. Remove backup and temporary files
5. Update .gitignore and .gitattributes
6. Organize SQL files into a better structure
7. Review and decide on deprecated directory
8. Commit all changes with clear commit messages
9. Push to remote repository

## Expected Outcome

After implementing these changes:

1. The repository will contain only relevant and current code
2. All important files will be tracked in git
3. Backup and temporary files will be excluded via .gitignore
4. SQL files will be organized in a logical structure
5. No unnecessary branches will exist
6. The repository will be clean, organized, and maintainable

This cleanup will make the codebase more approachable for developers, reduce confusion, and improve maintainability of the project in the long term.