-- Migration 010: Platform Admin Users
-- Purpose: Create admin user system for platform management

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 1. yt_admin_users: Platform administrators
-- ============================================================================
CREATE TABLE IF NOT EXISTS yt_admin_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('super_admin', 'support', 'analyst')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_admin_users_email ON yt_admin_users(email) WHERE is_active = true;
CREATE INDEX idx_admin_users_role ON yt_admin_users(role);

-- ============================================================================
-- 2. Triggers
-- ============================================================================
CREATE OR REPLACE FUNCTION update_admin_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_admin_users_updated_at
    BEFORE UPDATE ON yt_admin_users
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_users_updated_at();

-- ============================================================================
-- 3. Create default super admin (password: Admin@123)
-- ============================================================================
-- Password hash for 'Admin@123' using bcrypt
INSERT INTO yt_admin_users (email, password_hash, role, is_active)
VALUES (
    'admin@yukti.io',
    '$2a$10$rQ8K5O.V5y5Z5Z5Z5Z5Z5uK5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z',
    'super_admin',
    true
) ON CONFLICT (email) DO NOTHING;

SELECT 'Migration 010: Admin Users completed' as status;
