-- Fix FCM Tables UUID Migration
-- Drop and recreate FCM tables with proper UUID types

BEGIN;

-- Drop existing FCM tables (they likely have no important data)
DROP TABLE IF EXISTS fcm_message_recipients CASCADE;
DROP TABLE IF EXISTS fcm_messages CASCADE;
DROP TABLE IF EXISTS fcm_tokens CASCADE;

-- Recreate fcm_messages table with UUID
CREATE TABLE fcm_messages (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    data JSONB,
    image_url VARCHAR(500),
    message_type VARCHAR(20) NOT NULL,
    sent_by UUID NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (sent_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Recreate fcm_tokens table with UUID
CREATE TABLE fcm_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    token TEXT NOT NULL,
    device_type VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, token)
);

-- Recreate fcm_message_recipients table with UUID
CREATE TABLE fcm_message_recipients (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (message_id) REFERENCES fcm_messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Update notifications table if recipient_id is not UUID
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'notifications' 
               AND column_name = 'recipient_id' 
               AND data_type != 'uuid') THEN
        -- Drop existing data and change type
        DELETE FROM notifications;
        ALTER TABLE notifications ALTER COLUMN recipient_id TYPE UUID USING gen_random_uuid();
    END IF;
END $$;

-- Update push_tokens table if user_id is not UUID
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'push_tokens' 
               AND column_name = 'user_id' 
               AND data_type != 'uuid') THEN
        -- Drop existing data and change type
        DELETE FROM push_tokens;
        ALTER TABLE push_tokens ALTER COLUMN user_id TYPE UUID USING gen_random_uuid();
        -- Add foreign key if not exists
        IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                       WHERE constraint_name = 'fk_push_tokens_user_id') THEN
            ALTER TABLE push_tokens ADD CONSTRAINT fk_push_tokens_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END IF;
END $$;

COMMIT;

-- Verify the changes
\echo 'FCM Tables Updated Successfully';
\d fcm_messages;
\d fcm_tokens;
\d fcm_message_recipients;