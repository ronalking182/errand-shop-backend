-- Fix users table ID column type from integer to UUID
-- This script will backup existing data and recreate the table with proper UUID IDs

BEGIN;

-- Create a backup of existing users data
CREATE TABLE users_backup AS SELECT * FROM users;

-- Drop existing foreign key constraints that reference users.id
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS fk_refresh_tokens_user;
ALTER TABLE addresses DROP CONSTRAINT IF EXISTS fk_addresses_user;
ALTER TABLE otps DROP CONSTRAINT IF EXISTS fk_otps_user;

-- Drop the existing users table
DROP TABLE IF EXISTS users CASCADE;

-- Recreate users table with UUID ID
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(255) UNIQUE,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255),
    role VARCHAR(255) DEFAULT 'customer',
    permissions JSONB,
    status VARCHAR(255) DEFAULT 'active',
    is_verified BOOLEAN DEFAULT false,
    force_reset BOOLEAN DEFAULT false,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Create index on deleted_at for soft deletes
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Update related tables to use UUID foreign keys
ALTER TABLE refresh_tokens ALTER COLUMN user_id TYPE UUID USING gen_random_uuid();
ALTER TABLE addresses ALTER COLUMN user_id TYPE UUID USING gen_random_uuid();
ALTER TABLE otps ALTER COLUMN user_id TYPE UUID USING gen_random_uuid();

-- Add foreign key constraints back
ALTER TABLE refresh_tokens ADD CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE addresses ADD CONSTRAINT fk_addresses_user FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE otps ADD CONSTRAINT fk_otps_user FOREIGN KEY (user_id) REFERENCES users(id);

COMMIT;