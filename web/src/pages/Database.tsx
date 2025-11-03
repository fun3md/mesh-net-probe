import React from 'react';
import { Database as DatabaseIcon } from 'lucide-react';

const Database: React.FC = () => {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Database</h1>
        <p className="text-gray-600">Database management and monitoring</p>
      </div>

      <div className="bg-white p-6 rounded-lg shadow">
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <DatabaseIcon className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">Database Management</h3>
            <p className="mt-1 text-sm text-gray-500">
              Database monitoring and management tools will appear here.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Database;