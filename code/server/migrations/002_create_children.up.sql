-- Create children table
CREATE TABLE children (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    family_id UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    device_id VARCHAR(100),
    daily_time_limit_sec INTEGER DEFAULT 3600,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on family_id for fast lookup
CREATE INDEX idx_children_family_id ON children(family_id);
CREATE INDEX idx_children_device_id ON children(device_id);
