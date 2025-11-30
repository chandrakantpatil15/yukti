import React, { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import api from '../../services/api';
import Loading from '../Loading';

interface OnboardingGuardProps {
  children: React.ReactNode;
}

export const OnboardingGuard: React.FC<OnboardingGuardProps> = ({ children }) => {
  const [loading, setLoading] = useState(true);
  const [hasOnboarded, setHasOnboarded] = useState(false);

  useEffect(() => {
    checkOnboardingStatus();
  }, []);

  const checkOnboardingStatus = async () => {
    try {
      console.log('[OnboardingGuard] Checking onboarding status...');
      const response = await api.getAWSConnection();
      console.log('[OnboardingGuard] Response:', response);
      
      if (response.success && response.data) {
        // AWS connection exists - user has completed onboarding
        console.log('[OnboardingGuard] AWS connection found - user has onboarded');
        setHasOnboarded(true);
      } else {
        // No AWS connection - user needs to complete onboarding
        console.log('[OnboardingGuard] No AWS connection - redirecting to onboarding');
        setHasOnboarded(false);
      }
    } catch (error) {
      // Error or no connection - redirect to onboarding
      console.error('[OnboardingGuard] Error checking status:', error);
      setHasOnboarded(false);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <Loading message="Checking onboarding status..." />;
  }

  if (!hasOnboarded) {
    // Redirect to onboarding if not completed
    return <Navigate to="/onboarding" replace />;
  }

  // User has completed onboarding - allow access
  return <>{children}</>;
};
