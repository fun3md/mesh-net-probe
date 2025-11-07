import React, { useEffect, useState } from 'react';
import type { DashboardStats } from '@/types';
import { apiService } from '@/services/api';

interface StatCardsState {
  stats: DashboardStats | null;
  loading: boolean;
  error: string | null;
}

const StatCards: React.FC = () => {
  const [state, setState] = useState<StatCardsState>({
    stats: null,
    loading: true,
    error: null,
  });

  useEffect(() => {
    let isMounted = true;

    const loadStats = async () => {
      try {
        const stats = await apiService.getDashboardStats();
        if (!isMounted) return;
        setState({
          stats,
          loading: false,
          error: null,
        });
      } catch (err: any) {
        if (!isMounted) return;
        const message =
          err?.response?.data?.error ||
          err?.message ||
          'Failed to load dashboard statistics';
        setState({
          stats: null,
          loading: false,
          error: message,
        });
      }
    };

    loadStats();
    const interval = setInterval(loadStats, 10000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  if (state.loading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {[1, 2, 3, 4].map(i => (
          <div
            key={i}
            className="bg-slate-900 border border-slate-800 rounded-xl p-3 animate-pulse"
          >
            <div className="h-3 w-16 bg-slate-800 rounded mb-2" />
            <div className="h-5 w-12 bg-slate-800 rounded" />
          </div>
        ))}
      </div>
    );
  }

  if (state.error || !state.stats) {
    return (
      <div className="bg-slate-900 border border-rose-900/60 rounded-xl p-3">
        <div className="text-xs text-rose-400 font-medium">
          Unable to load dashboard statistics
        </div>
        <div className="text-[10px] text-slate-500 mt-1">
          {state.error || 'No data returned from /monitoring/dashboard'}
        </div>
      </div>
    );
  }

  const { stats } = state;

  // Derive some basic counts if needed (offline/degraded placeholders)
  const online = stats.onlineProbes;
  const total = stats.totalProbes;
  const offline = Math.max(total - online, 0);

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-3">
        <div className="text-[10px] text-slate-500">Total Probes</div>
        <div className="text-lg font-semibold text-slate-100">
          {total}
        </div>
        <div className="text-[9px] text-slate-500">
          {online} online / {offline} offline
        </div>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl p-3">
        <div className="text-[10px] text-slate-500">Total Measurements</div>
        <div className="text-lg font-semibold text-slate-100">
          {stats.totalMeasurements}
        </div>
        <div className="text-[9px] text-slate-500">
          Aggregated from monitoring backend
        </div>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl p-3">
        <div className="text-[10px] text-slate-500">Avg Latency</div>
        <div className="text-lg font-semibold text-emerald-400">
          {stats.averageLatency?.toFixed
            ? `${stats.averageLatency.toFixed(1)} ms`
            : `${stats.averageLatency} ms`}
        </div>
        <div className="text-[9px] text-slate-500">
          From /monitoring/dashboard
        </div>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl p-3">
        <div className="text-[10px] text-slate-500">Last Update</div>
        <div className="text-[10px] font-mono text-slate-300">
          {new Date(stats.lastUpdate).toLocaleTimeString()}
        </div>
        <div className="text-[9px] text-slate-500">
          Auto-refreshing every 10s
        </div>
      </div>
    </div>
  );
};

export default StatCards;
