-- Create daily_usage table
CREATE TABLE daily_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    child_id UUID NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_play_sec INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(child_id, date)
);

-- Create indexes
CREATE INDEX idx_daily_usage_child_id ON daily_usage(child_id);
CREATE INDEX idx_daily_usage_date ON daily_usage(date);
