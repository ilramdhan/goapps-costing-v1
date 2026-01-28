-- Migration: Create mst_parameter table
-- Parameter master data for costing configuration

CREATE TABLE IF NOT EXISTS mst_parameter (
    parameter_code VARCHAR(50) PRIMARY KEY,
    parameter_name VARCHAR(200) NOT NULL,
    parameter_category VARCHAR(20) NOT NULL CHECK (parameter_category IN ('MACHINE', 'MATERIAL', 'QUALITY', 'OUTPUT', 'PROCESS')),
    data_type VARCHAR(20) NOT NULL CHECK (data_type IN ('NUMERIC', 'TEXT', 'BOOLEAN', 'DROPDOWN')),
    uom VARCHAR(20),
    min_value DECIMAL(18,6),
    max_value DECIMAL(18,6),
    allowed_values JSONB, -- For dropdown options: ["option1", "option2"]
    is_mandatory BOOLEAN DEFAULT FALSE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMPTZ,
    updated_by VARCHAR(100),
    
    -- Foreign key to UOM (optional)
    CONSTRAINT fk_mst_parameter_uom FOREIGN KEY (uom) REFERENCES mst_uom(uom_code) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_mst_parameter_category ON mst_parameter(parameter_category);
CREATE INDEX IF NOT EXISTS idx_mst_parameter_active ON mst_parameter(is_active);
CREATE INDEX IF NOT EXISTS idx_mst_parameter_data_type ON mst_parameter(data_type);

-- Comments
COMMENT ON TABLE mst_parameter IS 'Master table for costing parameters';
COMMENT ON COLUMN mst_parameter.parameter_code IS 'Unique code for parameter';
COMMENT ON COLUMN mst_parameter.parameter_category IS 'Category: MACHINE, MATERIAL, QUALITY, OUTPUT, PROCESS';
COMMENT ON COLUMN mst_parameter.data_type IS 'Data type: NUMERIC, TEXT, BOOLEAN, DROPDOWN';
COMMENT ON COLUMN mst_parameter.allowed_values IS 'JSON array of allowed values for DROPDOWN type';
