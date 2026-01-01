import React from 'react';
import { useRole } from '../../contexts/RoleContext';

interface RoleGuardProps {
  allowedRoles: string[];
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * RoleGuard component for conditionally rendering UI based on user role
 * This is different from ProtectedRoute which redirects - RoleGuard just hides/shows UI elements
 */
export const RoleGuard: React.FC<RoleGuardProps> = ({
  allowedRoles,
  children,
  fallback = null,
}) => {
  const { role } = useRole();

  // Check if current role is in allowed roles (case-insensitive)
  const isAllowed = allowedRoles.some(
    (allowedRole) => allowedRole.toLowerCase() === role.toLowerCase()
  );

  if (isAllowed) {
    return <>{children}</>;
  }

  return <>{fallback}</>;
};

