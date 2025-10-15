-- Fix OTPs table to use UUID for id column
DROP TABLE IF EXISTS otps CASCADE;

CREATE TABLE otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL,
    type VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Add foreign key constraint
ALTER TABLE otps ADD CONSTRAINT fk_otps_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for better performance
CREATE INDEX idx_otps_user_id ON otps(user_id);
CREATE INDEX idx_otps_email ON otps(email);
CREATE INDEX idx_otps_code ON otps(code);
CREATE INDEX idx_otps_type ON otps(type);
CREATE INDEX idx_otps_expires_at ON otps(expires_at);
CREATE INDEX idx_otps_deleted_at ON otps(deleted_at);