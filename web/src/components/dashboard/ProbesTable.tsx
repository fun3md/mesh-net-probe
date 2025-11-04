import React from 'react';
import { ArrowRight } from 'lucide-react';

const mockProbes = [
    { id: 'Probe #101-US-EAST', status: 'success', latency: 32, packetLoss: '0.1%', uptime: '99.98%', location: 'New York, USA' },
    { id: 'Probe #102-US-WEST', status: 'success', latency: 78, packetLoss: '0.0%', uptime: '100%', location: 'California, USA' },
    { id: 'Probe #25-EU-CENTRAL', status: 'warning', latency: 120, packetLoss: '1.5%', uptime: '98.5%', location: 'Frankfurt, DE' },
    { id: 'Probe #30-APAC-SE', status: 'error', latency: '--', packetLoss: '--', uptime: '95.2%', location: 'Singapore' },
    { id: 'Probe #121-US-CENTRAL', status: 'success', latency: 45, packetLoss: '0.2%', uptime: '99.95%', location: 'Texas, USA' },
    { id: 'Probe #55-EU-WEST', status: 'success', latency: 95, packetLoss: '0.0%', uptime: '99.99%', location: 'London, UK' },
    { id: 'Probe #87-SA-EAST', status: 'success', latency: 150, packetLoss: '0.3%', uptime: '99.8%', location: 'São Paulo, BR' },
];

const StatusIndicator: React.FC<{ status: 'success' | 'warning' | 'error' }> = ({ status }) => {
    const colorClass = {
        success: 'bg-success',
        warning: 'bg-warning',
        error: 'bg-error',
    }[status];
    return <div className={`h-3 w-3 rounded-full ${colorClass}`}></div>;
};

const ProbesTable: React.FC = () => {
    return (
        <div className="lg:col-span-3 rounded-xl border border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark overflow-hidden flex flex-col">
            <div className="p-5 border-b border-border-light dark:border-border-dark flex items-center justify-between">
                <h3 className="text-lg font-semibold text-text-light-primary dark:text-text-dark-primary">All Probes Detailed Statistics</h3>
                <button className="text-primary text-sm font-medium flex items-center gap-1">
                    <span>View All</span>
                    <ArrowRight className="h-4 w-4" />
                </button>
            </div>
            <div className="overflow-x-auto flex-1">
                <table className="w-full text-left">
                    <thead className="border-b border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark/50 sticky top-0">
                        <tr>
                            <th className="p-4 text-sm font-semibold text-text-light-secondary dark:text-text-dark-secondary">Status</th>
                            <th className="p-4 text-sm font-semibold text-text-light-secondary dark:text-text-dark-secondary">Probe Name/ID</th>
                            <th className="p-4 text-sm font-semibold text-text-light-secondary dark:text-text-dark-secondary">Latency (ms)</th>
                            <th className="p-4 text-sm font-semibold text-text-light-secondary dark:text-text-dark-secondary">Packet Loss</th>
                            <th className="p-4 text-sm font-semibold text-text-light-secondary dark:text-text-dark-secondary">Uptime</th>
                            <th className="p-4 text-sm font-semibold text-text-light-secondary dark:text-text-dark-secondary">Location</th>
                        </tr>
                    </thead>
                    <tbody>
                        {mockProbes.map((probe) => (
                            <tr key={probe.id} className="border-b border-border-light dark:border-border-dark hover:bg-background-light dark:hover:bg-background-dark/50">
                                <td className="p-4"><StatusIndicator status={probe.status as 'success' | 'warning' | 'error'} /></td>
                                <td className="p-4 font-medium text-primary cursor-pointer">{probe.id}</td>
                                <td className="p-4">{probe.latency}</td>
                                <td className="p-4">{probe.packetLoss}</td>
                                <td className="p-4">{probe.uptime}</td>
                                <td className="p-4">{probe.location}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default ProbesTable;
