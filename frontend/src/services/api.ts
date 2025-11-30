import { getAuthHeader, getCurrentUser } from '../lib/auth';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export class ApiService {
  private async request(endpoint: string, options: RequestInit = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    const authHeader = getAuthHeader();
    const user = getCurrentUser();

    const headers = {
      'Content-Type': 'application/json',
      ...(authHeader && { Authorization: authHeader }),
      // Keep X-Tenant-ID for backward compatibility with existing endpoints
      ...(user && { 'X-Tenant-ID': user.tenant_id.toString() }),
      ...options.headers,
    };

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: response.statusText }));
      throw new Error(error.message || `API Error: ${response.statusText}`);
    }

    return response.json();
  }

  // Resource endpoints
  async getResources(filters?: {
    region?: string;
    type?: string;
    tags?: Record<string, string>;
  }) {
    return this.request('/resources', {
      method: 'GET',
      headers: { 'X-Include-Filters': JSON.stringify(filters) },
    });
  }

  async getResourceDetails(resourceId: string) {
    return this.request(`/resources/${resourceId}`);
  }

  async getResourceHistory(resourceId: string, page = 1, limit = 10) {
    return this.request(`/resources/${resourceId}/history?page=${page}&limit=${limit}`);
  }

  async getResourceTags(resourceId: string) {
    return this.request(`/resources/${resourceId}/tags`);
  }

  async updateResourceTags(resourceId: string, tags: Record<string, string>) {
    return this.request(`/resources/${resourceId}/tags`, {
      method: 'PUT',
      body: JSON.stringify({ tags }),
    });
  }

  async generateIaC(params: {
    resource_id: string;
    resource_type: string;
    format: 'terraform' | 'cloudformation';
  }) {
    return this.request('/iac/generate', {
      method: 'POST',
      body: JSON.stringify(params),
    });
  }

  async get(endpoint: string, headers?: Record<string, string>) {
    return this.request(endpoint, { method: 'GET', headers });
  }

  async post(endpoint: string, data: any, headers?: Record<string, string>) {
    return this.request(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
      headers,
    });
  }

  async put(endpoint: string, data: any, headers?: Record<string, string>) {
    return this.request(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
      headers,
    });
  }

  async delete(endpoint: string, data?: any, headers?: Record<string, string>) {
    return this.request(endpoint, {
      method: 'DELETE',
      ...(data && { body: JSON.stringify(data) }),
      headers,
    });
  }
}

const api = new ApiService();
export default api;
