-- Migration: Add first_name and last_name fields to users table
-- Date: 2024-12-03
-- Description: Add first_name and last_name columns and make phone required for mobile app compatibility

-- Add first_name and last_name columns
ALTER TABLE users 
ADD COLUMN first_name VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN last_name VARCHAR(255) NOT NULL DEFAULT '';

-- Make phone column required (remove nullable constraint)
ALTER TABLE users 
ALTER COLUMN phone SET NOT NULL;

-- Update existing users with name split (for existing data)
-- This will split the name field into first_name and last_name
UPDATE users 
SET 
    first_name = CASE 
        WHEN position(' ' in name) > 0 THEN substring(name from 1 for position(' ' in name) - 1)
        ELSE name
    END,
    last_name = CASE 
        WHEN position(' ' in name) > 0 THEN substring(name from position(' ' in name) + 1)
        ELSE ''
    END
WHERE first_name = '' AND last_name = '';

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_first_name ON users(first_name);
CREATE INDEX IF NOT EXISTS idx_users_last_name ON users(last_name);

-- Update any users with null phone to have a default value (if any exist)
UPDATE users SET phone = '+1234567890' WHERE phone IS NULL;