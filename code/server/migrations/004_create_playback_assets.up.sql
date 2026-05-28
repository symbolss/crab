-- Create playback_assets table
CREATE TABLE playback_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    kind VARCHAR(50) NOT NULL CHECK (kind IN ('webview', 'private_video')),
    asset_url TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'ready', 'failed')),
    meta_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_playback_assets_item_id ON playback_assets(item_id);
CREATE INDEX idx_playback_assets_status ON playback_assets(status);
