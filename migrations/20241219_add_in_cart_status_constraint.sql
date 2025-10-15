-- Migration: Update custom_requests status constraint to include 'in_cart'
-- Date: 2024-12-19
-- Description: Add 'in_cart' status to the allowed values in chk_custom_requests_status constraint

-- Drop existing constraint
ALTER TABLE custom_requests DROP CONSTRAINT IF EXISTS chk_custom_requests_status;

-- Add updated constraint with 'in_cart' status
ALTER TABLE custom_requests ADD CONSTRAINT chk_custom_requests_status 
    CHECK (status IN ('submitted', 'under_review', 'quote_sent', 'quote_ready', 'needs_info', 'customer_accepted', 'customer_declined', 'approved', 'in_cart', 'cancelled'));