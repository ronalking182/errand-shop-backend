-- Migration: Update FCM Tables to Use UUID for User IDs
-- Date: 2024-12-04
-- Description: Fix FCM system type mismatch by changing user_id columns from integer to UUID

BEGIN;

-- Update fcm_tokens table
ALTER TABLE fcm_tokens 
ALTER COLUMN user_id TYPE UUID USING user_id::text::uuid;

-- Update fcm_messages table (sent_by column)
ALTER TABLE fcm_messages 
ALTER COLUMN sent_by TYPE UUID USING sent_by::text::uuid;

-- Update fcm_message_recipients table
ALTER TABLE fcm_message_recipients 
ALTER COLUMN user_id TYPE UUID USING user_id::text::uuid;

-- Update push_tokens table (if exists)
ALTER TABLE push_tokens 
ALTER COLUMN user_id TYPE UUID USING user_id::text::uuid;

-- Update notifications table (recipient_id column)
ALTER TABLE notifications 
ALTER COLUMN recipient_id TYPE UUID USING recipient_id::text::uuid;

-- Add foreign key constraints to ensure referential integrity
ALTER TABLE fcm_tokens 
ADD CONSTRAINT fk_fcm_tokens_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE fcm_messages 
ADD CONSTRAINT fk_fcm_messages_sent_by 
FOREIGN KEY (sent_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE fcm_message_recipients 
ADD CONSTRAINT fk_fcm_message_recipients_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE push_tokens 
ADD CONSTRAINT fk_push_tokens_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE notifications 
ADD CONSTRAINT fk_notifications_recipient_id 
FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_fcm_tokens_user_id ON fcm_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_fcm_messages_sent_by ON fcm_messages(sent_by);
CREATE INDEX IF NOT EXISTS idx_fcm_message_recipients_user_id ON fcm_message_recipients(user_id);
CREATE INDEX IF NOT EXISTS idx_push_tokens_user_id ON push_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_id ON notifications(recipient_id);

COMMIT;

-- Add comments for documentation
COMMENT ON COLUMN fcm_tokens.user_id IS 'UUID reference to users table';
COMMENT ON COLUMN fcm_messages.sent_by IS 'UUID reference to users table (nullable)';
COMMENT ON COLUMN fcm_message_recipients.user_id IS 'UUID reference to users table';
COMMENT ON COLUMN push_tokens.user_id IS 'UUID reference to users table';
COMMENT ON COLUMN notifications.recipient_id IS 'UUID reference to users table';