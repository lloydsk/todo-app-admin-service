-- Initial schema for Todo App Admin Service
-- This migration creates all the core tables needed for the application

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'admin')),
    password_hash TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for users table
CREATE INDEX idx_users_email ON users(email) WHERE NOT is_deleted;
CREATE INDEX idx_users_role ON users(role) WHERE NOT is_deleted;
CREATE INDEX idx_users_is_deleted ON users(is_deleted);

-- Categories table
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7) NOT NULL DEFAULT '#6B7280', -- Default gray color
    parent_id UUID REFERENCES categories(id),
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    creator_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for categories table
CREATE INDEX idx_categories_name ON categories(name) WHERE NOT is_deleted;
CREATE INDEX idx_categories_creator_id ON categories(creator_id) WHERE NOT is_deleted;
CREATE INDEX idx_categories_parent_id ON categories(parent_id) WHERE NOT is_deleted;
CREATE INDEX idx_categories_is_public ON categories(is_public) WHERE NOT is_deleted;
CREATE INDEX idx_categories_is_deleted ON categories(is_deleted);

-- Tags table
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#10B981', -- Default green color
    creator_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for tags table
CREATE UNIQUE INDEX idx_tags_name_creator_unique ON tags(name, creator_id) WHERE NOT is_deleted;
CREATE INDEX idx_tags_name ON tags(name) WHERE NOT is_deleted;
CREATE INDEX idx_tags_creator_id ON tags(creator_id) WHERE NOT is_deleted;
CREATE INDEX idx_tags_is_deleted ON tags(is_deleted);

-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    assignee_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM' CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT')),
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for tasks table
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id) WHERE NOT is_deleted;
CREATE INDEX idx_tasks_status ON tasks(status) WHERE NOT is_deleted;
CREATE INDEX idx_tasks_priority ON tasks(priority) WHERE NOT is_deleted;
CREATE INDEX idx_tasks_due_date ON tasks(due_date) WHERE NOT is_deleted AND due_date IS NOT NULL;
CREATE INDEX idx_tasks_created_at ON tasks(created_at) WHERE NOT is_deleted;
CREATE INDEX idx_tasks_is_deleted ON tasks(is_deleted);

-- Task Categories junction table
CREATE TABLE task_categories (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, category_id)
);

-- Create indexes for task_categories table
CREATE INDEX idx_task_categories_task_id ON task_categories(task_id);
CREATE INDEX idx_task_categories_category_id ON task_categories(category_id);

-- Task Tags junction table
CREATE TABLE task_tags (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, tag_id)
);

-- Create indexes for task_tags table
CREATE INDEX idx_task_tags_task_id ON task_tags(task_id);
CREATE INDEX idx_task_tags_tag_id ON task_tags(tag_id);

-- Task History table for audit trail
CREATE TABLE task_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    actor_id UUID NOT NULL REFERENCES users(id),
    service_name VARCHAR(100) NOT NULL DEFAULT 'admin-service',
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    details JSONB
);

-- Create indexes for task_history table
CREATE INDEX idx_task_history_task_id ON task_history(task_id);
CREATE INDEX idx_task_history_timestamp ON task_history(timestamp);
CREATE INDEX idx_task_history_actor_id ON task_history(actor_id);
CREATE INDEX idx_task_history_action ON task_history(action);

-- Task Reminders table (optional future feature)
CREATE TABLE task_reminders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reminder_time TIMESTAMP WITH TIME ZONE NOT NULL,
    message TEXT,
    is_sent BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(task_id, user_id, reminder_time)
);

-- Create indexes for task_reminders table
CREATE INDEX idx_task_reminders_task_id ON task_reminders(task_id);
CREATE INDEX idx_task_reminders_user_id ON task_reminders(user_id);
CREATE INDEX idx_task_reminders_reminder_time ON task_reminders(reminder_time) WHERE NOT is_sent;

-- User Sessions table (for authentication)
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_agent TEXT,
    ip_address INET
);

-- Create indexes for user_sessions table
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_token_hash ON user_sessions(token_hash);

-- Create triggers to automatically update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply the trigger to all tables with updated_at columns
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create a function to handle optimistic locking
CREATE OR REPLACE FUNCTION check_version()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.version != NEW.version - 1 THEN
        RAISE EXCEPTION 'Optimistic lock error: version mismatch. Expected %, got %', OLD.version + 1, NEW.version;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply version checking to tables with version columns
CREATE TRIGGER check_users_version BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION check_version();
CREATE TRIGGER check_categories_version BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION check_version();
CREATE TRIGGER check_tags_version BEFORE UPDATE ON tags FOR EACH ROW EXECUTE FUNCTION check_version();
CREATE TRIGGER check_tasks_version BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION check_version();