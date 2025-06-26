-- +migrate Up
-- Create chat_session_status enum
CREATE TYPE chat_session_status AS ENUM ('active', 'paused', 'completed', 'archived');

-- Create chat_sessions table
CREATE TABLE chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status chat_session_status DEFAULT 'active',
    context JSONB DEFAULT '{}', -- Store conversation context and settings
    metadata JSONB DEFAULT '{}', -- Additional metadata
    message_count INTEGER DEFAULT 0,
    last_message_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE, -- For session expiry
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_chat_sessions_user_id ON chat_sessions(user_id);
CREATE INDEX idx_chat_sessions_project_id ON chat_sessions(project_id);
CREATE INDEX idx_chat_sessions_status ON chat_sessions(status);
CREATE INDEX idx_chat_sessions_created_at ON chat_sessions(created_at);
CREATE INDEX idx_chat_sessions_last_message_at ON chat_sessions(last_message_at);
CREATE INDEX idx_chat_sessions_expires_at ON chat_sessions(expires_at);
CREATE INDEX idx_chat_sessions_context ON chat_sessions USING GIN(context);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_chat_sessions_updated_at
    BEFORE UPDATE ON chat_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +migrate Down
-- Drop chat_sessions table and related objects
DROP TRIGGER IF EXISTS update_chat_sessions_updated_at ON chat_sessions;
DROP TABLE IF EXISTS chat_sessions;
DROP TYPE IF EXISTS chat_session_status;
