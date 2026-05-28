-- Create families table
CREATE TABLE families (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pairing_code VARCHAR(8) UNIQUE NOT NULL,
    pairing_code_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on pairing_code for fast lookup
CREATE INDEX idx_families_pairing_code ON families(pairing_code);
