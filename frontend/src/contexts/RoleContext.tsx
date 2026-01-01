import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { getCurrentUser } from '../lib/auth';
import api from '../services/api';

interface RoleContextType {
  role: string;
  hasPermission: (permission: string) => boolean;
  canManageUser: (targetRole: string) => boolean;
  refreshRole: () => void;
}

const RoleContext = createContext<RoleContextType | null>(null);

// Permission matrix based on backend models/permissions.go
const ROLE_PERMISSIONS: Record<string, string[]> = {
  owner: [
    'view_aws',
    'manage_aws',
    'scan_resources',
    'view_findings',
    'manage_findings',
    'view_whitelists',
    'manage_whitelists',
    'view_budgets',
    'manage_budgets',
    'generate_iac',
    'view_team',
    'manage_team',
    'view_billing',
    'manage_billing',
  ],
  admin: [
    'view_aws',
    'manage_aws',
    'scan_resources',
    'view_findings',
    'manage_findings',
    'view_whitelists',
    'manage_whitelists',
    'view_budgets',
    'manage_budgets',
    'generate_iac',
    'view_team',
    'manage_team',
    'view_billing', // Can view but not manage billing
  ],
  editor: [
    'view_aws',
    'scan_resources',
    'view_findings',
    'manage_findings',
    'view_whitelists',
    'manage_whitelists',
    'view_budgets',
    'generate_iac',
    'view_team',
  ],
  viewer: ['view_aws', 'view_findings', 'view_whitelists', 'view_budgets', 'view_team'],
};

interface RoleProviderProps {
  children: ReactNode;
}

export const RoleProvider: React.FC<RoleProviderProps> = ({ children }) => {
  const [role, setRole] = useState<string>('viewer');

  const refreshRole = () => {
    const user = getCurrentUser();
    if (user?.role) {
      setRole(user.role.toLowerCase());
    }
  };

  useEffect(() => {
    refreshRole();
  }, []);

  const hasPermission = (permission: string): boolean => {
    const permissions = ROLE_PERMISSIONS[role.toLowerCase()] || [];
    return permissions.includes(permission);
  };

  const canManageUser = (targetRole: string): boolean => {
    const current = role.toLowerCase();
    const target = targetRole.toLowerCase();

    // Owner can manage everyone except other owners
    if (current === 'owner' && target !== 'owner') return true;
    // Admin can manage editors and viewers
    if (current === 'admin' && (target === 'editor' || target === 'viewer')) return true;
    return false;
  };

  return (
    <RoleContext.Provider value={{ role, hasPermission, canManageUser, refreshRole }}>
      {children}
    </RoleContext.Provider>
  );
};

export const useRole = (): RoleContextType => {
  const context = useContext(RoleContext);
  if (!context) {
    throw new Error('useRole must be used within a RoleProvider');
  }
  return context;
};

