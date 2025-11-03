import React, { useState, useEffect } from 'react';
import { Activity, Plus, Settings, AlertCircle, CheckCircle, Clock } from 'lucide-react';
import { apiService } from '@/services/api';
import { webSocketService } from '@/services/websocket';
import type { Probe } from '@/types';

const Probes: React.FC = () => {
  const [probes, setProbes] = useState<Probe[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProbes();
    setupWebSocket();
  }, []);

  const loadProbes = async () => {
    try {
      setLoading(true);
      const data = await apiService.getProbes();
      setProbes(data);
    } catch (error) {
      console.error('Failed to load probes:', error);
    } finally {
      setLoading(false);
    }
  };

  const setupWebSocket = () => {
    webSocketService.on('probe_update', (data: Probe[]) => {
      setProbes(data);
    });
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
      case 'offline': return <AlertCircle className="h-5 w-5 text-red-600" />;
      case 'degraded': return <Clock className="h-5 w-5 text-yellow-600" />;
      default: return <Activity className="h-5 w-5 text-gray-600" />;
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Probes</h1>
          <p className="text-gray-600">Manage and monitor your mesh probe network</p>
        </div>
        <button className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700">
          <Plus className="h-4 w-4 mr-2" />
          Add Probe
        </button>
      </div>

      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <ul className="divide-y divide-gray-200">
          {probes.map((probe) => (
            <li key={probe.id}>
              <div className="px-4 py-4 sm:px-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    {getStatusIcon(probe.status)}
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-900">{probe.name}</p>
                      <p className="text-sm text-gray-500">{probe.platform} • {probe.arch} • {probe.ipAddress}</p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(probe.status)}`}>
                      {probe.status}
                    </span>
                    <button className="text-gray-400 hover:text-gray-500">
                      <Settings className="h-4 w-4" />
                    </button>
                  </div>
                </div>
                <div className="mt-2 sm:flex sm:justify-between">
                  <div className="sm:flex">
                    <p className="flex items-center text-sm text-gray-500">
                      Version: {probe.version}
                    </p>
                  </div>
                  <div className="mt-2 flex items-center text-sm text-gray-500 sm:mt-0">
                    <p>
                      Last seen: {new Date(probe.lastSeen).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default Probes;