# EntityDB Workaround Scripts

This document describes the usage of workaround scripts created to access agent and session information directly from the database. These scripts provide an alternative way to interact with agents and sessions while the API endpoints are being fixed.

## Agent Management Script

The `/opt/entitydb/bin/check_agent.sh` script provides direct database access to agent information.

### List All Agents

```bash
/opt/entitydb/bin/check_agent.sh list
```

This command will display all agents in the database with their key properties (ID, handle, display name, type, status, etc.).

### View Agent Details

```bash
/opt/entitydb/bin/check_agent.sh view <agent_handle_or_id>
```

This command will display detailed information about a specific agent, including:
- All agent properties
- Agent capabilities (if any)
- Agent performance metrics (if any)

Examples:
```bash
# View by handle
/opt/entitydb/bin/check_agent.sh view claude-2

# View by ID
/opt/entitydb/bin/check_agent.sh view agent_claude
```

## Session Management Script

The `/opt/entitydb/bin/check_session.sh` script provides direct database access to session information and allows creating sessions directly in the database.

### List All Sessions

```bash
/opt/entitydb/bin/check_session.sh list
```

This command will display all sessions in the database along with related agent and workspace information.

### Create a New Session

```bash
/opt/entitydb/bin/check_session.sh create <agent> <workspace> <name> [description]
```

This command will create a new session directly in the database, bypassing the API.

Parameters:
- `agent`: Agent handle or ID
- `workspace`: Workspace ID or name
- `name`: Session name
- `description`: (Optional) Session description

Examples:
```bash
# Create session with agent handle
/opt/entitydb/bin/check_session.sh create claude-2 workspace_entitydb "Test Session" "This is a test session"

# Create session with agent ID
/opt/entitydb/bin/check_session.sh create agent_claude workspace_entitydb "Another Session" "Another test session"
```

### View Session Details

```bash
/opt/entitydb/bin/check_session.sh view <session_id>
```

This command will display detailed information about a specific session, including:
- All session properties
- Associated agent and workspace
- Session context data (if any)

Example:
```bash
/opt/entitydb/bin/check_session.sh view sess_1234567890_1234
```

## Important Notes

1. These scripts are temporary workarounds while the API endpoints are being fixed.
2. They access the database directly, bypassing authentication and authorization checks.
3. Use caution when creating or modifying data with these scripts.
4. The scripts are designed to be used by system administrators for maintenance purposes.

## Future Improvements

Once the API endpoints are fixed, the following improvements should be made:

1. Update the main client tool (`entitydbc.sh`) to properly handle agent IDs and session creation.
2. Fix the agent listing endpoint to return proper agent information.
3. Fix the agent profile endpoint to correctly accept agent IDs.
4. Fix the session creation endpoint to properly handle agent IDs.

After these fixes are implemented, these workaround scripts will no longer be necessary.