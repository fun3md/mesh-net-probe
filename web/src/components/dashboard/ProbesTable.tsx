import React, { useEffect, useState } from 'react';
import type { Probe } from '@/types';
import { apiService } from '@/services/api';

interface ProbesTableState {
  probes: Probe[];
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

const statusColor = (status: string) => {
  switch (status) {
    case 'online':
      return 'text-emerald-400';
    case 'degraded':
      return 'text-amber-400';
    case 'offline':
      return 'text-rose-400';
    default:
      return 'text-slate-400';
  }
};

const ProbesTable: React.FC = () => {
  const [state, setState] = useState<ProbesTableState>({
    probes: [],
    loading: true,
    error: null,
    lastUpdated: null,
  });

  const loadProbes = async () => {
    try {
      const probes = await apiService.getProbes();
      setState({
        probes,
        loading: false,
        error: null,
        lastUpdated: new Date().toISOString(),
      });
    } catch (err: any) {
      const message =
        err?.response?.data?.error ||
        err?.message ||
        'Failed to load probes from backend';
      setState(prev => ({
        ...prev,
        probes: [],
        loading: false,
        error: message,
      }));
    }
  };

  useEffect(() => {
    let isMounted = true;

    const init = async () => {
      if (!isMounted) return;
      setState(prev => ({ ...prev, loading: true, error: null }));
      await loadProbes();
    };

    init();
    const interval = setInterval(() => {
      if (isMounted) {
        loadProbes();
      }
    }, 10000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (state.loading) {
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
        <div className="text-sm text-slate-400">Loading probes...</div>
      </div>
    );
  }

  if (state.error) {
    return (
      <div className="bg-slate-900 border border-rose-900/60 rounded-xl p-4">
        <div className="text-sm text-rose-400 font-medium">
          Unable to load probes
        </div>
        <div className="text-xs text-slate-500 mt-1">
          {state.error}
        </div>
      </div>
    );
  }

  if (!state.probes.length) {
    return (
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
        <div className="flex items-center justify-between mb-1">
          <h2 className="text-sm font-semibold text-slate-200">
            Probes
          </h2>
          <button
            onClick={loadProbes}
            className="px-2 py-1 text-[10px] rounded bg-slate-800 text-slate-200 hover:bg-slate-700"
          >
            Refresh
          </button>
        </div>
        <div className="text-sm text-slate-400">
          No probes registered yet. Deploy probes so they register via /probes and report config-applied.
        </div>
      </div>
    );
  }

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 overflow-x-auto">
      <div className="flex items-center justify-between mb-3">
        <div>
          <h2 className="text-sm font-semibold text-slate-200">
            Probes ({state.probes.length})
          </h2>
          <span className="block text-[9px] text-slate-500">
            Data: /probes via ProbeRegistry, incl. config-applied metadata where available.
          </span>
        </div>
        <div className="flex items-center gap-2">
          {state.lastUpdated && (
            <span className="text-[8px] text-slate-500">
              Updated {new Date(state.lastUpdated).toLocaleTimeString()}
            </span>
          )}
          <button
            onClick={loadProbes}
            className="px-2 py-1 text-[10px] rounded bg-slate-800 text-slate-200 hover:bg-slate-700"
          >
            Refresh
          </button>
        </div>
      </div>
      <table className="min-w-full text-left text-[10px] text-slate-300">
        <thead className="border-b border-slate-800 text-slate-500">
          <tr>
            <th className="py-2 pr-4">ID</th>
            <th className="py-2 pr-4">Name</th>
            <th className="py-2 pr-4">Platform</th>
            <th className="py-2 pr-4">IP</th>
            <th className="py-2 pr-4">Status</th>
            <th className="py-2 pr-4">Last Seen</th>
            <th className="py-2 pr-4">Config Version</th>
          </tr>
        </thead>
        <tbody>
          {state.probes.map((probe) => {
            // ProbeRegistry (via admin-web contract) may expose:
            // config_id, config_version, config_source, config_applied_at
            const anyProbe: any = probe as any;
            const configId = anyProbe.config_id || anyProbe.configId;
            const configVersion = anyProbe.config_version ?? anyProbe.configVersion;
            const configSource = anyProbe.config_source || anyProbe.configSource;
            const configAppliedAt = anyProbe.config_applied_at || anyProbe.configAppliedAt;

            let configSummary = '—';
            if (configId || configVersion || configSource) {
              const parts: string[] = [];
              if (configId) parts.push(String(configId));
              if (configVersion != null) parts.push(`v${configVersion}`);
              if (configSource) parts.push(`from ${configSource}`);
              configSummary = parts.join(' ');
            }

            const appliedDisplay = configAppliedAt
              ? new Date(configAppliedAt).toLocaleTimeString()
              : null;

            return (
              <tr
                key={probe.id}
                className="border-b border-slate-900 hover:bg-slate-900/60"
              >
                <td className="py-2 pr-4 font-mono text-[9px] text-slate-500">
                  {probe.id}
                </td>
                <td className="py-2 pr-4">
                  <div className="text-[10px] text-slate-200">
                    {probe.name || '—'}
                  </div>
                  <div className="text-[8px] text-slate-500">
                    {configSummary}
                  </div>
                </td>
                <td className="py-2 pr-4 text-[9px] text-slate-400">
                  {probe.platform}/{probe.arch}
                </td>
                <td className="py-2 pr-4 text-[9px] text-slate-400">
                  {probe.ipAddress || '—'}
                </td>
                <td className="py-2 pr-4">
                  <span
                    className={`text-[9px] font-medium ${statusColor(
                      probe.status
                    )}`}
                  >
                    {probe.status || 'unknown'}
                  </span>
                </td>
                <td className="py-2 pr-4 text-[9px] text-slate-500">
                  {probe.lastSeen
                    ? new Date(probe.lastSeen).toLocaleString()
                    : '—'}
                </td>
                <td className="py-2 pr-4 text-[9px] text-slate-500">
                  {appliedDisplay ? `${configSummary} @ ${appliedDisplay}` : configSummary}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};

export default ProbesTable;
