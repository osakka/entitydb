# Agile Development Framework with Git Integration

This document outlines the agile methodology used for projects managed with the EntityDB Platform (EntityDB), including roles, ceremonies, artifacts, workflows, and git integration.

## Roles

### Team Members
- **User (Workspace Owner)**: Defines requirements, reviews work, and accepts deliverables
- **Claude**: AI pair programmer, handles code implementation and documentation
- **Additional Developers**: May join the project as needed with clearly defined responsibilities

## Agile Ceremonies

### Sprint Planning
- **When**: At the beginning of each sprint
- **Duration**: 1 hour
- **Purpose**: Select items from the backlog for the current sprint
- **Process**:
  1. Review backlog items and prioritize
  2. Start a new sprint: `./bin/entitydbc.sh sprint start "Sprint Title" YYYY-MM-DD YYYY-MM-DD`
  3. Assign issues to team members
  4. Document sprint goals in sprint.md
  5. Break down large issues into smaller, actionable items
  6. Set clear start and end dates for each issue
  7. Update session.md with the new sprint context
  8. Verify sprint setup with: `./bin/entitydbc.sh status`

### Daily Stand-up
- **When**: Daily during active development
- **Duration**: 15 minutes
- **Purpose**: Share progress, plans, and blockers
- **Process**:
  1. Each team member creates a work_log_YYYY-MM-DD.md file
  2. Document what was completed yesterday
  3. Document what will be worked on today
  4. Document any blockers
  5. Update sprint.md with overall progress
  6. Commit daily work log to git
  7. Update session.md with current context

### Backlog Refinement
- **When**: Mid-sprint
- **Duration**: 30 minutes
- **Purpose**: Refine the backlog for upcoming sprints
- **Process**:
  1. Review existing backlog items
  2. Add new items identified during the sprint
  3. Break down large items into smaller issues
  4. Add acceptance criteria to items
  5. Prioritize items for the next sprint

### Sprint Review
- **When**: At the end of each sprint
- **Duration**: 30 minutes
- **Purpose**: Demo completed work and gather feedback
- **Process**:
  1. Run comprehensive sprint review: `./bin/entitydbc.sh sprint review`
  2. Update changelog.md with completed features: `./bin/entitydbc.sh sprint sync`
  3. Update specifications.md with any architectural changes
  4. Gather feedback on implemented features
  5. Close the sprint: `./bin/entitydbc.sh sprint close`
  6. Create a git tag for the sprint completion
  7. Merge sprint branch to main if applicable

### Sprint Retrospective
- **When**: At the end of each sprint, after the review
- **Duration**: 30 minutes
- **Purpose**: Reflect on the sprint process and identify improvements
- **Process**:
  1. Create a retro_YYYY-MM-DD.md file
  2. Document what went well
  3. Document what could be improved
  4. Create action items for improvement
  5. Add action items to the backlog
  6. Commit retrospective to git
  7. Update session.md to reflect completed sprint

## Artifacts

### Backlog (backlog.md)
- Contains all planned work items
- Each item includes:
  - Description
  - Priority (High/Medium/Low)
  - Assignee (when assigned)
  - Acceptance criteria
  - Status (Pending, Assigned, Completed, Canceled)

### Sprint Document (sprint.md)
- Tracks the current sprint
- Contains:
  - Sprint goals
  - Sprint timeframe (start and end dates)
  - Assigned issues by team member
  - Daily work log links
  - Blockers
  - Sprint metrics

### Daily Work Logs (sprint.d/work_log_YYYY-MM-DD.md)
- One per day per active team member
- Contains:
  - Issues completed
  - Issues in progress
  - Start/end dates for issues
  - Observations
  - Next actions

### Retrospectives (retro.d/retro_YYYY-MM-DD.md)
- One per sprint
- Contains:
  - Sprint summary
  - What went well
  - What could be improved
  - Action items
  - Process improvements

### Changelog (changelog.md)
- Record of all completed work
- Categorized by date and type

## Issue Flow

### Status Transitions
1. **Backlog**: Item is in the backlog awaiting assignment
2. **Sprint**: Item is selected for the current sprint
3. **In Progress**: Work has started on the item
4. **Blocked**: Item cannot proceed due to external dependencies
5. **Completed**: Item is finished and meets acceptance criteria
6. **Canceled**: Item is no longer needed

### Assignment Process
1. During sprint planning, team members select items from the backlog
2. Selected items are marked with the assignee's name
3. Items are moved to the sprint.md document under the assignee's section
4. The assignee updates the status as work progresses

