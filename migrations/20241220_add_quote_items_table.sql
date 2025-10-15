-- Migration: Add quote_items table
-- This migration adds the quote_items table to store individual quoted items
-- and establishes the relationship between quotes and quote items

-- Create quote_items table
CREATE TABLE IF NOT EXISTS quote_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id UUID NOT NULL,
    request_item_id UUID NOT NULL,
    quoted_price BIGINT NOT NULL, -- in kobo
    admin_notes TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key constraints
    CONSTRAINT fk_quote_items_quote_id 
        FOREIGN KEY (quote_id) 
        REFERENCES quotes(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_quote_items_request_item_id 
        FOREIGN KEY (request_item_id) 
        REFERENCES request_items(id) 
        ON DELETE CASCADE,
    
    -- Ensure unique combination of quote_id and request_item_id
    CONSTRAINT unique_quote_request_item 
        UNIQUE (quote_id, request_item_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_quote_items_quote_id ON quote_items(quote_id);
CREATE INDEX IF NOT EXISTS idx_quote_items_request_item_id ON quote_items(request_item_id);
CREATE INDEX IF NOT EXISTS idx_quote_items_created_at ON quote_items(created_at);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_quote_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_quote_items_updated_at
    BEFORE UPDATE ON quote_items
    FOR EACH ROW
    EXECUTE FUNCTION update_quote_items_updated_at();