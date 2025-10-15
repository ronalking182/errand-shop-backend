-- Add missing user_type column to FCM tables
-- This column is referenced in the Go models but missing from database

BEGIN;

-- Add user_type column to fcm_tokens table
ALTER TABLE fcm_tokens 
ADD COLUMN IF NOT EXISTS user_type VARCHAR(20) NOT NULL DEFAULT 'customer';

-- Add user_type column to fcm_message_recipients table
ALTER TABLE fcm_message_recipients 
ADD COLUMN IF NOT EXISTS user_type VARCHAR(20) NOT NULL DEFAULT 'customer';

-- Add missing columns to fcm_tokens if they don't exist
ALTER TABLE fcm_tokens 
ADD COLUMN IF NOT EXISTS device_type VARCHAR(20),
ADD COLUMN IF NOT EXISTS device_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS platform VARCHAR(20);

-- Add missing columns to fcm_message_recipients if they don't exist
ALTER TABLE fcm_message_recipients 
ADD COLUMN IF NOT EXISTS token_id BIGINT,
ADD COLUMN IF NOT EXISTS error TEXT,
ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMP WITH TIME ZONE;

-- Update existing records to have default user_type
UPDATE fcm_tokens SET user_type = 'customer' WHERE user_type IS NULL;
UPDATE fcm_message_recipients SET user_type = 'customer' WHERE user_type IS NULL;

COMMIT;

-- Verify the changes
\echo 'FCM Tables Updated with user_type column';
\d fcm_tokens;
\d fcm_message_recipients;