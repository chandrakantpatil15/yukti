-- Migration 008: Multi-User RBAC System
-- Purpose: Enable multiple users per tenant with role-based access control

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 1. yt_tenant_users: Maps users to tenants with roles
-- ============================================================================
CREATE TABLE IF NOT EXISTS yt_tenant_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES yt_users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(50) NOT NULL REFERENCES yt_customers(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    invited_by UUID REFERENCES yt_users(id) ON DELETE SET NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_tenant UNIQUE (user_id, tenant_id)
);

CREATE INDEX idx_tenant_users_user ON yt_tenant_users(user_id);
CREATE INDEX idx_tenant_users_tenant ON yt_tenant_users(tenant_id);
CREATE INDEX idx_tenant_users_role ON yt_tenant_users(tenant_id, role);

-- ============================================================================
-- 2. yt_user_invitations: Pending invitations
-- ============================================================================
CREATE TABLE IF NOT EXISTS yt_user_invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(50) NOT NULL REFERENCES yt_customers(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'editor', 'viewer')),
    invitation_token TEXT NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES yt_users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_invitations_tenant ON yt_user_invitations(tenant_id);
CREATE INDEX idx_invitations_email ON yt_user_invitations(email);
CREATE INDEX idx_invitations_token ON yt_user_invitations(invitation_token);
CREATE INDEX idx_invitations_status ON yt_user_invitations(tenant_id, status);

-- ============================================================================
-- 3. Update yt_users for multi-tenant support
-- ============================================================================
ALTER TABLE yt_users 
    ADD COLUMN IF NOT EXISTS first_name TEXT,
    ADD COLUMN IF NOT EXISTS last_name TEXT,
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS last_login_ip TEXT;

ALTER TABLE yt_users ALTER COLUMN tenant_id DROP NOT NULL;

-- ============================================================================
-- 4. Triggers
-- ============================================================================
CREATE OR REPLACE FUNCTION update_tenant_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_tenant_users_updated_at
    BEFORE UPDATE ON yt_tenant_users
    FOR EACH ROW
    EXECUTE FUNCTION update_tenant_users_updated_at();

CREATE OR REPLACE FUNCTION update_invitations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_invitations_updated_at
    BEFORE UPDATE ON yt_user_invitations
    FOR EACH ROW
    EXECUTE FUNCTION update_invitations_updated_at();

-- ============================================================================
-- 5. Migrate existing users as owners
-- ============================================================================
INSERT INTO yt_tenant_users (user_id, tenant_id, role, is_active, joined_at)
SELECT u.id, CAST(u.tenant_id AS VARCHAR(50)), 'owner', u.is_active, u.created_at
FROM yt_users u
WHERE u.tenant_id IS NOT NULL
  AND EXISTS (SELECT 1 FROM yt_customers c WHERE c.id = CAST(u.tenant_id AS VARCHAR(50)))
ON CONFLICT (user_id, tenant_id) DO NOTHING;

-- ============================================================================
-- 6. Helper views
-- ============================================================================
CREATE OR REPLACE VIEW v_user_tenants AS
SELECT 
    u.id as user_id,
    u.email,
    u.first_name,
    u.last_name,
    tu.tenant_id,
    c.company_name as tenant_name,
    tu.role,
    tu.is_active,
    tu.joined_at
FROM yt_users u
JOIN yt_tenant_users tu ON u.id = tu.user_id
JOIN yt_customers c ON tu.tenant_id = c.id
WHERE u.is_active = true AND tu.is_active = true;

CREATE OR REPLACE VIEW v_tenant_members AS
SELECT 
    tu.tenant_id,
    c.company_name as tenant_name,
    u.id as user_id,
    u.email,
    u.first_name,
    u.last_name,
    tu.role,
    tu.is_active,
    tu.joined_at,
    u.last_login_at
FROM yt_tenant_users tu
JOIN yt_users u ON tu.user_id = u.id
JOIN yt_customers c ON tu.tenant_id = c.id
ORDER BY tu.tenant_id, tu.role, u.email;

SELECT 'Migration 008: Multi-User RBAC completed' as status;
