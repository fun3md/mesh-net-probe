import axios, { AxiosResponse } from 'axios';
import type {
  Probe,
  Measurement,
  Configuration,
  User,
  LoginRequest,
  DashboardStats,
  Alert,
  HealthStatus
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
    this.baseURL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8080/api';
    
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

  // Backend: POST /auth/login → { token: string; user: { id, username, role }, expiresAt: number }
  async login(credentials: LoginRequest): Promise<{ token: string; user: User; expiresAt: number }> {
    const response: AxiosResponse<{ token: string; user: User; expiresAt: number }> =
      await this.client.post('/auth/login', credentials);
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

  // Backend: POST /auth/refresh → { token: string; expiresAt: number }
  async refreshToken(): Promise<{ token: string; expiresAt: number }> {
    const response: AxiosResponse<{ token: string; expiresAt: number }> =
      await this.client.post('/auth/refresh');
    return response.data;
  }

  // Configuration endpoints

  // Backend: GET /config → single active configuration (types.Configuration)
  async getConfig(): Promise<Configuration> {
    const response: AxiosResponse<Configuration> = await this.client.get('/config');
    return response.data;
  }

  // Backend: GET /config/status → manager status object
  async getConfigStatus(): Promise<any> {
    const response: AxiosResponse<any> = await this.client.get('/config/status');
    return response.data;
  }

  // NOTE: Backend does not currently expose GET /config/:id - callers should use getConfig() only.
  // Keeping getConfiguration for compatibility, but delegate to getConfig.
  async getConfiguration(_id: string): Promise<Configuration> {
    return this.getConfig();
  }

  async createConfiguration(
    config: Omit<Configuration, 'id' | 'createdAt' | 'updatedAt'>
  ): Promise<Configuration> {
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

  // Backend: POST /config/propagate (no id in path)
  async propagateConfiguration(): Promise<{ message: string; status?: any }> {
    const response: AxiosResponse<{ message: string; status?: any }> =
      await this.client.post('/config/propagate');
    return response.data;
  }

  // Probe endpoints

  // Backend: GET /probes → { probes: Probe[] }
  async getProbes(): Promise<Probe[]> {
    const response: AxiosResponse<{ probes: any[] }> = await this.client.get('/probes');
    const probes = response.data.probes ?? [];

    // Normalize snake_case fields from backend into Probe type expectations.
    return probes.map((p: any) => ({
      id: p.id,
      name: p.name,
      version: p.version,
      platform: p.platform,
      arch: p.arch,
      ipAddress: p.ip_address ?? p.ipAddress,
      tags: p.tags ?? [],
      metadata: p.metadata ?? {},
      status: p.status ?? 'unknown',
      lastSeen: p.lastSeen ?? p.last_seen ?? '',
      health: p.health,
      createdAt: p.createdAt ?? '',
      updatedAt: p.updatedAt ?? '',
    })) as Probe[];
  }

  async getProbe(id: string): Promise<Probe> {
    const response: AxiosResponse<any> = await this.client.get(`/probes/${id}`);
    const p = response.data;
    return {
      id: p.id,
      name: p.name,
      version: p.version,
      platform: p.platform,
      arch: p.arch,
      ipAddress: p.ip_address ?? p.ipAddress,
      tags: p.tags ?? [],
      metadata: p.metadata ?? {},
      status: p.status ?? 'unknown',
      lastSeen: p.lastSeen ?? p.last_seen,
      health: p.health,
      createdAt: p.createdAt ?? '',
      updatedAt: p.updatedAt ?? '',
    } as Probe;
  }

  async registerProbe(
    probe: Omit<Probe, 'id' | 'createdAt' | 'updatedAt' | 'lastSeen'>
  ): Promise<Probe> {
    const response: AxiosResponse<{ probe: Probe }> = await this.client.post('/probes', probe);
    // Backend returns { message, probe }
    return (response.data as any).probe ?? (response.data as any);
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
    const response: AxiosResponse<any> = await this.client.get(`/probes/${id}/health`);
    // Backend shape: { probe_id, health, status, last_seen }
    // For now, pass through as-is; callers expecting HealthStatus should be adjusted accordingly in Phase 5.2.
    return response.data as HealthStatus;
  }

  // Measurement endpoints

  // Backend: GET /measurements → { measurements: Measurement[] }
  async getMeasurements(params?: {
    probeId?: string;
    target?: string;
    limit?: number;
    offset?: number;
  }): Promise<Measurement[]> {
    const response: AxiosResponse<{ measurements: Measurement[] }> =
      await this.client.get('/measurements', { params });
    return response.data.measurements ?? [];
  }

  async getMeasurement(id: string): Promise<Measurement> {
    const response: AxiosResponse<Measurement> = await this.client.get(`/measurements/${id}`);
    return response.data;
  }

  async createMeasurement(
    measurement: Omit<Measurement, 'id' | 'timestamp'>
  ): Promise<Measurement> {
    const response: AxiosResponse<Measurement> = await this.client.post(
      '/measurements',
      measurement
    );
    return response.data;
  }

  // Backend: GET /measurements/statistics → flat statistics object (demo)
  async getMeasurementStatistics(): Promise<any> {
    const response: AxiosResponse<any> = await this.client.get('/measurements/statistics');
    return response.data;
  }

  // Monitoring endpoints

  // Backend: GET /monitoring/dashboard → DashboardStats with snake_case; map to frontend shape here
  async getDashboardStats(): Promise<DashboardStats> {
    const response: AxiosResponse<any> = await this.client.get('/monitoring/dashboard');
    const d = response.data || {};
    const stats: DashboardStats = {
      totalProbes: d.total_probes ?? d.totalProbes ?? 0,
      onlineProbes: d.active_probes ?? d.activeProbes ?? 0,
      offlineProbes: 0,
      degradedProbes: 0,
      totalMeasurements: d.total_measurements ?? d.totalMeasurements ?? 0,
      averageLatency: d.average_rtt ?? d.averageLatency ?? 0,
      lastUpdate: d.last_updated ?? d.lastUpdate ?? new Date().toISOString(),
    };
    return stats;
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

  // Backend: POST /monitoring/alerts → Alert
  async createAlert(
    alert: Omit<Alert, 'id' | 'timestamp' | 'acknowledged' | 'resolved'>
  ): Promise<Alert> {
    const response: AxiosResponse<any> = await this.client.post('/monitoring/alerts', alert);
    const a = response.data;
    return {
      id: a.id,
      type: a.type ?? 'alert',
      severity: a.severity,
      title: a.title,
      message: a.description ?? a.message,
      probeId: a.source,
      timestamp: a.created_at ?? a.timestamp ?? new Date().toISOString(),
      acknowledged: a.status === 'acknowledged',
      resolved: a.status === 'resolved' || !!a.resolved_at,
    } as Alert;
  }

  // Backend: GET /monitoring/alerts → { alerts: Alert[] }
  async getAlerts(): Promise<Alert[]> {
    const response: AxiosResponse<{ alerts: any[] }> =
      await this.client.get('/monitoring/alerts');
    const alerts = response.data.alerts ?? [];
    return alerts.map(a => ({
      id: a.id,
      type: a.type ?? 'alert',
      severity: a.severity,
      title: a.title,
      message: a.description ?? a.message,
      probeId: a.source,
      timestamp: a.created_at ?? a.timestamp ?? new Date().toISOString(),
      acknowledged: a.status === 'acknowledged',
      resolved: a.status === 'resolved' || !!a.resolved_at,
    })) as Alert[];
  }

  // Health check
  // NOTE: Backend does not expose /health in routes.go; keep this as optional/experimental.
  async healthCheck(): Promise<{ status: string; time: string }> {
    const response: AxiosResponse<{ status: string; time: string }> =
      await this.client.get('/health');
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