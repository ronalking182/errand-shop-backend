-- Migration: Update custom_requests status constraint to include 'quote_ready'
-- Date: 2024-12-18
-- Description: Add 'quote_ready' status to the allowed values in chk_custom_requests_status constraint

-- Drop existing constraint
ALTER TABLE custom_requests DROP CONSTRAINT IF EXISTS chk_custom_requests_status;

-- Add updated constraint with 'quote_ready' status
ALTER TABLE custom_requests ADD CONSTRAINT chk_custom_requests_status 
    CHECK (status IN ('submitted', 'under_review', 'quote_sent', 'quote_ready', 'needs_info', 'customer_accepted', 'customer_declined', 'approved', 'cancelled'));