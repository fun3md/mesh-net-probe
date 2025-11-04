import axios, { AxiosResponse } from 'axios';
import type {
  Probe,
  Measurement,
  Configuration,
  User,
  LoginRequest,
  PaginatedResponse,
  DashboardStats,
  Alert,
  AuthToken,
  HealthStatus,
  MeasurementStatistics
} from '@/types';

// Extend the ImportMeta interface to include env
declare global {
  interface ImportMeta {
    env: Record<string, string>;
  }
}

class ApiService {
  private client: any;
  private baseURL: string;

  constructor() {
    this.baseURL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
    
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.client.interceptors.request.use((config: any) => {
      const token = localStorage.getItem('auth_token');
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Add response interceptor for error handling
    this.client.interceptors.response.use(
      (response: AxiosResponse) => response,
      (error: any) => {
        if (error.response?.status === 401) {
          // Token expired or invalid
          localStorage.removeItem('auth_token');
          localStorage.removeItem('user');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Authentication endpoints
  async login(credentials: LoginRequest): Promise<any> {
    const response: AxiosResponse<any> = await this.client.post('/auth/login', credentials);
    return response.data;
  }

  async logout(): Promise<void> {
    await this.client.post('/auth/logout');
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user');
  }

  async getCurrentUser(): Promise<User> {
    const response: AxiosResponse<User> = await this.client.get('/auth/me');
    return response.data;
  }

  async refreshToken(): Promise<AuthToken> {
    const response: AxiosResponse<AuthToken> = await this.client.post('/auth/refresh');
    return response.data;
  }

  // Configuration endpoints
  async getConfigurations(): Promise<Configuration[]> {
    const response: AxiosResponse<Configuration[]> = await this.client.get('/config');
    return response.data;
  }

  async getConfiguration(id: string): Promise<Configuration> {
    const response: AxiosResponse<Configuration> = await this.client.get(`/config/${id}`);
    return response.data;
  }

  async createConfiguration(config: Omit<Configuration, 'id' | 'createdAt' | 'updatedAt'>): Promise<Configuration> {
    const response: AxiosResponse<Configuration> = await this.client.post('/config', config);
    return response.data;
  }

  async updateConfiguration(id: string, config: Partial<Configuration>): Promise<Configuration> {
    const response: AxiosResponse<Configuration> = await this.client.put(`/config/${id}`, config);
    return response.data;
  }

  async deleteConfiguration(id: string): Promise<void> {
    await this.client.delete(`/config/${id}`);
  }

  async propagateConfiguration(id: string): Promise<void> {
    await this.client.post(`/config/${id}/propagate`);
  }

  // Probe endpoints
  async getProbes(): Promise<Probe[]> {
    const response: AxiosResponse<Probe[]> = await this.client.get('/probes');
    return response.data;
  }

  async getProbe(id: string): Promise<Probe> {
    const response: AxiosResponse<Probe> = await this.client.get(`/probes/${id}`);
    return response.data;
  }

  async registerProbe(probe: Omit<Probe, 'id' | 'createdAt' | 'updatedAt' | 'lastSeen'>): Promise<Probe> {
    const response: AxiosResponse<Probe> = await this.client.post('/probes', probe);
    return response.data;
  }

  async updateProbe(id: string, probe: Partial<Probe>): Promise<Probe> {
    const response: AxiosResponse<Probe> = await this.client.put(`/probes/${id}`, probe);
    return response.data;
  }

  async unregisterProbe(id: string): Promise<void> {
    await this.client.delete(`/probes/${id}`);
  }

  async sendHeartbeat(id: string): Promise<void> {
    await this.client.post(`/probes/${id}/heartbeat`);
  }

  async getProbeHealth(id: string): Promise<HealthStatus> {
    const response: AxiosResponse<HealthStatus> = await this.client.get(`/probes/${id}/health`);
    return response.data;
  }

  // Measurement endpoints
  async getMeasurements(params?: {
    probeId?: string;
    target?: string;
    limit?: number;
    offset?: number;
  }): Promise<PaginatedResponse<Measurement>> {
    const response: AxiosResponse<PaginatedResponse<Measurement>> = await this.client.get('/measurements', { params });
    return response.data;
  }

  async getMeasurement(id: string): Promise<Measurement> {
    const response: AxiosResponse<Measurement> = await this.client.get(`/measurements/${id}`);
    return response.data;
  }

  async createMeasurement(measurement: Omit<Measurement, 'id' | 'timestamp'>): Promise<Measurement> {
    const response: AxiosResponse<Measurement> = await this.client.post('/measurements', measurement);
    return response.data;
  }

  async getMeasurementStatistics(probeId: string, target: string, duration?: string): Promise<MeasurementStatistics> {
    const params = { probeId, target, duration };
    const response: AxiosResponse<MeasurementStatistics> = await this.client.get('/measurements/statistics', { params });
    return response.data;
  }

  // Monitoring endpoints
  async getDashboardStats(): Promise<DashboardStats> {
    const response: AxiosResponse<DashboardStats> = await this.client.get('/monitoring/dashboard');
    return response.data;
  }

  async getHealthSummary(): Promise<any> {
    const response: AxiosResponse<any> = await this.client.get('/monitoring/health/summary');
    return response.data;
  }

  async getMonitoringStats(): Promise<any> {
    const response: AxiosResponse<any> = await this.client.get('/monitoring/stats');
    return response.data;
  }

  // Alert endpoints
  async createAlert(alert: Omit<Alert, 'id' | 'timestamp' | 'acknowledged' | 'resolved'>): Promise<Alert> {
    const response: AxiosResponse<Alert> = await this.client.post('/monitoring/alerts', alert);
    return response.data;
  }

  async getAlerts(): Promise<Alert[]> {
    const response: AxiosResponse<Alert[]> = await this.client.get('/monitoring/alerts');
    return response.data;
  }

  // Health check
  async healthCheck(): Promise<{ status: string; time: string }> {
    const response: AxiosResponse<{ status: string; time: string }> = await this.client.get('/health');
    return response.data;
  }

  // Utility methods
  setAuthToken(token: string): void {
    localStorage.setItem('auth_token', token);
    this.client.defaults.headers.Authorization = `Bearer ${token}`;
  }

  clearAuthToken(): void {
    localStorage.removeItem('auth_token');
    delete this.client.defaults.headers.Authorization;
  }

  isAuthenticated(): boolean {
    return !!localStorage.getItem('auth_token');
  }
}

export const apiService = new ApiService();
export default apiService;