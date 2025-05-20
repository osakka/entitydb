#!/bin/bash

# EntityDB Temporal CLI commands
# Extension to entitydb-api.sh for temporal queries

# Source the main CLI for common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/entitydb-api.sh"

# Temporal command help
temporal_help() {
    echo "EntityDB Temporal Commands:"
    echo ""
    echo "  as-of <entity_id> <timestamp>"
    echo "    Get entity state as of a specific time"
    echo "    Example: entitydb-cli temporal as-of ent_123 2025-01-15T10:00:00Z"
    echo ""
    echo "  history <entity_id> [--from=<timestamp>] [--to=<timestamp>]"
    echo "    Get entity history within a time range"
    echo "    Example: entitydb-cli temporal history ent_123 --from=2025-01-01T00:00:00Z"
    echo ""
    echo "  changes [--since=<timestamp>]"
    echo "    Get entities that changed since a timestamp"
    echo "    Example: entitydb-cli temporal changes --since=2025-05-17T10:00:00Z"
    echo ""
    echo "  diff <entity_id> <timestamp1> <timestamp2>"
    echo "    Compare entity state between two timestamps"
    echo "    Example: entitydb-cli temporal diff ent_123 2025-01-01T00:00:00Z 2025-01-15T00:00:00Z"
}

# Get entity as of specific time
entity_as_of() {
    local entity_id=$1
    local timestamp=$2
    
    if [ -z "$entity_id" ] || [ -z "$timestamp" ]; then
        echo "Error: entity_id and timestamp required"
        echo "Usage: entitydb-cli temporal as-of <entity_id> <timestamp>"
        return 1
    fi
    
    api_request GET "entities/as-of?id=$entity_id&as_of=$timestamp"
}

# Get entity history
entity_history() {
    local entity_id=$1
    shift
    
    if [ -z "$entity_id" ]; then
        echo "Error: entity_id required"
        echo "Usage: entitydb-cli temporal history <entity_id> [--from=<timestamp>] [--to=<timestamp>]"
        return 1
    fi
    
    local query="id=$entity_id"
    
    # Parse optional parameters
    while [ $# -gt 0 ]; do
        case $1 in
            --from=*)
                local from="${1#*=}"
                query="$query&from=$from"
                ;;
            --to=*)
                local to="${1#*=}"
                query="$query&to=$to"
                ;;
        esac
        shift
    done
    
    api_request GET "entities/history?$query"
}

# Get recent changes
recent_changes() {
    local query=""
    
    # Parse optional parameters
    while [ $# -gt 0 ]; do
        case $1 in
            --since=*)
                local since="${1#*=}"
                query="since=$since"
                ;;
        esac
        shift
    done
    
    api_request GET "entities/changes?$query"
}

# Compare entity between timestamps
entity_diff() {
    local entity_id=$1
    local t1=$2
    local t2=$3
    
    if [ -z "$entity_id" ] || [ -z "$t1" ] || [ -z "$t2" ]; then
        echo "Error: entity_id, t1, and t2 required"
        echo "Usage: entitydb-cli temporal diff <entity_id> <timestamp1> <timestamp2>"
        return 1
    fi
    
    api_request GET "entities/diff?id=$entity_id&t1=$t1&t2=$t2"
}

# Main temporal command handler
temporal_main() {
    local subcommand=$1
    shift
    
    case $subcommand in
        as-of)
            entity_as_of "$@"
            ;;
        history)
            entity_history "$@"
            ;;
        changes)
            recent_changes "$@"
            ;;
        diff)
            entity_diff "$@"
            ;;
        help|--help|-h)
            temporal_help
            ;;
        *)
            echo "Unknown temporal command: $subcommand"
            temporal_help
            return 1
            ;;
    esac
}

# Allow direct execution
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    temporal_main "$@"
fi