import React, { useEffect, useState } from 'react';
import { apiService } from '@/services/api';

interface PacketLossChartState {
  lossRate: number | null;
  loading: boolean;
  error: string | null;
}

const PacketLossChart: React.FC = () => {
  const [state, setState] = useState<PacketLossChartState>({
    lossRate: null,
    loading: true,
    error: null,
  });

  useEffect(() => {
    let isMounted = true;

    const loadStats = async () => {
      try {
        const stats = await apiService.getMeasurementStatistics();
        if (!isMounted) return;

        // Backend: stats.packet_loss_rate is a number (percentage) in current demo impl
        const raw = stats.packet_loss_rate ?? stats.packetLossRate ?? 0;
        const lossRate =
          typeof raw === 'string'
            ? parseFloat(raw.replace('%', '').trim()) || 0
            : Number(raw) || 0;

        setState({
          lossRate,
          loading: false,
          error: null,
        });
      } catch (err: any) {
        if (!isMounted) return;
        const message =
          err?.response?.data?.error ||
          err?.message ||
          'Failed to load packet loss statistics';
        setState({
          lossRate: null,
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
          Packet Loss (loading from /measurements/statistics)
        </div>
        <div className="h-16 bg-slate-900 animate-pulse rounded" />
      </div>
    );
  }

  if (state.error || state.lossRate === null) {
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
        <div className="text-xs text-slate-500 mb-1">
          Packet Loss
        </div>
        <div className="text-[10px] text-slate-500">
          {state.error ||
            'No packet loss statistics available yet. Once probes run, this will display loss percentage.'}
        </div>
      </div>
    );
  }

  const loss = Math.max(0, Math.min(state.lossRate, 100));

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
      <div className="flex items-center justify-between mb-2">
        <div className="text-xs text-slate-500">
          Packet Loss Rate
        </div>
        <div className="text-[9px] text-slate-500">
          /measurements/statistics
        </div>
      </div>
      <div className="h-4 w-full bg-slate-900 rounded-full border border-slate-800 overflow-hidden">
        <div
          className={`h-full rounded-full ${
            loss < 1
              ? 'bg-emerald-500/80'
              : loss < 5
              ? 'bg-amber-400/80'
              : 'bg-rose-500/80'
          }`}
          style={{ width: `${loss}%` }}
        />
      </div>
      <div className="mt-1 text-[10px] text-slate-300">
        {loss.toFixed(2)}% loss
      </div>
      <div className="text-[8px] text-slate-500">
        Lower is better. Values are demo/aggregated until real measurement feeds are enabled.
      </div>
    </div>
  );
};

export default PacketLossChart;
