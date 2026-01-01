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

  // Cost Analytics APIs
  async getCostTrend(startDate: string, endDate: string, groupBy: string = 'service') {
    return this.request(`/analytics/cost-trend?start_date=${startDate}&end_date=${endDate}&group_by=${groupBy}`);
  }

  // Billing APIs
  async getBillingInfo() {
    return this.request('/billing/info');
  }

  async createCheckoutSession(planSlug: string, currency: string = 'USD') {
    return this.post('/billing/checkout-session', {
      plan_slug: planSlug,
      currency: currency,
    });
  }

  // Team Management APIs
  async getTeamMembers() {
    return this.request('/team/members');
  }

  async getTeamInvitations() {
    return this.request('/team/invitations');
  }

  async inviteUser(email: string, role: string) {
    return this.post('/team/invite', {
      email,
      role,
    });
  }

  async acceptInvite(token: string) {
    return this.post('/team/accept-invite', { token });
  }

  async getInviteDetails(token: string) {
    return this.get(`/team/invite-details?token=${token}`);
  }

  async updateMemberRole(memberId: string, role: string) {
    return this.put(`/team/members/${memberId}/role`, { role });
  }

  async removeMember(memberId: string) {
    return this.delete(`/team/members/${memberId}`);
  }

  async resendInvitation(invitationId: string) {
    return this.post(`/team/invitations/${invitationId}/resend`, {});
  }

  async revokeInvitation(invitationId: string) {
    return this.delete(`/team/invitations/${invitationId}`);
  }

  async getCostDrivers(startDate: string, endDate: string) {
    return this.request(`/analytics/cost-drivers?start_date=${startDate}&end_date=${endDate}`);
  }

  async getAnomalies(days: number = 30) {
    return this.request(`/analytics/anomalies?days=${days}`);
  }

  // Resource Utilization APIs
  async getUtilizationMetrics(timeRange: string = '7d') {
    return this.request(`/analytics/utilization?time_range=${timeRange}`);
  }

  async getIdleResources(days: number = 7) {
    return this.request(`/analytics/idle-resources?days=${days}`);
  }

  async getRightSizingRecommendations() {
    return this.request('/analytics/right-sizing');
  }

  // Enhanced Dashboard APIs
  async getDashboardMetrics() {
    return this.request('/dashboard/metrics');
  }

  async getCostBreakdown() {
    return this.request('/dashboard/cost-breakdown');
  }

  async getUtilizationSummary() {
    return this.request('/dashboard/utilization-summary');
  }

  // Auth endpoints
  async signup(email: string, password: string, companyName?: string) {
    return this.post('/auth/signup', { email, password, company_name: companyName });
  }

  async verifyEmail(email: string, code: string) {
    return this.post('/auth/verify-email', { email, code });
  }

  async resendCode(email: string) {
    return this.post('/auth/resend-code', { email });
  }
}

const api = new ApiService();
export default api;
