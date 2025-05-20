# Renaming to EntityDB (EntityDB Platform)

This document outlines the rename from AIWM (AI Workforce Management Platform) to EntityDB (EntityDB Platform).

## Naming Evolution

1. Original: CCMF (Claude Code Management Framework)
2. First rename: AIWM (AI Workforce Management Platform)
3. Current rename: EntityDB (EntityDB Platform)

## Changes Required

### Script Renames
- `aiwmc.sh` → `entitydbc.sh` (Client wrapper)
- `aiwmd.sh` → `entitydbd.sh` (Server/daemon wrapper)

### Documentation Updates
- All references to AIWM updated to EntityDB in:
  - CLAUDE.md
  - agile.md
  - Any other documentation

### File Path Changes
- PID file: `/opt/ccmf/var/aiwm.pid` → `/opt/ccmf/var/entitydb.pid`
- Log file: `/opt/ccmf/var/log/aiwm.log` → `/opt/ccmf/var/log/entitydb.log`
- Database: `/opt/ccmf/var/db/aiwm.db` → `/opt/ccmf/var/db/entitydb.db`

### Implementation Notes
- The core binary itself may also need renaming from `aiwm` to `entitydb`
- The documentation updates have been prepared first, assuming the script renames will follow
- The web dashboard URL remains unchanged (http://localhost:8085)

## Updated Documentation

The updated documentation files have been prepared and are located at:
- `/tmp/claude-update/CLAUDE.md.entitydb` - Updated main documentation
- `/tmp/claude-update/agile.md.entitydb` - Updated agile methodology documentation

These files contain all necessary naming updates and are ready to be used once the script renames are completed.

## Recommended Next Steps

1. Rename the wrapper scripts:
   ```bash
   cp /opt/ccmf/aiwmc.sh /opt/ccmf/entitydbc.sh
   cp /opt/ccmf/aiwmd.sh /opt/ccmf/entitydbd.sh
   ```

2. Update script internals to reference the new names

3. Rename the core binary (if applicable):
   ```bash
   cp /opt/ccmf/aiwm /opt/ccmf/entitydb
   ```

4. Apply the updated documentation files

5. Update any other references to the old name in the codebase

Date: 2025-05-06