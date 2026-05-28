-- Create play_events table
CREATE TABLE play_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    child_id UUID NOT NULL REFERENCES children(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('open', 'start', 'pause', 'complete', 'favorite')),
    position_sec INTEGER DEFAULT 0,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_play_events_child_id ON play_events(child_id);
CREATE INDEX idx_play_events_item_id ON play_events(item_id);
CREATE INDEX idx_play_events_occurred_at ON play_events(occurred_at);