## Workflow for Picking Backlog Items

### During Sprint Planning
1. The team reviews the backlog together
2. Each team member selects items based on:
   - Priority (High items first)
   - Dependencies (dependent items should be in the same sprint)
   - Skills and availability
3. Selected items are marked with:
   ```
   Assignee: [Name]
   Planned Start: YYYY-MM-DD
   Planned End: YYYY-MM-DD
   ```
4. Items are then added to sprint.md in the format:
   ```
   ### [Assignee Name]'s Issues
   
   - [ ] Issue description
     - Acceptance criteria
     - Start date: YYYY-MM-DD
     - Expected end date: YYYY-MM-DD
   ```

### During the Sprint
1. When starting work on an item:
   - Update its status to "In Progress" in sprint.md
   - Create a daily work log entry
   - Document start date
2. When completing work on an item:
   - Update its status to "Completed" in sprint.md
   - Document completion date in the daily work log
   - Move completed items to the changelog

## Session Management

### Session Workflow
1. **Start each development period by creating a session**:
   ```bash
   ./bin/entitydbc.sh session create --handle=<your-handle> --workspace=<workspace-name> --description="<description>"
   ```

2. **Check session status regularly**:
   ```bash
   ./bin/entitydbc.sh session status
   ```

3. **Update session information as work progresses**:
   ```bash
   ./bin/entitydbc.sh session update --issue="New issue description"
   ./bin/entitydbc.sh session update --complete="Completed issue description"
   ./bin/entitydbc.sh session update --description="Updated description"
   ```

4. **Switch workspaces within a session when needed**:
   ```bash
   ./bin/entitydbc.sh workspace switch <workspace-name> <session-id> [handle]
   ```

5. **End session when work is complete**:
   ```bash
   ./bin/entitydbc.sh session end
   ```

6. **List all sessions**:
   ```bash
   ./bin/entitydbc.sh session list
   ```

### Session Guidelines
- **Every development period must have a session**: Never work without an active session
- **Always start a session with the proper process**: Use the session create command to ensure compliance checks
- **Daily status check is mandatory**: Run at the beginning of each development session
- **Keep session information current**: Update issues and status as work progresses
- **Link sessions to work logs**: Each session should have a corresponding work log
- **Include session ID in commit messages**: Reference the active session in all commits
- **End sessions when switching context**: Create new sessions for different focus areas
- **Session timeouts**: Sessions inactive for 1 hour will be automatically marked as stale
- **Multi-instance compatibility**: Multiple Claude instances can have active sessions simultaneously
- **Dashboard integration**: All active sessions appear in the metrics dashboard by workspace
- **Include session ID, handle ID, and workspace in every message**: Every communication from Claude must include the current session ID, handle ID, and workspace name
- **Claude will only respond to assigned handle**: Each Claude instance must only respond when addressed by its assigned handle ID
- **Context compaction**: When summarizing conversation context, only pass the session ID for lookup; Claude will retrieve the full session context from the database
- **Autonomous work is expected**: Claude should work autonomously without requiring explicit direction unless interrupted
- **Work follows priority hierarchy**: Sprint issues → Backlog items → Recurring Issues → Autonomous improvements

### Daily Status Check
- **Run at the start of each development session**: `./bin/entitydbc.sh status check`
- **Ensures compliance with all policies**: Including "no ticket, no work"
- **Provides comprehensive status overview**: Sessions, work items, git status, compliance
- **Offers actionable recommendations**: What to do before starting work
- **Integrated with session startup**: Automatically runs when creating a session
- **Identifies policy violations**: Prevents unauthorized work without tickets
- **Verifies session and git hook setup**: Ensures proper environment configuration

### Workspace Context Management
- **CRITICAL: Always determine the active workspace from the session file**: Never assume the workspace context
- **Check current workspace at the start of each session**: Use the session status command
- **Workspace switching only updates the session.md file**: No other files are modified when switching workspaces
- **Session is the single source of truth for workspace context**: All scripts and tools should reference the session file
- **Verify workspace context before making changes**: Always confirm you're working in the intended workspace

### Metrics Dashboard Usage
- **Regular monitoring**: Check the dashboard at the start and end of each session
- **View all workspaces**: Access the web dashboard at http://localhost:8085
- **View specific workspace**: Filter by workspace in the dashboard
- **Track session activity**: Monitor active sessions and work in progress
- **Track backlog progress**: Monitor completion rates and pending work
- **See changelog**: Review recent changes by workspace

