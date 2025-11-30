-- Refresh Tokens Table
CREATE TABLE IF NOT EXISTS yt_refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    tenant_id INTEGER NOT NULL,
    token VARCHAR(512) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP,
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    is_revoked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES yt_users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES yt_tenants(id) ON DELETE CASCADE
);

-- Token Blacklist Table
CREATE TABLE IF NOT EXISTS yt_token_blacklist (
    id SERIAL PRIMARY KEY,
    token_jti VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    tenant_id INTEGER NOT NULL,
    revoked_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    reason VARCHAR(255),
    FOREIGN KEY (user_id) REFERENCES yt_users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES yt_tenants(id) ON DELETE CASCADE
);

-- Session Activity Log
CREATE TABLE IF NOT EXISTS yt_session_activity (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    tenant_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES yt_users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES yt_tenants(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_refresh_tokens_user ON yt_refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON yt_refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires ON yt_refresh_tokens(expires_at);
CREATE INDEX idx_blacklist_jti ON yt_token_blacklist(token_jti);
CREATE INDEX idx_blacklist_expires ON yt_token_blacklist(expires_at);
CREATE INDEX idx_session_activity_user ON yt_session_activity(user_id);
CREATE INDEX idx_session_activity_created ON yt_session_activity(created_at);
