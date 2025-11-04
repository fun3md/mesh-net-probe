import React, { useState } from 'react';
import { Search, ListFilter, Plus, Pencil, Trash2, FlaskConical, Tag, X } from 'lucide-react';

const mockTargets = [
    { id: '192.168.1.1', type: 'IP Address', status: 'Valid', lastChecked: '2024-07-28 10:30 AM', selected: true },
    { id: 'server.example.com', type: 'Hostname', status: 'Warning', lastChecked: '2024-07-28 10:28 AM', selected: false },
    { id: '10.0.0.256', type: 'IP Address', status: 'Invalid', lastChecked: '2024-07-28 09:15 AM', selected: true },
    { id: 'db-primary.internal', type: 'Hostname', status: 'Valid', lastChecked: '2024-07-28 10:25 AM', selected: false },
    { id: 'unresolvable-host', type: 'Hostname', status: 'Invalid', lastChecked: '2024-07-28 09:12 AM', selected: true },
];

const StatusBadge: React.FC<{ status: 'Valid' | 'Warning' | 'Invalid' }> = ({ status }) => {
    const statusClasses = {
        Valid: 'bg-success/10 text-success dark:bg-success/20 dark:text-green-300',
        Warning: 'bg-warning/10 text-warning dark:bg-warning/20 dark:text-yellow-300',
        Invalid: 'bg-error/10 text-error dark:bg-error/20 dark:text-red-300',
    };
    const dotClasses = {
        Valid: 'bg-success',
        Warning: 'bg-warning',
        Invalid: 'bg-error',
    };
    return (
        <span className={`inline-flex items-center gap-1.5 rounded-full px-2 py-1 text-xs font-medium ${statusClasses[status]}`}>
            <span className={`h-1.5 w-1.5 rounded-full ${dotClasses[status]}`}></span>
            {status}
        </span>
    );
};

