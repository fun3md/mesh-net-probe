import React, { useState, useEffect } from 'react';
import {
  Activity,
  AlertTriangle,
  CheckCircle,
  Clock,
  TrendingUp,
  Server
} from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts';
import type { DashboardStats, Probe, Alert } from '@/types';
import { apiService } from '@/services/api';
import { webSocketService } from '@/services/websocket';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [probes, setProbes] = useState<Probe[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [chartData, setChartData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboardData();
    setupWebSocketConnection();

    return () => {
      webSocketService.disconnect();
    };
  }, []);

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      const [dashboardStats, probeData, alertData] = await Promise.all([
        apiService.getDashboardStats(),
        apiService.getProbes(),
        apiService.getAlerts()
      ]);

      setStats(dashboardStats);
      setProbes(probeData);
      setAlerts(alertData);
      generateChartData(probeData);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const setupWebSocketConnection = async () => {
    try {
      await webSocketService.connect();
      
      // Subscribe to real-time updates
      webSocketService.subscribeToProbeUpdates();
      webSocketService.subscribeToHealthUpdates();
      
      // Set up event handlers
      webSocketService.on('probe_update', (data: Probe[]) => {
        setProbes(data);
        updateStats(data);
      });

      webSocketService.on('health_status', (data: any) => {
        // Update health indicators
        console.log('Health status update:', data);
      });

      webSocketService.on('alert', (alert: Alert) => {
        setAlerts(prev => [alert, ...prev.slice(0, 9)]); // Keep last 10 alerts
      });

    } catch (error) {
      console.error('Failed to connect to WebSocket:', error);
    }
  };

  const updateStats = (probeData: Probe[]) => {
    const online = probeData.filter(p => p.status === 'online').length;
    const offline = probeData.filter(p => p.status === 'offline').length;
    const degraded = probeData.filter(p => p.status === 'degraded').length;

    setStats(prev => prev ? {
      ...prev,
      totalProbes: probeData.length,
      onlineProbes: online,
      offlineProbes: offline,
      degradedProbes: degraded
    } : null);
  };

  const generateChartData = (probeData: Probe[]) => {
    // Generate sample chart data (in real app, this would come from API)
    const now = new Date();
    const data = Array.from({ length: 24 }, (_, i) => ({
      time: new Date(now.getTime() - (23 - i) * 60 * 60 * 1000).toLocaleTimeString(),
      latency: Math.random() * 100 + 10,
      success: 95 + Math.random() * 5,
      probes: probeData.length + Math.floor(Math.random() * 3) - 1
    }));
    setChartData(data);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'text-green-600 bg-green-100';
      case 'offline': return 'text-red-600 bg-red-100';
      case 'degraded': return 'text-yellow-600 bg-yellow-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online': return <CheckCircle className="h-5 w-5 text-green-600" />;
      case 'offline': return <AlertTriangle className="h-5 w-5 text-red-600" />;
      case 'degraded': return <Clock className="h-5 w-5 text-yellow-600" />;
      default: return <Activity className="h-5 w-5 text-gray-600" />;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600">Monitor your mesh probe network in real-time</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Server className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">Total Probes</dt>
                  <dd className="text-lg font-medium text-gray-900">{stats?.totalProbes || 0}</dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <CheckCircle className="h-6 w-6 text-green-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">Online Probes</dt>
                  <dd className="text-lg font-medium text-gray-900">{stats?.onlineProbes || 0}</dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <TrendingUp className="h-6 w-6 text-blue-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">Avg Latency</dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {stats?.averageLatency.toFixed(1) || '0.0'}ms
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <AlertTriangle className="h-6 w-6 text-red-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">Active Alerts</dt>
                  <dd className="text-lg font-medium text-gray-900">{alerts.length}</dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Latency Chart */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Network Latency</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Line type="monotone" dataKey="latency" stroke="#3b82f6" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Probe Status Chart */}
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Probe Availability</h3>
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Area type="monotone" dataKey="success" stackId="1" stroke="#10b981" fill="#10b981" fillOpacity={0.6} />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Probe Status and Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Probes */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Recent Probes</h3>
            <div className="space-y-3">
              {probes.slice(0, 5).map((probe) => (
                <div key={probe.id} className="flex items-center justify-between">
                  <div className="flex items-center">
                    {getStatusIcon(probe.status)}
                    <div className="ml-3">
                      <p className="text-sm font-medium text-gray-900">{probe.name}</p>
                      <p className="text-sm text-gray-500">{probe.platform} {probe.arch}</p>
                    </div>
                  </div>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(probe.status)}`}>
                    {probe.status}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Recent Alerts */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Recent Alerts</h3>
            <div className="space-y-3">
              {alerts.slice(0, 5).map((alert) => (
                <div key={alert.id} className="flex items-start">
                  <AlertTriangle className={`h-5 w-5 mt-0.5 ${
                    alert.severity === 'critical' ? 'text-red-500' :
                    alert.severity === 'high' ? 'text-orange-500' :
                    alert.severity === 'medium' ? 'text-yellow-500' :
                    'text-blue-500'
                  }`} />
                  <div className="ml-3">
                    <p className="text-sm font-medium text-gray-900">{alert.title}</p>
                    <p className="text-sm text-gray-500">{alert.message}</p>
                    <p className="text-xs text-gray-400">{new Date(alert.timestamp).toLocaleString()}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;