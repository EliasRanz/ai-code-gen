-- +migrate Up
-- Create message_role enum
CREATE TYPE message_role AS ENUM ('user', 'assistant', 'system', 'function');

-- Create message_type enum  
CREATE TYPE message_type AS ENUM ('text', 'code', 'image', 'file', 'ui_component', 'error');

-- Create chat_messages table
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    parent_message_id UUID REFERENCES chat_messages(id) ON DELETE SET NULL,
    role message_role NOT NULL,
    type message_type DEFAULT 'text',
    content TEXT NOT NULL,
    content_metadata JSONB DEFAULT '{}', -- For structured content like code blocks, UI components
    tokens_used INTEGER DEFAULT 0,
    model_used VARCHAR(100), -- Track which AI model was used
    processing_time_ms INTEGER, -- Track response time
    sequence_number INTEGER NOT NULL, -- Order within the session
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_chat_messages_session_id ON chat_messages(chat_session_id);
CREATE INDEX idx_chat_messages_parent_id ON chat_messages(parent_message_id);
CREATE INDEX idx_chat_messages_role ON chat_messages(role);
CREATE INDEX idx_chat_messages_type ON chat_messages(type);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);
CREATE INDEX idx_chat_messages_sequence ON chat_messages(chat_session_id, sequence_number);
CREATE INDEX idx_chat_messages_content_metadata ON chat_messages USING GIN(content_metadata);

-- Create unique constraint for sequence number within session
CREATE UNIQUE INDEX idx_chat_messages_session_sequence 
    ON chat_messages(chat_session_id, sequence_number);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_chat_messages_updated_at
    BEFORE UPDATE ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger to update message count in chat_sessions
CREATE OR REPLACE FUNCTION update_chat_session_message_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE chat_sessions 
        SET message_count = message_count + 1,
            last_message_at = NEW.created_at
        WHERE id = NEW.chat_session_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE chat_sessions 
        SET message_count = message_count - 1
        WHERE id = OLD.chat_session_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_session_message_count_on_insert
    AFTER INSERT ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_session_message_count();

CREATE TRIGGER update_session_message_count_on_delete
    AFTER DELETE ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_session_message_count();

-- +migrate Down
-- Drop chat_messages table and related objects
DROP TRIGGER IF EXISTS update_session_message_count_on_delete ON chat_messages;
DROP TRIGGER IF EXISTS update_session_message_count_on_insert ON chat_messages;
DROP TRIGGER IF EXISTS update_chat_messages_updated_at ON chat_messages;
DROP FUNCTION IF EXISTS update_chat_session_message_count();
DROP TABLE IF EXISTS chat_messages;
DROP TYPE IF EXISTS message_type;
DROP TYPE IF EXISTS message_role;
