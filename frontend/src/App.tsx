import React, { useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from 'react-query';
import { ReactQueryDevtools } from 'react-query/devtools';
import Dashboard from './pages/Dashboard';
import HiddenCosts from './pages/HiddenCosts';
import { Whitelists } from './pages/Whitelists';
import Resources from './pages/Resources';
import Onboarding from './pages/Onboarding';
import SimpleOnboarding from './pages/SimpleOnboarding';
import AdminDashboard from './pages/AdminDashboard';
import TestAdmin from './pages/TestAdmin';
import AuditLogs from './pages/AuditLogs';
import Login from './pages/Auth/Login';
import Signup from './pages/Auth/Signup';
import LogoutButton from './components/LogoutButton';
import ErrorBoundary from './components/ErrorBoundary';
import { Navigation } from './components/Navigation/Navigation';
import { ThemeProvider } from './context/ThemeContext';
import { ThemeToggle } from './components/ThemeToggle';
import { ProtectedRoute } from './components/Auth/ProtectedRoute';
import { isAuthenticated } from './lib/auth';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      refetchOnWindowFocus: false,
    },
  },
});

// 403 Forbidden page component
const ForbiddenPage: React.FC = () => (
  <div className="min-h-screen flex items-center justify-center bg-gray-50">
    <div className="text-center">
      <h1 className="text-4xl font-bold text-gray-900 mb-4">403</h1>
      <p className="text-xl text-gray-600 mb-8">Forbidden - Insufficient permissions</p>
      <a href="/dashboard" className="text-indigo-600 hover:text-indigo-500">
        Return to Dashboard
      </a>
    </div>
  </div>
);

function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <Routes>
              {/* Public routes - redirect if already authenticated */}
              <Route
                path="/login"
                element={isAuthenticated() ? <Navigate to="/dashboard" replace /> : <Login />}
              />
              <Route
                path="/signup"
                element={isAuthenticated() ? <Navigate to="/dashboard" replace /> : <Signup />}
              />
              <Route path="/403" element={<ForbiddenPage />} />

              {/* Protected routes */}
              <Route
                path="/"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <Dashboard />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/dashboard"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <Dashboard />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/hidden-costs"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <HiddenCosts />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/resources"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <Resources />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/whitelists"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <Whitelists />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/onboarding"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <SimpleOnboarding />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/admin"
                element={
                  <ProtectedRoute allowedRoles={['admin']}>
                    <AppLayout>
                      <AdminDashboard />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/audit-logs"
                element={
                  <ProtectedRoute allowedRoles={['admin']}>
                    <AppLayout>
                      <AuditLogs />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />

              {/* Default redirect */}
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Routes>
            <ReactQueryDevtools initialIsOpen={false} />
          </BrowserRouter>
        </QueryClientProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

// AppLayout component for protected routes
interface AppLayoutProps {
  children: React.ReactNode;
}

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const getPageFromPath = () => {
    const path = window.location.pathname;
    if (path.includes('/audit-logs')) return 'audit-logs';
    if (path.includes('/admin')) return 'admin';
    if (path.includes('/onboarding')) return 'onboarding';
    if (path.includes('/hidden-costs')) return 'hidden-costs';
    if (path.includes('/resources')) return 'resources';
    if (path.includes('/whitelists')) return 'whitelists';
    if (path.includes('/iac-generator')) return 'iac-generator';
    if (path.includes('/dashboard') || path === '/') return 'dashboard';
    return 'dashboard';
  };

  const [currentPage, setCurrentPage] = useState(getPageFromPath());

  React.useEffect(() => {
    const handlePopState = () => {
      setCurrentPage(getPageFromPath());
    };
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  return (
    <div className="min-h-screen bg-neutral-100 dark:bg-neutral-900">
      <div className="sticky top-0 z-10">
        <Navigation 
          currentPath={currentPage} 
          onNavigate={(page) => {
            setCurrentPage(page);
            window.history.pushState({}, '', `/${page === 'dashboard' ? '' : page}`);
          }} 
        />
        <div className="absolute right-4 top-4">
          <ThemeToggle />
        </div>
      </div>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  );
}

export default App;