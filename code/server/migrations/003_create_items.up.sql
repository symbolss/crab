-- Create items table
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    family_id UUID NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    title VARCHAR(500),
    cover_url VARCHAR(1000),
    summary TEXT,
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('douyin', 'x', 'wechat_channels', 'web_article')),
    source_url TEXT NOT NULL,
    normalized_url TEXT NOT NULL,
    playback_mode VARCHAR(50) DEFAULT 'webview' CHECK (playback_mode IN ('webview', 'private_video', 'article', 'audio')),
    processing_status VARCHAR(50) DEFAULT 'pending' CHECK (processing_status IN ('pending', 'ready', 'failed')),
    playback_status VARCHAR(50) DEFAULT 'unknown' CHECK (playback_status IN ('unknown', 'playable', 'fallback_required')),
    age_band VARCHAR(50),
    duration_sec INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_items_family_id ON items(family_id);
CREATE INDEX idx_items_normalized_url ON items(normalized_url);
CREATE INDEX idx_items_processing_status ON items(processing_status);
