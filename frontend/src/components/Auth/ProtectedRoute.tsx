import React from 'react';
import { Navigate } from 'react-router-dom';
import { getCurrentUser, hasRole, isAuthenticated } from '../../lib/auth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  allowedRoles?: string[];
  redirectTo?: string;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  allowedRoles = [],
  redirectTo = '/login',
}) => {
  if (!isAuthenticated()) {
    return <Navigate to={redirectTo} replace />;
  }

  const user = getCurrentUser();
  
  if (allowedRoles.length > 0 && !hasRole(user, allowedRoles)) {
    return <Navigate to="/403" replace />;
  }

  return <>{children}</>;
};

export default ProtectedRoute;

