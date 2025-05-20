-- SQLite database schema for CCMF Task Repository

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Agents table
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    handle TEXT UNIQUE NOT NULL,
    display_name TEXT,
    type TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_active TIMESTAMP,
    specialization TEXT,
    personality_profile TEXT,
    capability_score REAL DEFAULT 0,
    worker_pool_id TEXT,
    expertise TEXT
);

-- Create indexes for agents
CREATE INDEX IF NOT EXISTS idx_agents_handle ON agents(handle);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_worker_pool_id ON agents(worker_pool_id);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    status TEXT NOT NULL,
    repository_url TEXT,
    tags TEXT
);

-- Create indexes for projects
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

-- Project members table
CREATE TABLE IF NOT EXISTS project_members (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    role TEXT NOT NULL,
    joined_at TIMESTAMP NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (agent_id) REFERENCES agents(id),
    UNIQUE(project_id, agent_id)
);

-- Create indexes for project members
CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_agent_id ON project_members(agent_id);

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT UNIQUE,
    full_name TEXT,
    created_at TIMESTAMP NOT NULL,
    last_login TIMESTAMP,
    roles TEXT NOT NULL, -- Comma-separated list of roles
    agent_id TEXT,
    active BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL
);

-- Auth tokens table
CREATE TABLE IF NOT EXISTS auth_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    agent_id TEXT,
    token_type TEXT NOT NULL CHECK (token_type IN ('access', 'refresh')),
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP,
    is_revoked BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL
);

-- Create indexes for auth tables
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_agent_id ON users(agent_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_id ON auth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires_at ON auth_tokens(expires_at);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL CHECK (type IN ('epic', 'story', 'task', 'subtask')) DEFAULT 'task',
    priority TEXT NOT NULL CHECK (priority IN ('high', 'medium', 'low')),
    estimated_effort REAL DEFAULT 0,
    due_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    created_by TEXT NOT NULL,
    project_id TEXT NOT NULL,
    parent_id TEXT,
    child_count INTEGER DEFAULT 0,
    child_completed INTEGER DEFAULT 0,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Create index on parent_id for faster hierarchy lookups
CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);

-- Task tags table (for many-to-many relationship between tasks and tags)
CREATE TABLE IF NOT EXISTS task_tags (
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (task_id, tag),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Task assignments table
CREATE TABLE IF NOT EXISTS task_assignments (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    assigned_at TIMESTAMP NOT NULL,
    assigned_by TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed', 'blocked')),
    progress INTEGER NOT NULL CHECK (progress BETWEEN 0 AND 100),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Create index on task_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_task_assignments_task_id ON task_assignments(task_id);
-- Create index on agent_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_task_assignments_agent_id ON task_assignments(agent_id);

-- Task dependencies table
CREATE TABLE IF NOT EXISTS task_dependencies (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    depends_on_id TEXT NOT NULL,
    dependency_type TEXT NOT NULL CHECK (dependency_type IN ('blocks', 'related')),
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Create indexes for faster dependency lookups
CREATE INDEX IF NOT EXISTS idx_task_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends_on_id ON task_dependencies(depends_on_id);

-- Task history table
CREATE TABLE IF NOT EXISTS task_history (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    agent_id TEXT,
    timestamp TIMESTAMP NOT NULL,
    from_status TEXT,
    to_status TEXT,
    comment TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL
);

-- Create index on task_id for faster history lookups
CREATE INDEX IF NOT EXISTS idx_task_history_task_id ON task_history(task_id);

-- Task metrics table for detailed performance tracking
CREATE TABLE IF NOT EXISTS task_metrics (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    metric_value REAL NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    recorded_by TEXT,
    notes TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Create index on task_id for faster metrics lookups
CREATE INDEX IF NOT EXISTS idx_task_metrics_task_id ON task_metrics(task_id);
CREATE INDEX IF NOT EXISTS idx_task_metrics_type ON task_metrics(metric_type);

-- Task time tracking table for detailed time records
CREATE TABLE IF NOT EXISTS task_time_records (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_seconds INTEGER,
    activity_type TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Create index for time records
CREATE INDEX IF NOT EXISTS idx_task_time_records_task_id ON task_time_records(task_id);
CREATE INDEX IF NOT EXISTS idx_task_time_records_agent_id ON task_time_records(agent_id);

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL CHECK (status IN ('active', 'paused', 'ended', 'stale')),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    context_file TEXT,
    tags TEXT NOT NULL DEFAULT '[]', -- Stored as JSON array
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes for sessions
CREATE INDEX IF NOT EXISTS idx_sessions_agent_id ON sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_sessions_project_id ON sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON sessions(updated_at);

-- Session contexts table
CREATE TABLE IF NOT EXISTS session_contexts (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    UNIQUE(session_id, key)
);

-- Create indexes for session contexts
CREATE INDEX IF NOT EXISTS idx_session_contexts_session_id ON session_contexts(session_id);