### Agent Identity and Worker ID
- **CRITICAL: Agent identity MUST be based on the WORKER_ID environment variable**
- Each agent MUST register using their WORKER_ID value as their handle
- The WORKER_ID is the definitive source of agent identity in the system
- All agents MUST check their WORKER_ID using `echo $WORKER_ID` before registration
- Agents MUST identify themselves in ALL communications using format: `[Agent: $WORKER_ID (agent-id)]`
- This identity remains consistent across sessions for the same agent
- Format for agent IDs: `ag_YYYYMMDDHHMMSS_XXXXXXXX` (automatically generated)
- The first action of every agent MUST be to register with the system using their WORKER_ID

## Git Workflow

### Branching Strategy
- **main**: Stable, production-ready code
- **sprint/[name]**: Sprint-specific branches for ongoing work
- **feature/[name]**: Feature-specific branches for major features
- **fix/[name]**: Branches for bug fixes

### Commit Guidelines

#### Commit Frequency
- Commit changes frequently (at least once per significant issue)
- Never leave uncommitted changes at the end of a work session
- Commit after completing each meaningful unit of work
- When switching workspaces, always commit pending changes first
- Smaller, more frequent commits are preferred over large, infrequent ones

#### Commit Message Structure
- Format: `[type]: [description]` (e.g., "feat: Add session management")
- Types: feat, fix, docs, style, refactor, test, chore
- First line should be concise (under 50 characters) and descriptive
- Add a detailed description after the first line when needed
- Always include the session ID at the end of the commit message
- Example:
  ```
  feat: Add workspace switching functionality
  
  - Enable switching between multiple workspaces
  - Update session tracking for workspace context
  - Add workspace-specific configuration support
  
  Session: EntityDB-2025-05-04-001
  ```

#### Commit Content
- Group related changes in a single commit
- Each commit should represent a single logical change
- Include all necessary files for a complete change
- Reference issue IDs when applicable
- Include tests with implementation changes
- Never commit credentials, secrets, or sensitive data

#### Commit Timing
- Commit at logical stopping points
- Commit before switching issues
- Commit before switching workspaces
- Commit before ending a work session
- Commit after resolving merge conflicts

### Git Integration Points
1. **Sprint Start**: Create a new sprint branch
2. **Daily Work**: Commit daily work logs and code changes
3. **Issue Completion**: Commit completed issue with descriptive message
4. **Sprint End**: Tag and merge sprint branch to main

### Git Hooks
The system provides git hooks to automatically enforce the "no ticket, no work" policy and session requirements:

1. **pre-commit**: Validates that:
   - An active session exists
   - At least one work item is marked as in-progress in the sprint
   - The commit message contains a session ID reference

2. **prepare-commit-msg**: Automatically:
   - Adds the current session ID to commit messages if missing

3. **post-commit**: Automatically:
   - Logs the commit to the session file
   - Identifies work item references in commit messages

#### Installing Git Hooks
To install the git hooks, run:
```bash
./bin/entitydbc.sh git install-hooks
```

The hooks will enforce policies for all commits. To bypass hooks temporarily (not recommended), use:
```bash
git commit --no-verify
```

### Common Git Operations
- `git status`: Check current status
- `git add [files]`: Stage changes
- `git commit -m "[message]"`: Commit changes (hooks will verify policy compliance)
- `git pull`: Update local branch
- `git push`: Push changes to remote
- `git checkout -b [branch]`: Create and switch to new branch
- `git tag [tag-name]`: Create a tag

## Issue and Work Item Management

### Work Item Verification
- **Regularly verify compliance with work item requirements**
- Run the verification command to check your work item status: `./bin/entitydbc.sh issue verify`
- This tool ensures compliance with the "no ticket, no work" policy
- Automated checks prevent untracked development work
- All team members must run this verification at the start and end of each session

### Work Item Update Requirement
- **ALL work items MUST be updated after each meaningful progress step**
- Work items are tracked in the system database
- Updates must include current status, progress percentage, and any blockers
- Dashboard views rely on accurate work item status

### Work Item Update Process
1. **Before starting work**: Mark the item as "in progress" using the issue start command
2. **During implementation**: Update progress percentage (25%, 50%, 75%)
3. **After implementation**: Mark the item as "completed" using the issue complete command
4. **If blocked**: Immediately flag the item with blocker information
5. **Daily**: Review all assigned work items to ensure status accuracy

### Work Authorization Policy
- **No ticket, no work principle**: All development work MUST be associated with a tracked work item
- Work items must be properly documented in the backlog or sprint before implementation begins
- Ad-hoc work outside of tracked items is prohibited
- Emergency fixes require retroactive creation of work items for tracking purposes
- All commits must reference the associated work item ID or ticket number
- This policy ensures complete traceability from requirements to implementation
- Exception requests must be documented and approved in the session log

