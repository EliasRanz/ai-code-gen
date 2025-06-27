-- +migrate Up
-- Add password field to users table
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);

-- Add username field for consistency with domain model
ALTER TABLE users ADD COLUMN username VARCHAR(100) UNIQUE;

-- Add role field for consistency with domain model (single role, not array)
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user';

-- Add status field for consistency with domain model
ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active';

-- Update column naming for consistency
ALTER TABLE users RENAME COLUMN is_active TO active;

-- Create indexes for new fields
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- +migrate Down
-- Remove added columns and revert changes
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_status;

ALTER TABLE users RENAME COLUMN active TO is_active;
ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users DROP COLUMN IF EXISTS username;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
