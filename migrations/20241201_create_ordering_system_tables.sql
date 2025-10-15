-- Migration: Create Ordering System Tables
-- This migration creates all the missing tables required for the complete ordering system
-- as specified in ORDER.MD

-- Custom Requests Table
CREATE TABLE IF NOT EXISTS custom_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    delivery_address_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'submitted',
    priority VARCHAR(20) DEFAULT 'MEDIUM',
    allow_substitutions BOOLEAN DEFAULT true,
    notes TEXT,
    assignee_id UUID,
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

-- Request Items Table
CREATE TABLE IF NOT EXISTS request_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    custom_request_id UUID NOT NULL REFERENCES custom_requests(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    quantity NUMERIC NOT NULL,
    unit VARCHAR(50),
    preferred_brand VARCHAR(255),
    estimated_price INTEGER, -- in kobo
    quoted_price INTEGER,    -- in kobo
    admin_notes TEXT,
    images JSONB DEFAULT '[]'::jsonb
);

-- Quotes Table
CREATE TABLE IF NOT EXISTS quotes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    custom_request_id UUID NOT NULL REFERENCES custom_requests(id) ON DELETE CASCADE,
    items_subtotal INTEGER NOT NULL, -- in kobo
    fees JSONB NOT NULL,            -- {delivery, service, packaging}
    fees_total INTEGER NOT NULL,    -- in kobo
    grand_total INTEGER NOT NULL,   -- in kobo
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    valid_until TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Update Orders Table to match ORDER.MD specification
ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_number VARCHAR(50) UNIQUE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS address_id UUID;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS subtotal INTEGER DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount INTEGER DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_fee INTEGER DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS total INTEGER DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS coupon_id UUID;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS notes TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(255);

-- Update Order Items Table
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS name VARCHAR(255);
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS sku VARCHAR(100);
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS unit_price INTEGER; -- in kobo
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS total_price INTEGER; -- in kobo
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'catalog'; -- 'catalog' | 'custom_request'

-- Update Payments Table to match ORDER.MD specification
ALTER TABLE payments ADD COLUMN IF NOT EXISTS provider VARCHAR(50); -- payment provider
ALTER TABLE payments ADD COLUMN IF NOT EXISTS reference VARCHAR(255) UNIQUE;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'NGN';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS raw_payload JSONB;

-- Order Status History Table
CREATE TABLE IF NOT EXISTS order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    by_admin_id UUID,
    note TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Order Sequences Table for order number generation
CREATE TABLE IF NOT EXISTS order_sequences (
    id SERIAL PRIMARY KEY,
    last_number INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert initial sequence record
INSERT INTO order_sequences (last_number) VALUES (0) ON CONFLICT DO NOTHING;

-- Custom Request Messages Table (for admin-user communication)
CREATE TABLE IF NOT EXISTS custom_request_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    custom_request_id UUID NOT NULL REFERENCES custom_requests(id) ON DELETE CASCADE,
    sender_type VARCHAR(20) NOT NULL, -- 'user' | 'admin'
    sender_id UUID NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Coupon Redemptions Table (if not exists)
CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_id UUID NOT NULL,
    user_id UUID NOT NULL,
    order_id UUID REFERENCES orders(id),
    redeemed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create Indexes for Performance
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);

CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);

CREATE INDEX IF NOT EXISTS idx_custom_requests_status ON custom_requests(status);
CREATE INDEX IF NOT EXISTS idx_custom_requests_user_id ON custom_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_custom_requests_submitted_at ON custom_requests(submitted_at);

CREATE INDEX IF NOT EXISTS idx_quotes_custom_request_id ON quotes(custom_request_id);
CREATE INDEX IF NOT EXISTS idx_quotes_status ON quotes(status);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_created_at ON order_status_history(created_at);

CREATE INDEX IF NOT EXISTS idx_request_items_custom_request_id ON request_items(custom_request_id);
CREATE INDEX IF NOT EXISTS idx_custom_request_messages_request_id ON custom_request_messages(custom_request_id);

-- Add constraints
ALTER TABLE custom_requests ADD CONSTRAINT chk_custom_requests_status 
    CHECK (status IN ('submitted', 'under_review', 'quote_sent', 'quote_ready', 'needs_info', 'customer_accepted', 'customer_declined', 'approved', 'cancelled'));

ALTER TABLE custom_requests ADD CONSTRAINT chk_custom_requests_priority 
    CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT'));

ALTER TABLE quotes ADD CONSTRAINT chk_quotes_status 
    CHECK (status IN ('DRAFT', 'SENT', 'ACCEPTED', 'DECLINED'));

ALTER TABLE order_items ADD CONSTRAINT chk_order_items_source 
    CHECK (source IN ('catalog', 'custom_request'));

ALTER TABLE custom_request_messages ADD CONSTRAINT chk_messages_sender_type 
    CHECK (sender_type IN ('user', 'admin'));

-- Update order status constraints to match ORDER.MD
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT chk_orders_status 
    CHECK (status IN ('pending', 'confirmed', 'preparing', 'out_for_delivery', 'delivered', 'cancelled'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_status_check;
ALTER TABLE orders ADD CONSTRAINT chk_orders_payment_status 
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded'));

-- Comments for documentation
COMMENT ON TABLE custom_requests IS 'User requests for items not in catalog';
COMMENT ON TABLE request_items IS 'Individual items within a custom request';
COMMENT ON TABLE quotes IS 'Admin quotes for custom requests';
COMMENT ON TABLE order_status_history IS 'Audit trail for order status changes';
COMMENT ON TABLE order_sequences IS 'Sequence generator for order numbers';
COMMENT ON TABLE custom_request_messages IS 'Communication between users and admins';

COMMENT ON COLUMN orders.order_number IS 'Human-readable order number (ORD-000123)';
COMMENT ON COLUMN orders.subtotal IS 'Order subtotal in kobo before discounts';
COMMENT ON COLUMN orders.discount IS 'Discount amount in kobo';
COMMENT ON COLUMN orders.delivery_fee IS 'Delivery fee in kobo';
COMMENT ON COLUMN orders.total IS 'Final total in kobo after discounts and fees';

COMMENT ON COLUMN payments.reference IS 'Unique payment reference for provider';
COMMENT ON COLUMN payments.provider IS 'Payment provider type';
COMMENT ON COLUMN payments.raw_payload IS 'Raw response from payment provider';

COMMENT ON COLUMN quotes.fees IS 'JSON object with delivery, service, packaging fees';
COMMENT ON COLUMN quotes.items_subtotal IS 'Sum of all quoted item prices in kobo';
COMMENT ON COLUMN quotes.fees_total IS 'Sum of all fees in kobo';
COMMENT ON COLUMN quotes.grand_total IS 'Final total including items and fees in kobo';