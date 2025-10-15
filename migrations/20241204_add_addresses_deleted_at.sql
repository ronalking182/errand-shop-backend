-- Add deleted_at column to addresses table for soft delete functionality
ALTER TABLE addresses ADD COLUMN deleted_at TIMESTAMP NULL;

-- Create index on deleted_at for better query performance
CREATE INDEX idx_addresses_deleted_at ON addresses(deleted_at);