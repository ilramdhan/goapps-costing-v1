-- Migration: Create mst_uom table
-- Unit of Measure master data

CREATE TABLE IF NOT EXISTS mst_uom (
    uom_code VARCHAR(20) PRIMARY KEY,
    uom_name VARCHAR(100) NOT NULL,
    uom_category VARCHAR(20) NOT NULL CHECK (uom_category IN ('WEIGHT', 'VOLUME', 'QUANTITY', 'LENGTH')),
    is_base_uom BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMPTZ,
    updated_by VARCHAR(100)
);

-- Index for category filtering
CREATE INDEX IF NOT EXISTS idx_mst_uom_category ON mst_uom(uom_category);

-- Comments
COMMENT ON TABLE mst_uom IS 'Master table for Unit of Measures';
COMMENT ON COLUMN mst_uom.uom_code IS 'Unique code for UOM, e.g., KG, TON, M';
COMMENT ON COLUMN mst_uom.uom_category IS 'Category: WEIGHT, VOLUME, QUANTITY, LENGTH';
COMMENT ON COLUMN mst_uom.is_base_uom IS 'True if this is the base UOM for its category';
