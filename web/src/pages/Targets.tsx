import React, { useEffect, useMemo, useState } from 'react';
import { Search, ListFilter, Plus, Pencil, Trash2, FlaskConical, Tag, X } from 'lucide-react';
import apiService from '@/services/api';
import type { Probe, BuildConfiguration, MeasurementTask, MeasurementTaskType } from '@/types';

type TargetRow = {
  id: string;
  address: string;
  name?: string;
  platform?: string;
  arch?: string;
  status?: string;
  lastSeen?: string;
  tags?: string[];
  // Config assignment metadata (from Probe)
  configId?: string;
  configVersion?: number;
  configSource?: string;
  configAppliedAt?: string;
};

type TargetFormState = {
  id: string;
  address: string;
  alias: string;
  platform: string;
  arch: string;
};

// Simple in-UI build config + task state shapes (aligned with types)
type LocalBuildConfig = BuildConfiguration;
type LocalMeasurementTask = MeasurementTask;

const mapProbeToRow = (probe: Probe): TargetRow => ({
  id: probe.id,
  address: (probe as any).ip_address || (probe as any).ipAddress || probe.id,
  name: probe.name,
  platform: (probe as any).platform,
  arch: (probe as any).arch,
  status: (probe as any).status || 'unknown',
  lastSeen: (probe as any).lastSeen,
  tags: (probe as any).tags || [],
  configId: (probe as any).configId ?? (probe as any).config_id,
  configVersion: (probe as any).configVersion ?? (probe as any).config_version,
  configSource: (probe as any).configSource ?? (probe as any).config_source,
  configAppliedAt: (probe as any).configAppliedAt ?? (probe as any).config_applied_at,
});

const buildCreatePayload = (form: TargetFormState) => {
  const payload: any = {
    id: form.id || form.address,
    name: form.alias || form.id || form.address,
    platform: form.platform || 'linux',
    arch: form.arch || 'amd64',
    ip_address: form.address || undefined,
  };
  if (!payload.id) {
    throw new Error('ID or Target Address is required');
  }
  return payload;
};

const buildUpdatePayload = (form: TargetFormState) => {
  const payload: any = {
    name: form.alias || undefined,
    platform: form.platform || undefined,
    arch: form.arch || undefined,
    ip_address: form.address || undefined,
  };
  return payload;
};

