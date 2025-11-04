import React from 'react';

const StatCards: React.FC = () => {
    const stats = [
        { title: 'Total Probes', value: '1,256', color: 'primary' },
        { title: 'Active / Online', value: '1,198', color: 'success' },
        { title: 'Inactive / Offline', value: '58', color: 'error' },
        { title: 'Probes with Alerts', value: '15', color: 'warning' },
    ];

    return (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {stats.map(stat => (
                <div key={stat.title} className="flex flex-col gap-2 rounded-xl p-5 bg-panel-light dark:bg-panel-dark border border-border-light dark:border-border-dark">
                    <p className="text-base font-medium text-text-light-secondary dark:text-text-dark-secondary">{stat.title}</p>
                    <p className={`tracking-tight text-3xl font-bold ${stat.color === 'primary' ? 'text-text-light-primary dark:text-text-dark-primary' : stat.color === 'error' ? 'text-error' : stat.color === 'warning' ? 'text-warning' : 'text-success'}`}>{stat.value}</p>
                </div>
            ))}
        </div>
    );
};

export default StatCards;
