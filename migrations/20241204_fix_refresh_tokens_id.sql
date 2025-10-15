-- Migration: Fix refresh_tokens table ID column type
-- Date: 2024-12-04
-- Description: Change refresh_tokens.id from integer to UUID

-- Drop the existing table and recreate with correct UUID type
DROP TABLE IF EXISTS refresh_tokens;

-- Recreate refresh_tokens table with UUID ID
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Add foreign key constraint
ALTER TABLE refresh_tokens 
ADD CONSTRAINT fk_refresh_tokens_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add index for better performance
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_deleted_at ON refresh_tokens(deleted_at);