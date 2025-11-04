import React from 'react';

const PacketLossChart: React.FC = () => {
    return (
        <div className="flex flex-col gap-2 rounded-xl border border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark p-5">
            <p className="text-base font-medium text-text-light-secondary dark:text-text-dark-secondary">System Packet Loss (%)</p>
            <p className="text-text-light-primary dark:text-text-dark-primary tracking-tight text-3xl font-bold truncate">0.8%</p>
            <div className="flex gap-1">
                <p className="text-text-light-secondary dark:text-text-dark-secondary text-sm">Last 24 Hours</p>
                <p className="text-error text-sm font-medium">-0.2%</p>
            </div>
            <div className="grid min-h-[160px] grid-flow-col gap-6 grid-rows-[1fr_auto] items-end justify-items-center px-3 pt-4">
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '100%' }}></div>
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '40%' }}></div>
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '80%' }}></div>
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '60%' }}></div>
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '90%' }}></div>
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '50%' }}></div>
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '70%' }}></div>
                <div className="bg-primary/30 rounded-t w-full" style={{ height: '30%' }}></div>
            </div>
        </div>
    );
};

export default PacketLossChart;
