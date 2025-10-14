import axios from 'axios';
import { CostSummary, Recommendation, Resource } from '../types';

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
});

export const costApi = {
  getSummary: (): Promise<CostSummary> =>
    api.get('/cost/summary').then(res => res.data),
    
  getRecommendations: (): Promise<Recommendation[]> =>
    api.get('/cost/recommendations').then(res => res.data),
    
  getResources: (): Promise<Resource[]> =>
    api.get('/cost/resources').then(res => res.data),
};