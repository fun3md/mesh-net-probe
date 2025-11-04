import React from 'react';
import { ArrowLeft } from 'lucide-react';

const NetworkConfiguration: React.FC = () => {
    return (
        <div className="relative flex min-h-screen w-full flex-col bg-background-light dark:bg-background-dark font-display text-text-light dark:text-text-dark">
            <header className="flex items-center border-b border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark p-4 sticky top-0 z-10">
                <ArrowLeft className="text-text-secondary-light dark:text-text-secondary-dark mr-4" />
                <h1 className="text-lg font-bold leading-tight tracking-tight">Network Configuration</h1>
            </header>
            <main className="flex-1 p-4 sm:p-6 lg:p-8">
                <div className="mx-auto max-w-3xl space-y-8">
                    <section className="rounded-xl border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark shadow-sm">
                        <header className="p-4 sm:p-6 border-b border-border-light dark:border-border-dark">
                            <h2 className="text-lg font-bold leading-tight tracking-tight">Interface &amp; Connectivity</h2>
                        </header>
                        <div className="p-4 sm:p-6 space-y-6">
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-2 md:gap-4 items-center">
                                <label className="text-sm font-medium text-text-secondary-light dark:text-text-secondary-dark" htmlFor="network-interface">Network Interface</label>
                                <div className="md:col-span-2">
                                    <select className="form-select w-full rounded-lg border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-light dark:text-text-dark focus:border-primary focus:ring-primary focus:ring-opacity-50" id="network-interface">
                                        <option>Ethernet 1</option>
                                        <option>Wi-Fi</option>
                                        <option>Bluetooth PAN</option>
                                    </select>
                                    <p className="mt-2 text-xs text-text-secondary-light dark:text-text-secondary-dark">Select the primary network interface for the probe system.</p>
                                </div>
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-2 md:gap-4 items-center">
                                <label className="text-sm font-medium text-text-secondary-light dark:text-text-secondary-dark" htmlFor="ttl">Time to Live (TTL)</label>
                                <div className="md:col-span-2">
                                    <input className="form-input w-full rounded-lg border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-light dark:text-text-dark focus:border-primary focus:ring-primary focus:ring-opacity-50" id="ttl" type="number" defaultValue="64" />
                                    <p className="mt-2 text-xs text-text-secondary-light dark:text-text-secondary-dark">Sets the maximum number of hops a packet can take.</p>
                                </div>
                            </div>
                        </div>
                    </section>
                    <section className="rounded-xl border border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark shadow-sm">
                        <header className="p-4 sm:p-6 border-b border-border-light dark:border-border-dark">
                            <h2 className="text-lg font-bold leading-tight tracking-tight">Packet &amp; Buffer Settings</h2>
                        </header>
                        <div className="p-4 sm:p-6">
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-2 md:gap-4 items-center">
                                <label className="text-sm font-medium text-text-secondary-light dark:text-text-secondary-dark" htmlFor="buffer-size">Buffer Size (KB)</label>
                                <div className="md:col-span-2">
                                    <input className="form-input w-full rounded-lg border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-light dark:text-text-dark focus:border-primary focus:ring-primary focus:ring-opacity-50" id="buffer-size" type="number" defaultValue="1024" />
                                    <p className="mt-2 text-xs text-text-secondary-light dark:text-text-secondary-dark">Defines the size of the packet buffer.</p>
                                </div>
                            </div>
                        </div>
                    </section>
                </div>
            </main>
            <footer className="sticky bottom-0 mt-auto bg-card-light/80 dark:bg-card-dark/80 backdrop-blur-sm border-t border-border-light dark:border-border-dark p-4">
                <div className="mx-auto flex max-w-3xl items-center justify-end gap-4">
                    <button className="rounded-lg px-4 py-2.5 text-sm font-semibold text-text-secondary-light dark:text-text-secondary-dark hover:bg-gray-100 dark:hover:bg-white/10">Reset to Default</button>
                    <button className="rounded-lg bg-primary px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:ring-offset-2 dark:focus:ring-offset-background-dark disabled:opacity-50" disabled>Save Changes</button>
                </div>
            </footer>
        </div>
    );
};

export default NetworkConfiguration;
