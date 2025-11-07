import React, { useEffect, useState } from 'react';
import { apiService } from '@/services/api';
import type { Configuration } from '@/types';

interface StatusState {
  loading: boolean;
  error: string | null;
  data: any | null;
}

interface ConfigState {
  loading: boolean;
  error: string | null;
  saving: boolean;
  config: Configuration | null;
  editorValue: string;
}

interface ToastState {
  message: string;
  type: 'success' | 'error' | 'info';
}

const parseConfig = (raw: any): Configuration | null => {
  if (!raw) return null;

  // Ensure we have a consistent Configuration-like view for UI
  return {
    id: raw.id ?? raw.ID ?? 'current',
    name: raw.name ?? 'Active configuration',
    description: raw.description,
    version: String(raw.version ?? raw.Version ?? '1'),
    data: raw.data ?? raw.Data ?? raw,
    createdAt: raw.createdAt ?? raw.CreatedAt ?? new Date().toISOString(),
    updatedAt: raw.updatedAt ?? raw.UpdatedAt ?? new Date().toISOString(),
    createdBy: raw.createdBy,
    tags: raw.tags,
  };
};

const NetworkConfiguration: React.FC = () => {
  const [status, setStatus] = useState<StatusState>({
    loading: true,
    error: null,
    data: null,
  });

  const [configState, setConfigState] = useState<ConfigState>({
    loading: true,
    error: null,
    saving: false,
    config: null,
    editorValue: '',
  });

  const [toast, setToast] = useState<ToastState | null>(null);

  const showToast = (message: string, type: ToastState['type']) => {
    setToast({ message, type });
    setTimeout(() => setToast(null), 4000);
  };

  const loadConfig = async () => {
    setConfigState(prev => ({ ...prev, loading: true, error: null }));
    try {
      const raw = await apiService.getConfig();
      const cfg = parseConfig(raw);
      const editorValue = JSON.stringify(cfg?.data ?? raw ?? {}, null, 2);
      setConfigState(prev => ({
        ...prev,
        loading: false,
        config: cfg,
        editorValue,
        error: null,
      }));
    } catch (err: any) {
      const msg =
        err?.response?.data?.error ||
        err?.message ||
        'Failed to load active configuration';
      setConfigState(prev => ({
        ...prev,
        loading: false,
        config: null,
        editorValue: '',
        error: msg,
      }));
    }
  };

  const loadStatus = async () => {
    setStatus(prev => ({ ...prev, loading: true, error: null }));
    try {
      const s = await apiService.getConfigStatus();
      setStatus({
        loading: false,
        error: null,
        data: s,
      });
    } catch (err: any) {
      const msg =
        err?.response?.data?.error ||
        err?.message ||
        'Failed to load configuration status';
      setStatus({
        loading: false,
        error: msg,
        data: null,
      });
    }
  };

  useEffect(() => {
    let isMounted = true;
    const bootstrap = async () => {
      await Promise.all([loadConfig(), loadStatus()]);
    };
    if (isMounted) bootstrap();
    const interval = setInterval(() => {
      loadStatus();
    }, 15000);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleEditorChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setConfigState(prev => ({ ...prev, editorValue: value }));
  };

  const handleSave = async () => {
    if (!configState.editorValue.trim()) {
      showToast('Configuration content cannot be empty', 'error');
      return;
    }

    let parsed: any;
    try {
      parsed = JSON.parse(configState.editorValue);
    } catch (err: any) {
      showToast('Invalid JSON: ' + (err.message || 'parse error'), 'error');
      return;
    }

    setConfigState(prev => ({ ...prev, saving: true, error: null }));
    try {
      const currentId = configState.config?.id || 'config-current';
      const payload = {
        ...configState.config,
        data: parsed,
      };

      const updated = await apiService.updateConfiguration(currentId, payload);
      const cfg = parseConfig(updated);
      setConfigState(prev => ({
        ...prev,
        saving: false,
        config: cfg,
        editorValue: JSON.stringify(cfg?.data ?? parsed, null, 2),
      }));
      showToast('Configuration updated successfully', 'success');
      await loadStatus();
    } catch (err: any) {
      const msg =
        err?.response?.data?.error ||
        err?.message ||
        'Failed to save configuration';
      setConfigState(prev => ({
        ...prev,
        saving: false,
        error: msg,
      }));
      showToast(msg, 'error');
    }
  };

  const handlePropagate = async () => {
    try {
      const res = await apiService.propagateConfiguration();
      showToast(
        res?.message || 'Configuration propagation triggered',
        'info'
      );
      await loadStatus();
    } catch (err: any) {
      const msg =
        err?.response?.data?.error ||
        err?.message ||
        'Failed to propagate configuration';
      showToast(msg, 'error');
    }
  };

  return (
    <div className="space-y-4">
      {toast && (
        <div
          className={`px-3 py-2 rounded text-xs ${
            toast.type === 'success'
              ? 'bg-emerald-900/40 text-emerald-300'
              : toast.type === 'error'
              ? 'bg-rose-900/40 text-rose-300'
              : 'bg-slate-800/80 text-slate-200'
          }`}
        >
          {toast.message}
        </div>
      )}

      <div className="grid md:grid-cols-3 gap-4">
        <div className="md:col-span-2 bg-slate-900 border border-slate-800 rounded-xl p-4">
          <div className="flex items-center justify-between mb-3">
            <div>
              <h1 className="text-sm font-semibold text-slate-100">
                Central Configuration
              </h1>
              <p className="text-[10px] text-slate-500">
                Backed by config.Manager via /config endpoints.
              </p>
            </div>
            <div className="flex gap-2">
              <button
                onClick={loadConfig}
                className="px-2 py-1 text-[10px] rounded bg-slate-800 text-slate-200 hover:bg-slate-700"
              >
                Reload
              </button>
              <button
                onClick={handlePropagate}
                className="px-2 py-1 text-[10px] rounded bg-emerald-600 text-slate-950 hover:bg-emerald-500"
              >
                Propagate
              </button>
            </div>
          </div>

          {configState.loading && (
            <div className="text-[10px] text-slate-500">
              Loading active configuration...
            </div>
          )}

          {configState.error && !configState.loading && (
            <div className="text-[10px] text-rose-400 mb-2">
              {configState.error}
            </div>
          )}

          <textarea
            className="w-full h-64 text-[10px] font-mono bg-slate-950 border border-slate-800 rounded-lg p-2 text-slate-200 focus:outline-none focus:ring-1 focus:ring-emerald-500"
            value={configState.editorValue}
            onChange={handleEditorChange}
            placeholder="{ /* JSON configuration managed by central providers */ }"
          />

          <div className="mt-2 flex justify-end">
            <button
              onClick={handleSave}
              disabled={configState.saving}
              className={`px-3 py-1 text-[10px] rounded ${
                configState.saving
                  ? 'bg-slate-700 text-slate-400'
                  : 'bg-emerald-500 text-slate-950 hover:bg-emerald-400'
              }`}
            >
              {configState.saving ? 'Saving...' : 'Save Configuration'}
            </button>
          </div>
        </div>

        <div className="bg-slate-900 border border-slate-800 rounded-xl p-4">
          <h2 className="text-xs font-semibold text-slate-100 mb-2">
            Provider & Propagation Status
          </h2>

          {status.loading && (
            <div className="text-[10px] text-slate-500">
              Loading status from /config/status...
            </div>
          )}

          {status.error && !status.loading && (
            <div className="text-[10px] text-rose-400">
              {status.error}
            </div>
          )}

          {status.data && !status.loading && (
            <div className="space-y-1 text-[10px] text-slate-300">
              <div>
                <span className="text-slate-500">Current Config ID: </span>
                <span>{status.data.currentConfigID || 'n/a'}</span>
              </div>
              <div>
                <span className="text-slate-500">Last Update: </span>
                <span>
                  {status.data.lastUpdate
                    ? new Date(status.data.lastUpdate).toLocaleString()
                    : 'n/a'}
                </span>
              </div>
              <div>
                <span className="text-slate-500">Update Count: </span>
                <span>{status.data.updateCount ?? 'n/a'}</span>
              </div>
              <div>
                <span className="text-slate-500">Health Score: </span>
                <span>
                  {status.data.healthScore != null
                    ? status.data.healthScore.toFixed
                      ? status.data.healthScore.toFixed(2)
                      : status.data.healthScore
                    : 'n/a'}
                </span>
              </div>
              <div className="mt-2">
                <div className="text-[9px] text-slate-500 mb-1">
                  Providers
                </div>
                <div className="space-y-1">
                  {(status.data.sources || []).map((src: any, idx: number) => (
                    <div
                      key={`${src.name || 'src'}-${idx}`}
                      className="flex items-center justify-between text-[9px]"
                    >
                      <span className="text-slate-400">
                        {src.name || 'provider'}
                      </span>
                      <span
                        className={`px-1.5 py-0.5 rounded-full ${
                          src.healthy
                            ? 'bg-emerald-900/40 text-emerald-300'
                            : 'bg-rose-900/40 text-rose-300'
                        }`}
                      >
                        {src.healthy ? 'healthy' : 'unhealthy'}
                      </span>
                    </div>
                  ))}
                  {(!status.data.sources ||
                    status.data.sources.length === 0) && (
                    <div className="text-[9px] text-slate-500">
                      No provider details reported.
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default NetworkConfiguration;
