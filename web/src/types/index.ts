// Common types for the mesh probe system

export interface Probe {
  id: string;
  name: string;
  version: string;
  platform: string;
  arch: string;
  ipAddress: string;
  tags: string[];
  metadata: Record<string, any>;
  status: ProbeStatus;
  lastSeen: string;
  health?: HealthStatus;
  createdAt: string;
  updatedAt: string;
  // Optional configuration tracking fields as exposed by ProbeRegistry/admin-web API
  configId?: string;
  configVersion?: number;
  configSource?: string;
  configAppliedAt?: string;
}

export type ProbeStatus = 'online' | 'offline' | 'degraded' | 'unknown';

export interface HealthStatus {
  overallStatus: string;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  networkLatency: number;
  icmpSuccessRate: number;
  lastHealthCheck: string;
  checks: Record<string, HealthCheck>;
}

export interface HealthCheck {
  name: string;
  status: string;
  message: string;
  duration: number;
  lastChecked: string;
}

export interface Measurement {
  id: string;
  probeId: string;
  target: string;
  type: MeasurementType;
  status: MeasurementStatus;
  value: number;
  unit: string;
  timestamp: string;
  duration: number;
  metadata?: Record<string, any>;
  statistics?: MeasurementStatistics;
}

export type MeasurementType = 'icmp' | 'latency' | 'jitter' | 'packet_loss';

export type MeasurementStatus = 'success' | 'failed' | 'timeout';

export interface MeasurementStatistics {
  min: number;
  max: number;
  avg: number;
  stdDev: number;
  count: number;
}

export interface MeasurementStream {
  probeId: string;
  target: string;
  measurements: Measurement[];
  timestamp: string;
}

export interface Configuration {
  id: string;
  name: string;
  description?: string;
  version: string;
  data: Record<string, any>;
  createdAt: string;
  updatedAt: string;
  createdBy?: string;
  tags?: string[];
}

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  createdAt: string;
  lastLogin?: string;
}

export type UserRole = 'admin' | 'operator' | 'viewer';

export interface AuthToken {
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: AuthToken;
  user: User;
}

export interface WebSocketMessage {
  type: WebSocketMessageType;
  data: any;
  timestamp: string;
}

export type WebSocketMessageType = 
  | 'probe_update'
  | 'measurement'
  | 'health_status'
  | 'config_change'
  | 'alert';

export interface DashboardStats {
  totalProbes: number;
  onlineProbes: number;
  offlineProbes: number;
  degradedProbes: number;
  totalMeasurements: number;
  averageLatency: number;
  lastUpdate: string;
}

export interface Alert {
  id: string;
  type: AlertType;
  severity: AlertSeverity;
  title: string;
  message: string;
  probeId?: string;
  timestamp: string;
  acknowledged: boolean;
  resolved: boolean;
}

export type AlertType = 'probe_offline' | 'high_latency' | 'packet_loss' | 'configuration_error';

export type AlertSeverity = 'low' | 'medium' | 'high' | 'critical';

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
  timestamp: string;
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

export interface ChartDataPoint {
  timestamp: string;
  value: number;
  label?: string;
}

export interface ChartSeries {
  name: string;
  data: ChartDataPoint[];
  color?: string;
}

// Frontend-only: Reusable build configuration model for assigning config to probes.
// These are stored under Configuration.data.buildConfigs in the backend configuration.
export interface BuildConfiguration {
  id: string;
  name: string;
  description?: string;
  // Optional link into central configuration; if set, ties this build config to a specific backend config
  configId?: string;
  configVersion?: number;
  configSource?: string;
  // Arbitrary configuration fragment that can be merged into /config.data if used
  spec?: Record<string, any>;
}

// Frontend-only: Ping/Traceroute measurement task definition bound to configs/probes.
// These are stored under Configuration.data.measurementTasks.
export type MeasurementTaskType = 'ping' | 'traceroute';

export interface MeasurementTaskTarget {
  id: string;
  address: string;
  description?: string;
}

export interface MeasurementTask {
  id: string;
  name: string;
  type: MeasurementTaskType;
  targets: MeasurementTaskTarget[];
  intervalSeconds: number;
  timeoutSeconds: number;
  // Bind to a build configuration that defines how probes execute this task
  buildConfigId?: string;
  // Optional explicit list of probe IDs for ad-hoc assignments
  probeIds?: string[];
  enabled: boolean;
}

// Optional configuration tracking fields for probes surfaced in UI; backend may provide via ProbeRegistry
export interface ProbeConfigMetadata {
  configId?: string;
  configVersion?: number;
  configSource?: string;
  configAppliedAt?: string;
}