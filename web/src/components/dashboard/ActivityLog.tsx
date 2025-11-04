import React from 'react';
import { CheckCircle, Info, AlertTriangle } from 'lucide-react';

const mockActivity = [
    { icon: AlertTriangle, color: 'text-error', text: 'Probe #30 disconnected', time: '15 min ago' },
    { icon: AlertTriangle, color: 'text-warning', text: 'High latency detected on Probe #25', time: '2 min ago' },
    { icon: CheckCircle, color: 'text-success', text: 'Probe #101-US-EAST back online', time: 'Just now' },
    { icon: Info, color: 'text-text-light-secondary dark:text-text-dark-secondary', text: 'System configuration updated by admin', time: '30 sec ago' },
];

const ActivityLog: React.FC = () => {
    return (
        <div className="rounded-xl border border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark flex flex-col">
            <div className="p-5 border-b border-border-light dark:border-border-dark">
                <h3 className="text-lg font-semibold text-text-light-primary dark:text-text-dark-primary">Real-Time Activity Log</h3>
            </div>
            <div className="flex-1 p-5 space-y-4 overflow-y-auto max-h-56">
                {mockActivity.map((item, index) => (
                    <div key={index} className="flex items-start gap-3">
                        <item.icon className={`text-xl mt-0.5 ${item.color}`} />
                        <div>
                            <p className="font-medium">{item.text}</p>
                            <p className="text-sm text-text-light-secondary dark:text-text-dark-secondary">{item.time}</p>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default ActivityLog;