## Code Commits

### Critical Commit Requirement
- **EVERY code or documentation change MUST be committed immediately after implementation**
- No exceptions - all changes must be tracked in git with proper commits
- Each discrete change should have its own commit with appropriate message
- Never leave changes uncommitted when switching issues or ending sessions

### Complete Update-Commit-Track Workflow
1. Identify a single discrete issue or change to implement
2. Implement the change in code or documentation
3. Test the change to ensure it works correctly
4. Update the corresponding work item using the issue progress command
5. Immediately commit the code/documentation changes
6. Verify dashboard reflects the changes correctly
7. Move to next issue, repeating the process

### Change Types Requiring Immediate Commits
- Code implementation (new features, bug fixes)
- Documentation updates (README, specifications, comments)
- Configuration changes (settings, environment files)
- File structure changes (new directories, file renames)
- Test additions or modifications
- Dependency updates

## Communication
- All decisions must be documented in appropriate files
- Blockers must be communicated immediately
- Changes to the sprint scope must be agreed upon by all team members
- Status updates must be provided daily through work logs
- Commit messages serve as additional communication

## Continuous Improvement
- Each retrospective should result in at least one process improvement
- Process improvements should be documented in project documentation
- The agile framework itself can be adjusted based on retrospective feedback

## Development Testing Approach

### API Testing

We use dedicated test scripts to verify API functionality:

#### Running Tests
```bash
# Ensure the server is running
./bin/entitydbd.sh status

# If the server is not running, start it
./bin/entitydbd.sh start

# Run the API tests
/opt/entitydb/src/test/test_api.sh
```

The test script performs the following checks:
- Verifies all API endpoints are accessible
- Tests basic CRUD operations for agents, sessions, issues, and workspaces
- Validates response status codes and payload structure

### Server Implementation

The EntityDB server implements a production-grade server with SQLite persistence:
- Uses SQLite database for persistent storage
- Implements all API endpoints with proper error handling and logging
- Supports RBAC (Role-Based Access Control) for secure access

To start the server:
```bash
# Start the server
/opt/entitydb/bin/entitydbd.sh start

# Stop the server
/opt/entitydb/bin/entitydbd.sh stop

# Check server status
/opt/entitydb/bin/entitydbd.sh status
```

## Role-Based Access Control (RBAC)

EntityDB implements a comprehensive RBAC system to ensure proper security and permissions.

### User Roles
- **Admin**: Full system access and management capabilities
- **Manager**: Workspace and team management, reporting capabilities
- **Agent**: Limited to assigned work and personal metrics
- **Observer**: Read-only access to specific resources

### Permission Format

All permissions follow the format `resource:action` where:
- `resource` is a system entity (e.g., `issue`, `agent`, `session`, `workspace`)
- `action` is an operation (e.g., `read`, `write`, `create`, `delete`, `assign`)

Examples:
- `issue:read`: Permission to view issues
- `agent:create`: Permission to create new agents
- `workspace:manage`: Permission to manage workspaces

### Permission Model
- **Identity-based**: Authentication tied strictly to WORKER_ID for agents
- **Assignment-based**: Agents can only view/edit work explicitly assigned to them
- **Workspace-level**: Issue creation permissions are granted at the workspace level
- **Role hierarchy**: Higher-level roles inherit permissions from lower levels
- **Direct permissions**: Individual permissions can be granted directly to users

### Role-Specific Capabilities
- **Admins**: User management, system configuration, all metrics (`*:*` permissions)
- **Managers**: Team formation, issue assignment, workspace planning, aggregated metrics
- **Agents**: Work on assigned issues, update status, view personal metrics
- **Observers**: View assigned resources without modification rights

### RBAC Management

The RBAC system can be managed through the API:

```bash
# List all roles
./bin/entitydbc.sh rbac role list

# Create a new role
./bin/entitydbc.sh rbac role create --name="Developer" --description="Software developer role"

# Add permission to role
./bin/entitydbc.sh rbac role permission add <role-id> --permission="issue:write"

# Assign role to user
./bin/entitydbc.sh rbac user role assign --username=<username> --role=<role-id>

# Check user permission
./bin/entitydbc.sh rbac user permission check --username=<username> --permission="issue:read"
```

For detailed documentation on the RBAC system, see `/opt/entitydb/docs/rbac_implementation.md`.

## Timestamp
2025-05-08 15:30:00