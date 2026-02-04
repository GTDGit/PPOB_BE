-- migrations/022_recreate_products_for_gtd.sql
-- Recreate products table to match GTD API sync approach
-- WARNING: This will DROP existing products data!

-- Drop existing products table and dependencies
DROP TABLE IF EXISTS products CASCADE;

-- Create new products table for GTD API sync
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    sku_code VARCHAR(50) UNIQUE NOT NULL,          -- GTD SKU (primary identifier)
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,                  -- Pulsa, Data, PLN, PDAM, etc
    brand VARCHAR(50) NOT NULL,                     -- TELKOMSEL, INDOSAT, PLN, PDAM, etc
    type VARCHAR(20) NOT NULL,                      -- prepaid, postpaid
    price BIGINT NOT NULL,                          -- Selling price (from GTD, includes profit)
    admin BIGINT DEFAULT 0,                         -- Admin fee (for postpaid)
    commission BIGINT DEFAULT 0,                    -- Commission (info only)
    is_active BOOLEAN DEFAULT TRUE,
    description TEXT,
    gtd_updated_at TIMESTAMP,                       -- Timestamp from GTD sync
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_products_sku_code ON products(sku_code);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_brand ON products(brand);
CREATE INDEX idx_products_type ON products(type);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_category_brand ON products(category, brand);
CREATE INDEX idx_products_category_type ON products(category, type);

-- Comments
COMMENT ON TABLE products IS 'Product catalog synced from GTD API (Gerbang Transaction Data)';
COMMENT ON COLUMN products.sku_code IS 'GTD SKU code - unique identifier from external API';
COMMENT ON COLUMN products.price IS 'Selling price including profit margin from GTD';
COMMENT ON COLUMN products.admin IS 'Admin fee for postpaid bills';
COMMENT ON COLUMN products.commission IS 'Commission info from GTD (for reference only)';
COMMENT ON COLUMN products.gtd_updated_at IS 'Last sync timestamp from GTD API';
