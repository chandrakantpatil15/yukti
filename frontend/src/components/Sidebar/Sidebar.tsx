import React from 'react';
import { 
  LayoutDashboard, 
  Box, 
  DollarSign, 
  Shield, 
  List,
  Settings,
  User,
  LogOut,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';
import { getCurrentUser } from '../../lib/auth';
import LogoutButton from '../LogoutButton';

interface SidebarProps {
  currentPath: string;
  onNavigate: (path: string) => void;
  collapsed: boolean;
  onToggleCollapse: () => void;
}

const navItems = [
  { path: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { path: 'hidden-costs', label: 'Hidden Costs', icon: DollarSign },
  { path: 'resources', label: 'Resources', icon: Box },
  { path: 'cost-analytics', label: 'Cost Analytics', icon: DollarSign },
  { path: 'resource-utilization', label: 'Utilization', icon: Box },
  { path: 'whitelists', label: 'Whitelists', icon: Shield },
  { path: 'profile', label: 'Profile', icon: User },
  { path: 'onboarding', label: 'Settings', icon: Settings },
  { path: 'audit-logs', label: 'Audit Logs', icon: List, roles: ['admin'] },
];

export const Sidebar: React.FC<SidebarProps> = ({ 
  currentPath, 
  onNavigate, 
  collapsed, 
  onToggleCollapse 
}) => {
  const user = getCurrentUser();
  
  const filteredNavItems = navItems.filter(item => {
    if (!item.roles) return true;
    return item.roles.includes(user?.role || 'user');
  });

  return (
    <div className={`bg-gray-900 text-white transition-all duration-300 ${
      collapsed ? 'w-16' : 'w-64'
    } min-h-screen flex flex-col`}>
      {/* Header */}
      <div className="p-4 border-b border-gray-700">
        <div className="flex items-center justify-between">
          {!collapsed && (
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">Y</span>
              </div>
              <span className="font-bold text-lg">Yukti FinOps</span>
            </div>
          )}
          <button
            onClick={onToggleCollapse}
            className="p-1 rounded hover:bg-gray-700 transition-colors"
          >
            {collapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
          </button>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4">
        <div className="space-y-2">
          {filteredNavItems.map((item) => {
            const Icon = item.icon;
            const isActive = currentPath === item.path;
            
            return (
              <button
                key={item.path}
                onClick={() => onNavigate(item.path)}
                className={`w-full flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors ${
                  isActive 
                    ? 'bg-blue-600 text-white' 
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'
                }`}
                title={collapsed ? item.label : undefined}
              >
                <Icon className="w-5 h-5 flex-shrink-0" />
                {!collapsed && <span className="text-sm font-medium">{item.label}</span>}
              </button>
            );
          })}
        </div>
      </nav>

      {/* User Section */}
      <div className="p-4 border-t border-gray-700">
        {!collapsed ? (
          <div className="space-y-3">
            <div className="flex items-center space-x-3 px-3 py-2">
              <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center">
                <User className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-white truncate">{user?.email}</p>
                <p className="text-xs text-gray-400 capitalize">{user?.role || 'Member'}</p>
              </div>
            </div>
            <LogoutButton className="w-full flex items-center justify-center space-x-2 px-3 py-2 text-sm text-gray-300 hover:bg-gray-700 hover:text-white rounded-lg transition-colors">
              <LogOut className="w-4 h-4" />
              <span>Logout</span>
            </LogoutButton>
          </div>
        ) : (
          <div className="space-y-2">
            <button
              onClick={() => onNavigate('profile')}
              className="w-full flex items-center justify-center p-2 text-gray-300 hover:bg-gray-700 hover:text-white rounded-lg transition-colors"
              title="Profile"
            >
              <User className="w-5 h-5" />
            </button>
            <LogoutButton className="w-full flex items-center justify-center p-2 text-gray-300 hover:bg-gray-700 hover:text-white rounded-lg transition-colors">
              <LogOut className="w-4 h-4" />
            </LogoutButton>
          </div>
        )}
      </div>
    </div>
  );
};