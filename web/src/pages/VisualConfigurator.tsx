import React from 'react';
import { Share2, Search, ChevronDown, Globe, Dna, Play, Trash2, ChevronRight, XCircle } from 'lucide-react';

const VisualConfigurator: React.FC = () => {
    return (
        <div className="flex h-screen w-full flex-col font-display bg-background-light dark:bg-background-dark text-text-light dark:text-text-dark">
            <header className="flex shrink-0 items-center justify-between border-b border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark p-4">
                <div className="flex items-center gap-4">
                    <Share2 className="text-text-light dark:text-text-dark text-3xl" />
                    <h1 className="text-xl font-bold">Probe Configuration Editor</h1>
                </div>
                <div className="flex items-center gap-4">
                    <label className="flex flex-col">
                        <p className="text-xs font-medium text-text-muted-light dark:text-text-muted-dark pb-1">Configuration Name</p>
                        <input className="form-input w-72 rounded border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark focus:border-primary focus:ring-primary h-9 placeholder:text-text-muted-light dark:placeholder:text-text-muted-dark text-sm" placeholder="e.g., Main Web Server Health Check" value="" />
                    </label>
                    <div className="flex gap-2">
                        <button className="flex min-w-[84px] cursor-pointer items-center justify-center overflow-hidden rounded h-9 px-4 bg-primary text-white text-sm font-bold leading-normal tracking-[0.015em] hover:bg-primary/90">
                            <span className="truncate">Save</span>
                        </button>
                        <button className="flex min-w-[84px] cursor-pointer items-center justify-center overflow-hidden rounded h-9 px-4 bg-gray-200 dark:bg-gray-700 text-text-light dark:text-text-dark text-sm font-bold leading-normal tracking-[0.015em] hover:bg-gray-300 dark:hover:bg-gray-600">
                            <span className="truncate">Save as Template</span>
                        </button>
                    </div>
                </div>
            </header>
            <main className="flex min-h-0 flex-1">
                <aside className="flex w-1/4 max-w-sm flex-col border-r border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark">
                    <div className="p-4 border-b border-border-light dark:border-border-dark">
                        <label className="flex flex-col">
                            <div className="flex w-full flex-1 items-stretch rounded-lg h-10">
                                <div className="text-text-muted-light dark:text-text-muted-dark flex border-none bg-background-light dark:bg-background-dark items-center justify-center pl-3 rounded-l-lg border-r-0">
                                    <Search className="text-xl" />
                                </div>
                                <input className="form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-r-lg text-sm focus:outline-0 focus:ring-0 border-none bg-background-light dark:bg-background-dark h-full placeholder:text-text-muted-light dark:placeholder:text-text-muted-dark" placeholder="Search templates & components" value="" />
                            </div>
                        </label>
                    </div>
                    <div className="flex-1 overflow-y-auto p-4">
                        <div className="flex flex-col gap-3">
                            <details className="flex flex-col rounded-lg bg-background-light dark:bg-background-dark group" open>
                                <summary className="flex cursor-pointer items-center justify-between gap-6 p-3">
                                    <p className="text-sm font-medium">Input Sources</p>
                                    <ChevronDown className="text-xl group-open:rotate-180 transition-transform" />
                                </summary>
                                <div className="p-3 pt-0 space-y-2">
                                    <div className="flex items-start gap-3 rounded-lg border border-border-light dark:border-border-dark p-3 cursor-grab bg-panel-light dark:bg-panel-dark hover:border-primary">
                                        <Globe className="text-primary text-xl mt-0.5" />
                                        <div>
                                            <p className="font-medium text-sm">HTTP Request</p>
                                            <p className="text-xs text-text-muted-light dark:text-text-muted-dark">Fetches data from a URL.</p>
                                        </div>
                                    </div>
                                    <div className="flex items-start gap-3 rounded-lg border border-border-light dark:border-border-dark p-3 cursor-grab bg-panel-light dark:bg-panel-dark hover:border-primary">
                                        <Dna className="text-primary text-xl mt-0.5" />
                                        <div>
                                            <p className="font-medium text-sm">DNS Query</p>
                                            <p className="text-xs text-text-muted-light dark:text-text-muted-dark">Performs a DNS lookup.</p>
                                        </div>
                                    </div>
                                </div>
                            </details>
                        </div>
                    </div>
                </aside>
                <section className="flex flex-1 flex-col">
                    <div className="flex shrink-0 items-center justify-between border-b border-border-light dark:border-border-dark p-2">
                        <nav className="flex items-center gap-2 text-sm text-text-muted-light dark:text-text-muted-dark px-2">
                            <span>Root</span>
                            <ChevronRight className="text-base" />
                            <span className="text-text-light dark:text-text-dark font-medium">HTTP_Health_Check</span>
                        </nav>
                        <div className="flex items-center gap-2">
                            <button className="flex items-center gap-2 min-w-[84px] cursor-pointer justify-center overflow-hidden rounded h-9 px-3 bg-primary/20 text-primary text-sm font-bold hover:bg-primary/30">
                                <Play className="text-xl" />
                                <span className="truncate">Validate</span>
                            </button>
                            <button className="flex items-center gap-2 min-w-[84px] cursor-pointer justify-center overflow-hidden rounded h-9 px-3 bg-gray-200 dark:bg-gray-700 text-text-light dark:text-text-dark text-sm font-bold hover:bg-gray-300 dark:hover:bg-gray-600">
                                <Trash2 className="text-xl" />
                                <span className="truncate">Clear Canvas</span>
                            </button>
                        </div>
                    </div>
                    <div className="flex-1 overflow-auto p-8 bg-background-light dark:bg-background-dark">
                        <div className="flex flex-col items-center gap-6 rounded-lg border-2 border-dashed border-border-light dark:border-border-dark px-6 py-14">
                            <div className="flex max-w-[480px] flex-col items-center gap-2">
                                <p className="text-lg font-bold text-center">Canvas Area</p>
                                <p className="text-sm text-center text-text-muted-light dark:text-text-muted-dark">Drag components from the library to start building your configuration.</p>
                            </div>
                        </div>
                    </div>
                </section>
                <aside className="flex w-1/3 max-w-xl flex-col border-l border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark">
                    <div className="flex shrink-0 items-center justify-between border-b border-border-light dark:border-border-dark p-3">
                        <h2 className="text-base font-semibold">Live JSON Output</h2>
                        <div className="flex items-center gap-2 text-sm font-medium text-error">
                            <XCircle className="text-xl" />
                            <span>Status: Invalid (3 errors)</span>
                        </div>
                    </div>
                    <div className="flex-1 overflow-hidden font-code text-sm">
                        <div className="h-full w-full bg-background-light dark:bg-background-dark p-4 overflow-auto text-text-muted-light dark:text-text-dark">
                            <pre className="text-xs leading-relaxed"><code>...</code></pre>
                        </div>
                    </div>
                    <div className="shrink-0 border-t border-border-light dark:border-border-dark p-3">
                        <p className="font-semibold text-sm mb-2">Validation Errors</p>
                    </div>
                </aside>
            </main>
        </div>
    );
};

export default VisualConfigurator;
