import React from 'react';
import { Settings, Plus, Pencil, Trash2 } from 'lucide-react';

const OpenTelemetryConfiguration: React.FC = () => {
    return (
        <div className="bg-background-light dark:bg-background-dark font-display">
            <div className="flex h-auto min-h-screen w-full flex-col">
                <div className="flex items-center bg-card-light dark:bg-card-dark p-4 border-b border-border-light dark:border-border-dark">
                    <div className="text-text-light dark:text-text-dark flex size-10 shrink-0 items-center justify-center">
                        <Settings className="text-3xl" />
                    </div>
                    <h1 className="text-text-light dark:text-text-dark text-xl font-bold leading-tight tracking-tight flex-1 ml-2">OpenTelemetry Configuration</h1>
                </div>
                <div className="border-b border-border-light dark:border-border-dark px-4 bg-card-light dark:bg-card-dark">
                    <div className="flex gap-8">
                        <a className="flex flex-col items-center justify-center border-b-[3px] border-b-primary text-primary pb-[13px] pt-4" href="#">
                            <p className="text-sm font-bold leading-normal">General/Endpoints</p>
                        </a>
                    </div>
                </div>
                <main className="flex-1 p-4 sm:p-6 lg:p-8 space-y-8">
                    <div id="general-endpoints">
                        <div className="max-w-4xl mx-auto space-y-8">
                            <section>
                                <h3 className="text-text-light dark:text-text-dark text-lg font-bold leading-tight tracking-tight pb-4">OTLP Endpoint Configuration</h3>
                                <div className="bg-card-light dark:bg-card-dark p-6 rounded-lg border border-border-light dark:border-border-dark space-y-6">
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        <label className="flex flex-col">
                                            <p className="text-text-light dark:text-text-dark text-sm font-medium leading-normal pb-2">OTLP/gRPC Endpoint</p>
                                            <input className="form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-md text-text-light dark:text-text-dark focus:outline-none focus:ring-2 focus:ring-primary border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark h-12 placeholder:text-text-secondary-light dark:placeholder:text-text-secondary-dark p-3 text-sm font-normal leading-normal" placeholder="e.g., otel-collector:4317" value="" />
                                        </label>
                                        <label className="flex flex-col">
                                            <p className="text-text-light dark:text-text-dark text-sm font-medium leading-normal pb-2">OTLP/HTTP Endpoint</p>
                                            <input className="form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-md text-text-light dark:text-text-dark focus:outline-none focus:ring-2 focus:ring-primary border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark h-12 placeholder:text-text-secondary-light dark:placeholder:text-text-secondary-dark p-3 text-sm font-normal leading-normal" placeholder="e.g., https://otel-collector:4318" value="https://otel-collector:4318" />
                                        </label>
                                    </div>
                                </div>
                            </section>
                            <section>
                                <div className="flex items-center justify-between pb-4">
                                    <h3 className="text-text-light dark:text-text-dark text-lg font-bold leading-tight tracking-tight">Custom HTTP Headers</h3>
                                    <button className="flex items-center gap-2 rounded-md bg-primary/20 dark:bg-primary/30 px-4 py-2 text-sm font-semibold text-primary hover:bg-primary/30 dark:hover:bg-primary/40">
                                        <Plus className="text-lg" />
                                        Add New Header
                                    </button>
                                </div>
                                <div className="bg-card-light dark:bg-card-dark rounded-lg border border-border-light dark:border-border-dark overflow-hidden">
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-left text-sm">
                                            <thead className="bg-background-light dark:bg-background-dark text-xs text-text-secondary-light dark:text-text-secondary-dark uppercase">
                                                <tr>
                                                    <th className="px-6 py-3 font-medium" scope="col">Header Name</th>
                                                    <th className="px-6 py-3 font-medium" scope="col">Value</th>
                                                    <th className="px-6 py-3 font-medium text-right" scope="col">Actions</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                <tr className="border-b border-border-light dark:border-border-dark">
                                                    <td className="px-6 py-4 font-mono text-text-light dark:text-text-dark">X-Auth-Token</td>
                                                    <td className="px-6 py-4 font-mono text-text-light dark:text-text-dark">********</td>
                                                    <td className="px-6 py-4 text-right">
                                                        <div className="flex justify-end gap-4">
                                                            <button className="text-text-secondary-light dark:text-text-secondary-dark hover:text-primary"><Pencil className="text-xl" /></button>
                                                            <button className="text-text-secondary-light dark:text-text-secondary-dark hover:text-error"><Trash2 className="text-xl" /></button>
                                                        </div>
                                                    </td>
                                                </tr>
                                                <tr className="border-b border-border-light dark:border-border-dark">
                                                    <td className="px-6 py-4 font-mono text-text-light dark:text-text-dark">X-Tenant-ID</td>
                                                    <td className="px-6 py-4 font-mono text-text-light dark:text-text-dark">prod-12345</td>
                                                    <td className="px-6 py-4 text-right">
                                                        <div className="flex justify-end gap-4">
                                                            <button className="text-text-secondary-light dark:text-text-secondary-dark hover:text-primary"><Pencil className="text-xl" /></button>
                                                            <button className="text-text-secondary-light dark:text-text-secondary-dark hover:text-error"><Trash2 className="text-xl" /></button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            </section>
                        </div>
                    </div>
                </main>
                <footer className="sticky bottom-0 bg-card-light/80 dark:bg-card-dark/80 backdrop-blur-sm border-t border-border-light dark:border-border-dark p-4">
                    <div className="max-w-4xl mx-auto flex justify-end gap-4">
                        <button className="rounded-md bg-transparent px-4 py-2 text-sm font-semibold text-text-light dark:text-text-dark border border-border-light dark:border-border-dark hover:bg-background-light dark:hover:bg-background-dark">Cancel</button>
                        <button className="rounded-md bg-primary px-4 py-2 text-sm font-semibold text-white hover:bg-primary/90 disabled:bg-primary/50 disabled:cursor-not-allowed">Save Changes</button>
                    </div>
                </footer>
            </div>
        </div>
    );
};

export default OpenTelemetryConfiguration;
