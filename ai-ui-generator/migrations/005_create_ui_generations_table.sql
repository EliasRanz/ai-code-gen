-- +migrate Up
-- Create generation_status enum
CREATE TYPE generation_status AS ENUM ('pending', 'in_progress', 'completed', 'failed', 'cancelled');

-- Create ui_generations table for tracking AI-generated UI components
CREATE TABLE ui_generations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    chat_session_id UUID REFERENCES chat_sessions(id) ON DELETE SET NULL,
    chat_message_id UUID REFERENCES chat_messages(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    prompt TEXT NOT NULL, -- The original prompt used for generation
    status generation_status DEFAULT 'pending',
    component_type VARCHAR(100), -- 'component', 'page', 'layout', etc.
    framework VARCHAR(50), -- 'react', 'vue', 'angular', etc.
    generated_code TEXT, -- The generated code
    preview_url TEXT, -- URL to preview the component
    assets JSONB DEFAULT '{}', -- Generated assets (images, icons, etc.)
    metadata JSONB DEFAULT '{}', -- Generation metadata (model used, tokens, etc.)
    error_message TEXT, -- Error details if generation failed
    processing_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_ui_generations_user_id ON ui_generations(user_id);
CREATE INDEX idx_ui_generations_project_id ON ui_generations(project_id);
CREATE INDEX idx_ui_generations_chat_session_id ON ui_generations(chat_session_id);
CREATE INDEX idx_ui_generations_status ON ui_generations(status);
CREATE INDEX idx_ui_generations_component_type ON ui_generations(component_type);
CREATE INDEX idx_ui_generations_framework ON ui_generations(framework);
CREATE INDEX idx_ui_generations_created_at ON ui_generations(created_at);
CREATE INDEX idx_ui_generations_metadata ON ui_generations USING GIN(metadata);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_ui_generations_updated_at
    BEFORE UPDATE ON ui_generations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +migrate Down
-- Drop ui_generations table and related objects
DROP TRIGGER IF EXISTS update_ui_generations_updated_at ON ui_generations;
DROP TABLE IF EXISTS ui_generations;
DROP TYPE IF EXISTS generation_status;
