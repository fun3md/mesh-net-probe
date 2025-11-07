import React from 'react';
import { useEffect, useState } from 'react';
import { apiService } from '@/services/api';

interface LatencyPoint {
  label: string;
  value: number;
}

interface LatencyChartState {
  points: LatencyPoint[];
  loading: boolean;
  error: string | null;
}

const LatencyChart: React.FC = () => {
  const [state, setState] = useState<LatencyChartState>({
    points: [],
    loading: true,
    error: null,
  });

  useEffect(() => {
    let isMounted = true;

    const loadStats = async () => {
      try {
        const stats = await apiService.getMeasurementStatistics();
        if (!isMounted) return;

        const points: LatencyPoint[] = [
          {
            label: 'Avg',
            value: parseFloat(
              typeof stats.average_rtt === 'string'
                ? stats.average_rtt
                    .toLowerCase()
                    .replace('ms', '')
                    .trim()
                : stats.average_rtt ?? 0
            ) || 0,
          },
          {
            label: 'Min',
            value: parseFloat(
              typeof stats.min_rtt === 'string'
                ? stats.min_rtt
                    .toLowerCase()
                    .replace('ms', '')
                    .trim()
                : stats.min_rtt ?? 0
            ) || 0,
          },
          {
            label: 'Max',
            value: parseFloat(
              typeof stats.max_rtt === 'string'
                ? stats.max_rtt
                    .toLowerCase()
                    .replace('ms', '')
                    .trim()
                : stats.max_rtt ?? 0
            ) || 0,
          },
        ];

        setState({
          points,
          loading: false,
          error: null,
        });
      } catch (err: any) {
        if (!isMounted) return;
        const message =
          err?.response?.data?.error ||
          err?.message ||
          'Failed to load latency statistics';
        setState({
          points: [],
          loading: false,
          error: message,
        });
      }
    };

    loadStats();
    const interval = setInterval(loadStats, 15000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  if (state.loading) {
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
        <div className="text-xs text-slate-500 mb-2">
          Latency (loading from /measurements/statistics)
        </div>
        <div className="h-16 bg-slate-900 animate-pulse rounded" />
      </div>
    );
  }

  if (state.error || !state.points.length) {
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
        <div className="text-xs text-slate-500 mb-1">
          Latency
        </div>
        <div className="text-[10px] text-slate-500">
          {state.error ||
            'No latency statistics available yet. Run probes to populate measurements.'}
        </div>
      </div>
    );
  }

  const maxValue = Math.max(...state.points.map(p => p.value), 1);

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
      <div className="text-xs text-slate-500 mb-2">
        Latency (ms) from /measurements/statistics
      </div>
      <div className="flex items-end gap-3 h-24">
        {state.points.map(point => (
          <div key={point.label} className="flex-1 flex flex-col items-center">
            <div
              className="w-4 bg-emerald-500/80 rounded-t"
              style={{
                height: `${(point.value / maxValue) * 100}%`,
              }}
            />
            <div className="text-[9px] text-slate-400 mt-1">
              {point.label}
            </div>
            <div className="text-[8px] text-slate-500">
              {point.value.toFixed(1)} ms
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default LatencyChart;
