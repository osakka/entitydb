# Documentation Update for AIWM

This document provides proposed updates for the CLAUDE.md and agile.md files to reflect the new client-server system architecture using aiwmc.sh and aiwmd.sh wrapper scripts.

## Changes Made

1. **CLAUDE.md Updates**:
   - Replaced all occurrences of `ccmf_client` with `aiwmc.sh`
   - Replaced references to `ccmf` server control with `aiwmd.sh`
   - Updated directory structure to reflect the simplified system
   - Updated file paths for PID, logs, and database
   - Updated server architecture sections to reflect new component names
   - Simplified client component section to reference wrapper script
   - Updated web dashboard URL to http://localhost:8085
   - Updated all command examples to use the new wrapper scripts
   - Streamlined documentation to match the simplified system design

2. **agile.md Updates**:
   - Replaced all occurrences of `CLAUDE/scripts/` script references with `aiwmc.sh` commands
   - Updated document title to reference AI Workforce Management Platform (AIWM)
   - Updated session management examples with aiwmc.sh commands
   - Replaced project monitoring script references with dashboard access
   - Updated git hooks installation instructions
   - Updated work item verification references
   - Updated task management workflow to use API commands
   - Updated dashboard access and monitoring sections
   - Added a timestamp for the update

## Implementation Notes

The updated files have been prepared in the following locations:
- `/tmp/claude-update/CLAUDE.md.updated`
- `/tmp/claude-update/agile.md.updated`

These files are ready to be used as replacements for the current documentation files once approved.

## Permission Issues

Note that direct updates to the existing documentation files were not possible due to permission restrictions. The files would need to be updated by a user with write permissions to these files.

## Recommended Next Steps

1. Review the updated documentation files
2. Apply the changes to the production documentation files (requires appropriate permissions)
3. Consider additional documentation updates for any other files referencing the old system structure
4. Update any remaining scripts that might reference ccmf_client command

Date: 2025-05-06