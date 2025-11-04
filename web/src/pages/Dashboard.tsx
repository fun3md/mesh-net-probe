import React from 'react';
import StatCards from '../components/dashboard/StatCards';
import ProbesTable from '../components/dashboard/ProbesTable';
import SystemHealthMap from '../components/dashboard/SystemHealthMap';
import ActivityLog from '../components/dashboard/ActivityLog';
import LatencyChart from '../components/dashboard/LatencyChart';
import PacketLossChart from '../components/dashboard/PacketLossChart';
import { Search } from 'lucide-react';

const Dashboard: React.FC = () => {
    return (
        <div className="p-6 xl:p-8 space-y-6">
            <header className="flex flex-col md:flex-row items-center justify-between gap-4">
                <h1 className="text-2xl font-bold text-text-light-primary dark:text-text-dark-primary">Mesh Probe System Dashboard</h1>
                <div className="w-full md:w-auto md:min-w-80">
                    <label className="flex flex-col h-11 w-full">
                        <div className="flex w-full flex-1 items-stretch rounded-lg h-full">
                            <div className="text-text-light-secondary dark:text-text-dark-secondary flex border border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark items-center justify-center pl-3 rounded-l-lg border-r-0">
                                <Search className="text-xl" />
                            </div>
                            <input className="form-input flex w-full min-w-0 flex-1 resize-none overflow-hidden rounded-lg text-text-light-primary dark:text-text-dark-primary focus:outline-0 focus:ring-2 focus:ring-primary border border-border-light dark:border-border-dark bg-panel-light dark:bg-panel-dark h-full placeholder:text-text-light-secondary dark:placeholder:text-text-dark-secondary px-4 rounded-l-none border-l-0 pl-2 text-base font-normal leading-normal" placeholder="Find a probe by name or ID" value="" />
                        </div>
                    </label>
                </div>
            </header>

            <StatCards />

            <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
                <ProbesTable />
                <div className="lg:col-span-2 flex flex-col gap-6">
                    <SystemHealthMap />
                    <ActivityLog />
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 gap-6">
                <LatencyChart />
                <PacketLossChart />
            </div>
        </div>
    );
};

export default Dashboard;
