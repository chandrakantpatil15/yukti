-- Migration: Add AWS pricing sync status table
CREATE TABLE IF NOT EXISTS yt_aws_pricing_sync_status (
    id SERIAL PRIMARY KEY,
    sync_start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    sync_end_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL, -- 'running', 'completed', 'failed'
    total_records INTEGER DEFAULT 0,
    processed_records INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    last_service VARCHAR(100),
    last_region VARCHAR(50),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for quick lookup of latest sync status
CREATE INDEX IF NOT EXISTS idx_aws_pricing_sync_status_time 
ON yt_aws_pricing_sync_status(sync_start_time DESC);