const Targets: React.FC = () => {
    const [isPanelOpen, setIsPanelOpen] = useState(true);

    return (
        <div className="relative flex h-auto min-h-screen w-full flex-col">
            <main className="flex-1 p-4 sm:p-6 lg:p-8">
                <div className="flex items-center justify-between pb-6">
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Manage Network Targets</h1>
                    <button onClick={() => setIsPanelOpen(true)} className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 dark:ring-offset-background-dark">
                        <Plus className="h-4 w-4" />
                        Add New Target
                    </button>
                </div>
                <div className="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div className="relative flex-1">
                        <Search className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 dark:text-slate-400 h-5 w-5" />
                        <input className="w-full rounded-lg border border-slate-300 bg-white py-2 pl-10 pr-4 text-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary" placeholder="Search targets..." type="text" />
                    </div>
                    <div className="flex items-center gap-2">
                        <div className="relative">
                            <button className="flex items-center gap-2 rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800">
                                <ListFilter className="h-4 w-4" />
                                Filter by Status
                            </button>
                        </div>
                    </div>
                </div>
                <div className="mb-4">
                    <div className="flex h-10 w-full items-center justify-start rounded-lg bg-slate-200 p-1 dark:bg-slate-800">
                        {['All (5)', 'Valid (2)', 'Warning (1)', 'Invalid (2)'].map((filter, index) => (
                            <label key={filter} className="flex h-full flex-1 cursor-pointer items-center justify-center overflow-hidden rounded-lg px-2 text-sm font-medium leading-normal text-slate-600 has-[:checked]:bg-white has-[:checked]:text-slate-900 has-[:checked]:shadow-sm dark:text-slate-400 dark:has-[:checked]:bg-slate-900 dark:has-[:checked]:text-white">
                                <span className="truncate">{filter}</span>
                                <input defaultChecked={index === 0} className="invisible w-0" name="status-filter" type="radio" value={filter.split(' ')[0]} />
                            </label>
                        ))}
                    </div>
                </div>
                <div className="mb-4 flex items-center justify-between rounded-lg bg-primary/10 p-3 dark:bg-primary/20">
                    <span className="text-sm font-medium text-primary dark:text-blue-300">3 items selected</span>
                    <div className="flex items-center gap-4">
                        <button className="flex items-center gap-1.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:text-primary dark:hover:text-white">
                            <FlaskConical className="h-4 w-4" />
                            Run Pre-checks
                        </button>
                        <button className="flex items-center gap-1.5 text-sm font-medium text-slate-700 dark:text-slate-300 hover:text-primary dark:hover:text-white">
                            <Tag className="h-4 w-4" />
                            Edit Tags
                        </button>
                        <button className="flex items-center gap-1.5 text-sm font-medium text-error dark:text-red-400 hover:text-red-700 dark:hover:text-red-300">
                            <Trash2 className="h-4 w-4" />
                            Delete Selected
                        </button>
                    </div>
                </div>
                <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-slate-200 dark:divide-slate-800">
                            <thead className="bg-slate-50 dark:bg-slate-800/50">
                                <tr>
                                    <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-slate-900 dark:text-white sm:pl-6" scope="col">
                                        <input className="h-4 w-4 rounded border-slate-300 bg-white text-primary dark:border-slate-600 dark:bg-slate-900 focus:ring-primary dark:focus:ring-offset-slate-900" type="checkbox" />
                                    </th>
                                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white" scope="col">Target</th>
                                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white" scope="col">Type</th>
                                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white" scope="col">Status</th>
                                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-slate-900 dark:text-white" scope="col">Last Checked</th>
                                    <th className="relative py-3.5 pl-3 pr-4 sm:pr-6" scope="col"><span className="sr-only">Actions</span></th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-200 bg-white dark:divide-slate-800 dark:bg-slate-900">
                                {mockTargets.map((target) => (
                                    <tr key={target.id}>
                                        <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
                                            <input defaultChecked={target.selected} className="h-4 w-4 rounded border-slate-300 bg-white text-primary dark:border-slate-600 dark:bg-slate-900 focus:ring-primary dark:focus:ring-offset-slate-900" type="checkbox" />
                                        </td>
                                        <td className="whitespace-nowrap px-3 py-4 text-sm font-medium text-slate-900 dark:text-white">{target.id}</td>
                                        <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-500 dark:text-slate-400">{target.type}</td>
                                        <td className="whitespace-nowrap px-3 py-4 text-sm"><StatusBadge status={target.status as 'Valid' | 'Warning' | 'Invalid'} /></td>
                                        <td className="whitespace-nowrap px-3 py-4 text-sm text-slate-500 dark:text-slate-400">{target.lastChecked}</td>
                                        <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                                            <button className="p-1 text-slate-500 hover:text-primary dark:text-slate-400 dark:hover:text-blue-400"><Pencil className="h-5 w-5" /></button>
                                            <button className="p-1 text-slate-500 hover:text-error dark:text-slate-400 dark:hover:text-red-400"><Trash2 className="h-5 w-5" /></button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </main>
            {isPanelOpen && (
                <aside className="fixed inset-y-0 right-0 z-10 w-full max-w-md transform translate-x-0 border-l border-slate-200 bg-white shadow-xl transition-transform duration-300 ease-in-out dark:border-slate-800 dark:bg-slate-950">
                    <div className="flex h-full flex-col">
                        <div className="flex items-center justify-between border-b border-slate-200 p-4 dark:border-slate-800">
                            <h2 className="text-lg font-semibold text-slate-900 dark:text-white">Add New Target</h2>
                            <button onClick={() => setIsPanelOpen(false)} className="rounded-md p-1 text-slate-500 hover:bg-slate-100 hover:text-slate-600 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-300" type="button">
                                <X className="h-5 w-5" />
                            </button>
                        </div>
                        <div className="flex-1 overflow-y-auto p-6">
                            <form className="space-y-6">
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 dark:text-slate-300" htmlFor="target-address">Target Address</label>
                                    <div className="mt-1">
                                        <input className="block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary" id="target-address" name="target-address" placeholder="e.g., 192.168.1.100 or server.com" type="text" />
                                    </div>
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 dark:text-slate-300" htmlFor="alias">Alias / Friendly Name <span className="text-slate-400">(Optional)</span></label>
                                    <div className="mt-1">
                                        <input className="block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary" id="alias" name="alias" placeholder="e.g., Primary Database Server" type="text" />
                                    </div>
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 dark:text-slate-300" htmlFor="target-type">Target Type</label>
                                    <div className="mt-1">
                                        <select className="block w-full rounded-lg border-slate-300 text-sm shadow-sm dark:border-slate-700 dark:bg-slate-900 focus:border-primary focus:ring-primary" id="target-type" name="target-type">
                                            <option>IP Address</option>
                                            <option>Hostname</option>
                                        </select>
                                    </div>
                                </div>
                            </form>
                        </div>
                        <div className="flex flex-shrink-0 justify-end gap-3 border-t border-slate-200 p-4 dark:border-slate-800">
                            <button onClick={() => setIsPanelOpen(false)} className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-300 dark:hover:bg-slate-800" type="button">Cancel</button>
                            <button className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-primary/90" type="submit">Save Target</button>
                        </div>
                    </div>
                </aside>
            )}
        </div>
    );
};

export default Targets;