import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8085';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const apiService = {
  // Health check
  getHealth: async () => {
    const response = await api.get('/health');
    return response.data.data || response.data;
  },

  // Resources
  getResources: async () => {
    const response = await api.get('/api/v1/resources');
    return response.data.data || response.data;
  },

  // Assessments
  getAssessments: async () => {
    const response = await api.get('/api/v1/assessments');
    return response.data;
  },

  runAssessment: async () => {
    const response = await api.post('/api/v1/assessments/run');
    return response.data;
  },

  // Cost analysis
  getCostAnalysis: async () => {
    const response = await api.get('/api/v1/cost-analysis');
    return response.data;
  },

  // Emergency actions
  emergencyStop: async () => {
    const response = await api.post('/api/v1/emergency-stop');
    return response.data;
  },

  // Kill switch
  getKillSwitch: async () => {
    const response = await api.get('/api/v1/kill-switch');
    return response.data;
  },

  setKillSwitch: async (enabled, reason = '') => {
    const response = await api.post('/api/v1/kill-switch', {
      enable: enabled,
      reason: reason
    });
    return response.data;
  },

  // Configuration
  getConfig: async () => {
    const response = await api.get('/api/v1/config');
    return response.data;
  },

  updateConfig: async (config) => {
    const response = await api.put('/api/v1/config', config);
    return response.data;
  }
};

// Request interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error);
    
    if (error.response?.status === 404) {
      console.warn('API endpoint not found:', error.config.url);
      // Return mock data for development
      return Promise.resolve({ data: [] });
    }
    
    return Promise.reject(error);
  }
);

export default api;