// Auth utilities for JWT token management and user context

export interface User {
  id: string;
  email: string;
  role: 'admin' | 'editor' | 'viewer';
  tenant_id: number;
}

export interface JWTPayload {
  user_id: string;
  tenant_id: number;
  tenant_code: string;
  email: string;
  role: string;
  iat: number;
  exp: number;
}

const TOKEN_KEY = 'yukti_auth_token';
const USER_KEY = 'yukti_user';

/**
 * Decode JWT token (client-side only, no verification)
 * In production, token verification should be done server-side
 */
function decodeJWT(token: string): JWTPayload | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      return null;
    }
    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decoded);
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return null;
  }
}

/**
 * Check if token is expired
 */
function isTokenExpired(token: string): boolean {
  const payload = decodeJWT(token);
  if (!payload || !payload.exp) {
    return true;
  }
  return Date.now() >= payload.exp * 1000;
}

/**
 * Get stored auth token
 */
export function getToken(): string | null {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) {
    return null;
  }
  
  // Check if expired
  if (isTokenExpired(token)) {
    removeToken();
    return null;
  }
  
  return token;
}

/**
 * Store auth token
 * TODO: Migrate to HttpOnly cookie via backend Set-Cookie header
 */
export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
  
  // Also decode and store user info for quick access
  const payload = decodeJWT(token);
  if (payload) {
    const user: User = {
      id: payload.user_id,
      email: payload.email,
      role: payload.role as 'admin' | 'editor' | 'viewer',
      tenant_id: payload.tenant_id,
    };
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

/**
 * Remove auth token
 */
export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

/**
 * Get current user from stored token
 */
export function getCurrentUser(): User | null {
  const userStr = localStorage.getItem(USER_KEY);
  if (!userStr) {
    return null;
  }
  
  try {
    return JSON.parse(userStr);
  } catch {
    return null;
  }
}

/**
 * Check if user has required role
 */
export function hasRole(user: User | null, allowedRoles: string[]): boolean {
  if (!user) {
    return false;
  }
  return allowedRoles.includes(user.role);
}

/**
 * Get Authorization header value
 */
export function getAuthHeader(): string | null {
  const token = getToken();
  return token ? `Bearer ${token}` : null;
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  return getToken() !== null;
}