const StatusBadge: React.FC<{ status: 'Valid' | 'Warning' | 'Invalid' | string }> = ({ status }) => {
  const normalized =
    status === 'Valid' || status === 'Warning' || status === 'Invalid'
      ? status
      : status === 'active' || status === 'online'
        ? 'Valid'
        : status === 'degraded' || status === 'warning'
          ? 'Warning'
          : status === 'invalid' || status === 'offline'
            ? 'Invalid'
            : 'Warning';

  const statusClasses: Record<string, string> = {
    Valid: 'bg-success/10 text-success dark:bg-success/20 dark:text-green-300',
    Warning: 'bg-warning/10 text-warning dark:bg-warning/20 dark:text-yellow-300',
    Invalid: 'bg-error/10 text-error dark:bg-error/20 dark:text-red-300',
  };

  const dotClasses: Record<string, string> = {
    Valid: 'bg-success',
    Warning: 'bg-warning',
    Invalid: 'bg-error',
  };

  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full px-2 py-1 text-xs font-medium ${statusClasses[normalized]}`}>

      <span className={`h-1.5 w-1.5 rounded-full ${dotClasses[normalized]}`}></span>
      {typeof status === 'string' ? status : normalized}
    </span>
  );
};

const Targets: React.FC = () => {
  const [targets, setTargets] = useState<TargetRow[]>([]);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Side panel / form
  const [isPanelOpen, setIsPanelOpen] = useState<boolean>(false);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [editTarget, setEditTarget] = useState<TargetRow | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [form, setForm] = useState<TargetFormState>({
    id: '',
    address: '',
    alias: '',
    platform: '',
    arch: '',
  });

  // Build configurations and measurement tasks
  const [buildConfigs, setBuildConfigs] = useState<LocalBuildConfig[]>([]);
  const [tasks, setTasks] = useState<LocalMeasurementTask[]>([]);
  const [configError, setConfigError] = useState<string | null>(null);
  const [configSaving, setConfigSaving] = useState<boolean>(false);
  const [activeConfigId, setActiveConfigId] = useState<string>('');

  const hasSelection = selectedIds.size > 0;

  const filteredTargets = useMemo(() => targets, [targets]);

  const loadTargetsAndConfigs = async () => {
    setLoading(true);
    setError(null);
    setConfigError(null);
    try {
      const [probes, cfgs, mts] = await Promise.all([
        apiService.getProbesAdmin(),
        apiService.getBuildConfigurations(),
        apiService.getMeasurementTasks(),
      ]);

      const rows = (probes || []).map(mapProbeToRow);
      setTargets(rows);
      setSelectedIds(new Set());

      setBuildConfigs(cfgs || []);
      setTasks(mts || []);

      try {
        const cfg = await apiService.getConfig();
        setActiveConfigId(cfg && cfg.id ? cfg.id : '');
      } catch {
        setActiveConfigId('');
      }
    } catch (err: any) {
      if (err?.response?.status !== 401) {
        const msg = err?.message || 'Failed to load probes/configuration';
        setError(msg);
        setConfigError(msg);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTargetsAndConfigs();
  }, []);

  const openCreatePanel = () => {
    setEditTarget(null);
    setFormError(null);
    setForm({
      id: '',
      address: '',
      alias: '',
      platform: '',
      arch: '',
    });
    setIsPanelOpen(true);
  };

  const openEditPanel = (row: TargetRow) => {
    setEditTarget(row);
    setFormError(null);
    setForm({
      id: row.id,
      address: row.address || '',
      alias: row.name || '',
      platform: row.platform || '',
      arch: row.arch || '',
    });
    setIsPanelOpen(true);
  };

  const closePanel = () => {
    if (isSubmitting) return;
    setIsPanelOpen(false);
    setEditTarget(null);
    setFormError(null);
  };

  const handleFormChange = (field: keyof TargetFormState, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (isSubmitting) return;

    setIsSubmitting(true);
    setFormError(null);

    try {
      if (editTarget) {
        const payload = buildUpdatePayload(form);
        await apiService.updateProbeAdmin(editTarget.id, payload);
      } else {
        const payload = buildCreatePayload(form);
        await apiService.createProbeAdmin(payload);
      }
      await loadTargetsAndConfigs();
      setIsPanelOpen(false);
      setEditTarget(null);
    } catch (err: any) {
      if (err?.response?.status === 401) {
        setFormError('You are not authorized. Please log in again.');
      } else if (err?.response?.data?.error) {
        setFormError(err.response.data.error);
      } else {
        setFormError(err?.message || 'Failed to save target');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Delete this target/probe?')) return;
    try {
      await apiService.deleteProbeAdmin(id);
      await loadTargetsAndConfigs();
    } catch (err: any) {
      const msg = err?.response?.data?.error || err?.message || 'Failed to delete target';
      alert(msg);
    }
  };

  const handleDeleteSelected = async () => {
    if (!hasSelection) return;
    if (!window.confirm(`Delete ${selectedIds.size} selected target(s)?`)) return;

    try {
      for (const id of selectedIds) {
        try {
          await apiService.deleteProbeAdmin(id);
        } catch {
          // best-effort per-probe
        }
      }
      await loadTargetsAndConfigs();
    } catch (err: any) {
      const msg = err?.response?.data?.error || err?.message || 'Failed to delete selected targets';
      alert(msg);
    }
  };

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === filteredTargets.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(filteredTargets.map((t) => t.id)));
    }
  };

  const selectedCountLabel =
    hasSelection
      ? `${selectedIds.size} item${selectedIds.size > 1 ? 's' : ''} selected`
      : 'No items selected';

  // Build configuration handlers
  const handleAddBuildConfig = () => {
    const id = `build-${Date.now()}`;
    setBuildConfigs((prev) => [
      ...prev,
      { id, name: `Config ${prev.length + 1}`, spec: {}, description: '' },
    ]);
  };

  const handleUpdateBuildConfig = (id: string, patch: Partial<LocalBuildConfig>) => {
    setBuildConfigs((prev) => prev.map((c) => (c.id === id ? { ...c, ...patch } : c)));
  };

  const handleRemoveBuildConfig = (id: string) => {
    setBuildConfigs((prev) => prev.filter((c) => c.id !== id));
    setTasks((prev) =>
      prev.map((t) => (t.buildConfigId === id ? { ...t, buildConfigId: undefined } : t))
    );
  };

  // Persist buildConfigs + tasks into configuration.data
  const saveConfigsAndTasks = async () => {
    setConfigSaving(true);
    setConfigError(null);
    try {
      let cfg: any | null = null;
      try {
        cfg = await apiService.getConfig();
      } catch {
        cfg = null;
      }

      let targetConfigId = cfg?.id;
      let baseConfig: any = cfg;

      if (!baseConfig) {
        const now = new Date().toISOString();
        baseConfig = {
          id: 'frontend-config',
          name: 'Frontend Generated Configuration',
          description: 'Created via Admin Web UI',
          version: 1,
          createdAt: now,
          updatedAt: now,
          data: {},
        };
        const created = await apiService.createConfiguration(baseConfig);
        targetConfigId = created.id;
        baseConfig = created;
      }

      if (!targetConfigId) {
        targetConfigId = 'frontend-config';
      }

      const updated = {
        ...baseConfig,
        data: {
          ...(baseConfig.data || {}),
          buildConfigs,
          measurementTasks: tasks,
        },
      };

      await apiService.updateConfiguration(targetConfigId, updated);
      await apiService.propagateConfiguration();
      setActiveConfigId(targetConfigId);
    } catch (err: any) {
      const msg = err?.response?.data?.error || err?.message || 'Failed to save configuration';
      setConfigError(msg);
    } finally {
      setConfigSaving(false);
    }
  };

  // Assign build config to selected probes
  const handleAssignBuildConfigToSelected = async (buildConfigId: string) => {
    if (!hasSelection) return;
    try {
      const probeIds = Array.from(selectedIds);
      await apiService.assignBuildConfigToProbes(buildConfigId, probeIds, {
        configId: activeConfigId,
      });
      await loadTargetsAndConfigs();
    } catch (err: any) {
      const msg = err?.response?.data?.error || err?.message || 'Failed to assign build config';
      setConfigError(msg);
    }
  };

  // Measurement task handlers
  const handleAddTask = () => {
    const id = `task-${Date.now()}`;
    setTasks((prev) => [
      ...prev,
      {
        id,
        name: `Ping Task ${prev.length + 1}`,
        type: 'ping' as MeasurementTaskType,
        targets: [],
        intervalSeconds: 60,
        timeoutSeconds: 5,
        enabled: true,
      },
    ]);
  };

  const handleUpdateTask = (id: string, patch: Partial<LocalMeasurementTask>) => {
    setTasks((prev) => prev.map((t) => (t.id === id ? { ...t, ...patch } : t)));
  };

  const handleRemoveTask = (id: string) => {
    setTasks((prev) => prev.filter((t) => t.id !== id));
  };

  const handleUpdateTaskTarget = (taskId: string, address: string) => {
    setTasks((prev) =>
      prev.map((t) =>
        t.id === taskId
          ? {
              ...t,
              targets:
                t.targets && t.targets.length
                  ? [{ ...t.targets[0], address }]
                  : [{ id: `${taskId}-target-1`, address }],
            }
          : t
      )
    );
  };

  return (
    <div className="relative flex h-auto min-h-screen w-full flex-col">
      <main className="flex-1 p-4 sm:p-6 lg:p-8">
        {/* Header */}
        <div className="flex items-center justify-between pb-6">
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">
            Manage Probes / Targets
          </h1>
          <button
            onClick={openCreatePanel}
            className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 dark:ring-offset-background-dark"
            type="button"
          >
            <Plus className="h-4 w-4" />
            Add New Target
          </button>
        </div>

        {/* Error banner */}
        {error && (
          <div className="mb-4 rounded-md border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-800 dark:bg-red-950/70 dark:text-red-300">
            {error}
          </div>
        )}

        {/* Search + filters */}
        <div className="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-slate-500 dark:text-slate-400" />
            <input
              className="w-full rounded-lg border border-slate-300 bg-white py-2 pl-10 pr-4 text-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary"
              placeholder="Search targets (by ID, address, alias)..."
              type="text"
              onChange={(e) => {
                const q = e.target.value.toLowerCase();
                setTargets((prev) =>
                  prev.filter((t) => {
                    const haystack = `${t.id} ${t.address} ${t.name || ''}`.toLowerCase();
                    return haystack.includes(q);
                  })
                );
              }}
            />
          </div>
          <div className="flex items-center gap-2">
            <div className="relative">
              <button
                className="flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
                type="button"
              >
                <ListFilter className="h-4 w-4" />
                Filter by Status
              </button>
            </div>
          </div>
        </div>

        {/* Selection + build config / task controls */}
        <div className="mb-4 flex flex-col gap-3 rounded-lg bg-primary/10 p-3 dark:bg-primary/20">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium text-primary dark:text-blue-300">
              {selectedCountLabel}
            </span>
            <div className="flex flex-wrap items-center gap-3">
              <button
                className="flex items-center gap-1.5 text-sm font-medium text-slate-500 dark:text-slate-500 cursor-not-allowed"
                type="button"
                disabled
              >
                <FlaskConical className="h-4 w-4" />
                Run Pre-checks (coming soon)
              </button>
              <button
                className="flex items-center gap-1.5 text-sm font-medium text-slate-500 dark:text-slate-500 cursor-not-allowed"
                type="button"
                disabled
              >
                <Tag className="h-4 w-4" />
                Edit Tags (coming soon)
              </button>
              <button
                onClick={handleDeleteSelected}
                className={`flex items-center gap-1.5 text-sm font-medium ${
                  hasSelection
                    ? 'text-error hover:text-red-700 dark:text-red-400 dark:hover:text-red-300'
                    : 'text-slate-400 cursor-not-allowed'
                }`}
                type="button"
                disabled={!hasSelection}
              >
                <Trash2 className="h-4 w-4" />
                Delete Selected
              </button>
            </div>
          </div>

          {/* Build configuration and task controls */}
          <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
            <div className="flex flex-wrap items-center gap-2">
              <button
                type="button"
                onClick={handleAddBuildConfig}
                className="inline-flex items-center gap-1.5 rounded-md bg-white px-3 py-1.5 text-xs font-medium text-primary shadow-sm ring-1 ring-primary/20 hover:bg-slate-50 dark:bg-slate-900 dark:hover:bg-slate-800"
              >
                <Plus className="h-3 w-3" />
                Add Build Config
              </button>
              <button
                type="button"
                onClick={handleAddTask}
                className="inline-flex items-center gap-1.5 rounded-md bg-white px-3 py-1.5 text-xs font-medium text-primary shadow-sm ring-1 ring-primary/20 hover:bg-slate-50 dark:bg-slate-900 dark:hover:bg-slate-800"
              >
                <Plus className="h-3 w-3" />
                Add Ping/Traceroute Task
              </button>
              {hasSelection && buildConfigs.length > 0 && (
                <div className="flex items-center gap-1.5">
                  <span className="text-xs text-slate-700 dark:text-slate-300">
                    Assign build config to selected:
                  </span>
                  <select
                    className="rounded-md border border-slate-300 bg-white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-900"
                    defaultValue=""
                    onChange={(e) => {
                      const id = e.target.value;
                      if (id) {
                        handleAssignBuildConfigToSelected(id);
                        e.target.value = '';
                      }
                    }}
                  >
                    <option value="">Select config</option>
                    {buildConfigs.map((cfg) => (
                      <option key={cfg.id} value={cfg.id}>
                        {cfg.name || cfg.id}
                      </option>
                    ))}
                  </select>
                </div>
              )}
            </div>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={saveConfigsAndTasks}
                disabled={configSaving}
                className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text:white shadow-sm hover:bg-primary/90 disabled:opacity-60"
              >
                {configSaving ? 'Saving & Propagating...' : 'Save Config & Tasks'}
              </button>
            </div>
          </div>

          {configError && (
            <div className="rounded-md border border-red-300 bg-red-50 px-3 py-2 text-xs text-red-800 dark:border-red-800 dark:bg-red-950/70 dark:text-red-300">
              {configError}
            </div>
          )}
        </div>

        {/* Probes table with config metadata */}
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-200 dark:divide-slate-800">
              <thead className="bg-slate-50 dark:bg-slate-800/50">
                <tr>
                  <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-slate-900 dark:text-white sm:pl-6">
                    <input
                      className="h-4 w-4 rounded border-slate-300 bg:white text-primary dark:border-slate-600 dark:bg-slate-900 focus:ring-primary dark:focus:ring-offset-slate-900"
                      type="checkbox"
                      onChange={toggleSelectAll}
                      checked={
                        filteredTargets.length > 0 &&
                        selectedIds.size === filteredTargets.length
                      }
                      aria-label="Select all"
                    />
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    ID
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    Target Address
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    Alias
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    Platform / Arch
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    Status
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    Last Seen
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    Config ID
                  </th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white">
                    Config Version
                  </th>
                  <th className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200 bg-white dark:divide-slate-800 dark:bg-slate-900">
                {loading && (
                  <tr>
                    <td
                      colSpan={10}
                      className="px-3 py-6 text-center text-sm text-slate-500 dark:text-slate-400"
                    >
                      Loading targets...
                    </td>
                  </tr>
                )}
                {!loading && filteredTargets.length === 0 && (
                  <tr>
                    <td
                      colSpan={10}
                      className="px-3 py-6 text-center text-sm text-slate-500 dark:text-slate-400"
                    >
                      No probes/targets registered yet. Use &quot;Add New Target&quot; to register one.
                    </td>
                  </tr>
                )}
                {!loading &&
                  filteredTargets.map((target) => (
                    <tr key={target.id}>
                      <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
                        <input
                          className="h-4 w-4 rounded border-slate-300 bg:white text-primary dark:border-slate-600 dark:bg-slate-900 focus:ring-primary dark:focus:ring-offset-slate-900"
                          type="checkbox"
                          checked={selectedIds.has(target.id)}
                          onChange={() => toggleSelect(target.id)}
                        />
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm font-medium text-slate-900 dark:text-white">
                        {target.id}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-500 dark:text-slate-400">
                        {target.address}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-500 dark:text-slate-400">
                        {target.name || '—'}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-500 dark:text-slate-400">
                        {target.platform || '—'} {target.arch ? `(${target.arch})` : ''}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm">
                        <StatusBadge status={target.status || 'unknown'} />
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-500 dark:text-slate-400">
                        {target.lastSeen || '—'}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-xs text-slate-500 dark:text-slate-400">
                        {target.configId || '—'}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-xs text-slate-500 dark:text-slate-400">
                        {typeof target.configVersion === 'number'
                          ? target.configVersion
                          : '—'}
                      </td>
                      <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                        <button
                          onClick={() => openEditPanel(target)}
                          className="p-1 text-slate-500 hover:text-primary dark:text-slate-400 dark:hover:text-blue-400"
                          type="button"
                        >
                          <Pencil className="h-5 w-5" />
                        </button>
                        <button
                          onClick={() => handleDelete(target.id)}
                          className="p-1 text-slate-500 hover:text-error dark:text-slate-400 dark:hover:text-red-400"
                          type="button"
                        >
                          <Trash2 className="h-5 w-5" />
                        </button>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Build configurations editor */}
        {buildConfigs.length > 0 && (
          <section className="mt-6 space-y-2">
            <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
              Build Configurations
            </h2>
            <div className="space-y-2">
              {buildConfigs.map((cfg) => (
                <div
                  key={cfg.id}
                  className="flex flex-wrap items-center gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-xs dark:border-slate-700 dark:bg-slate-900"
                >
                  <input
                    className="min-w-[140px] flex-1 rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={cfg.name}
                    onChange={(e) =>
                      handleUpdateBuildConfig(cfg.id, { name: e.target.value })
                    }
                    placeholder="Config name"
                  />
                  <input
                    className="min-w-[200px] flex-1 rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={cfg.description || ''}
                    onChange={(e) =>
                      handleUpdateBuildConfig(cfg.id, { description: e.target.value })
                    }
                    placeholder="Description (targets, build, etc.)"
                  />
                  <button
                    type="button"
                    onClick={() => handleRemoveBuildConfig(cfg.id)}
                    className="inline-flex items-center rounded bg-transparent p-1 text-red-500 hover:text-red-600"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* Measurement tasks editor */}
        {tasks.length > 0 && (
          <section className="mt-6 space-y-2">
            <h2 className="text-sm font-semibold text-slate-800 dark:text-slate-100">
              Measurement Tasks (Ping / Traceroute)
            </h2>
            <div className="space-y-2">
              {tasks.map((task) => (
                <div
                  key={task.id}
                  className="flex flex-wrap items-center gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-xs dark:border-slate-700 dark:bg-slate-900"
                >
                  <input
                    className="min-w-[140px] flex-1 rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={task.name}
                    onChange={(e) =>
                      handleUpdateTask(task.id, { name: e.target.value })
                    }
                    placeholder="Task name"
                  />
                  <select
                    className="rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={task.type}
                    onChange={(e) =>
                      handleUpdateTask(task.id, {
                        type: e.target.value as MeasurementTaskType,
                      })
                    }
                  >
                    <option value="ping">Ping</option>
                    <option value="traceroute">Traceroute</option>
                  </select>
                  <input
                    className="min-w-[180px] flex-1 rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={task.targets?.[0]?.address || ''}
                    onChange={(e) => handleUpdateTaskTarget(task.id, e.target.value)}
                    placeholder="Target address (e.g., 8.8.8.8)"
                  />
                  <input
                    type="number"
                    className="w-20 rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={task.intervalSeconds}
                    onChange={(e) =>
                      handleUpdateTask(task.id, {
                        intervalSeconds: Number(e.target.value) || 60,
                      })
                    }
                    min={5}
                    placeholder="Interval"
                    title="Interval seconds"
                  />
                  <input
                    type="number"
                    className="w-16 rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={task.timeoutSeconds}
                    onChange={(e) =>
                      handleUpdateTask(task.id, {
                        timeoutSeconds: Number(e.target.value) || 5,
                      })
                    }
                    min={1}
                    placeholder="Timeout"
                    title="Timeout seconds"
                  />
                  <select
                    className="rounded border border-slate-300 bg:white px-2 py-1 text-xs dark:border-slate-700 dark:bg-slate-950"
                    value={task.buildConfigId || ''}
                    onChange={(e) =>
                      handleUpdateTask(task.id, {
                        buildConfigId: e.target.value || undefined,
                      })
                    }
                  >
                    <option value="">No build config</option>
                    {buildConfigs.map((cfg) => (
                      <option key={cfg.id} value={cfg.id}>
                        {cfg.name || cfg.id}
                      </option>
                    ))}
                  </select>
                  <label className="inline-flex items-center gap-1">
                    <input
                      type="checkbox"
                      checked={task.enabled}
                      onChange={(e) =>
                        handleUpdateTask(task.id, { enabled: e.target.checked })
                      }
                    />
                    <span>Enabled</span>
                  </label>
                  <button
                    type="button"
                    onClick={() => handleRemoveTask(task.id)}
                    className="inline-flex items-center rounded bg-transparent p-1 text-red-500 hover:text-red-600"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
          </section>
        )}
      </main>

      {/* Side panel for create/edit target */}
      {isPanelOpen && (
        <aside className="fixed inset-y-0 right-0 z-10 w-full max-w-md translate-x-0 border-l border-slate-200 bg-white shadow-xl transition-transform duration-300 ease-in-out dark:border-slate-800 dark:bg-slate-950">
          <div className="flex h-full flex-col">
            <div className="flex items-center justify-between border-b border-slate-200 p-4 dark:border-slate-800">
              <h2 className="text-lg font-semibold text-slate-900 dark:text-white">
                {editTarget ? 'Edit Target / Probe' : 'Add New Target / Probe'}
              </h2>
              <button
                onClick={closePanel}
                className="rounded-md p-1 text-slate-500 hover:bg-slate-100 hover:text-slate-600 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-300"
                type="button"
                disabled={isSubmitting}
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-6">
              {formError && (
                <div className="mb-4 rounded-md border border-red-300 bg-red-50 px-3 py-2 text-xs text-red-800 dark:border-red-800 dark:bg-red-950/70 dark:text-red-300">
                  {formError}
                </div>
              )}
              <form className="space-y-6" onSubmit={handleSubmit}>
                {!editTarget && (
                  <div>
                    <label
                      className="block text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400"
                      htmlFor="id"
                    >
                      ID (required)
                    </label>
                    <input
                      id="id"
                      name="id"
                      className="mt-1 block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary"
                      placeholder="Unique probe/target ID (e.g., probe-eu-west-1)"
                      value={form.id}
                      onChange={(e) => handleFormChange('id', e.target.value)}
                      required={!editTarget}
                    />
                  </div>
                )}
                <div>
                  <label
                    className="block text-sm font-medium text-slate-700 dark:text-slate-300"
                    htmlFor="target-address"
                  >
                    Target Address
                  </label>
                  <div className="mt-1">
                    <input
                      className="block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary"
                      id="target-address"
                      name="target-address"
                      placeholder="e.g., 192.168.1.100 or server.example.com"
                      type="text"
                      value={form.address}
                      onChange={(e) => handleFormChange('address', e.target.value)}
                    />
                  </div>
                </div>
                <div>
                  <label
                    className="block text-sm font-medium text-slate-700 dark:text-slate-300"
                    htmlFor="alias"
                  >
                    Alias / Friendly Name <span className="text-slate-400">(Optional)</span>
                  </label>
                  <div className="mt-1">
                    <input
                      className="block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary"
                      id="alias"
                      name="alias"
                      placeholder="e.g., Primary Database Probe"
                      type="text"
                      value={form.alias}
                      onChange={(e) => handleFormChange('alias', e.target.value)}
                    />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label
                      className="block text-sm font-medium text-slate-700 dark:text-slate-300"
                      htmlFor="platform"
                    >
                      Platform
                    </label>
                    <input
                      id="platform"
                      name="platform"
                      className="mt-1 block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary"
                      placeholder="e.g., linux, windows"
                      value={form.platform}
                      onChange={(e) => handleFormChange('platform', e.target.value)}
                    />
                  </div>
                  <div>
                    <label
                      className="block text-sm font-medium text-slate-700 dark:text-slate-300"
                      htmlFor="arch"
                    >
                      Architecture
                    </label>
                    <input
                      id="arch"
                      name="arch"
                      className="mt-1 block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary"
                      placeholder="e.g., amd64, arm64"
                      value={form.arch}
                      onChange={(e) => handleFormChange('arch', e.target.value)}
                    />
                  </div>
                </div>
              </form>
            </div>
            <div className="flex flex-shrink-0 justify-end gap-3 border-t border-slate-200 p-4 dark:border-slate-800">
              <button
                onClick={closePanel}
                className="rounded-lg border border-slate-300 bg:white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800"
                type="button"
                disabled={isSubmitting}
              >
                Cancel
              </button>
              <button
                onClick={handleSubmit}
                className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text:white shadow-sm hover:bg-primary/90 disabled:opacity-70"
                type="submit"
                disabled={isSubmitting}
              >
                {isSubmitting
                  ? editTarget
                    ? 'Saving...'
                    : 'Creating...'
                  : editTarget
                    ? 'Save Changes'
                    : 'Save Target'}
              </button>
            </div>
          </div>
        </aside>
      )}
    </div>
  );
};

export default Targets;