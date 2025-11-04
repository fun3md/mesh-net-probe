import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import {
  Activity,
  Settings,
  Users,
  Database,
  Bell,
  Menu,
  X,
  Home,
  BarChart3,
  Shield
} from 'lucide-react';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();

  const navigation = [
    { name: 'Dashboard', href: '/', icon: Home },
    { name: 'Probes', href: '/probes', icon: Activity },
    { name: 'Measurements', href: '/measurements', icon: BarChart3 },
    { name: 'Configuration', href: '/config', icon: Settings },
    { name: 'Users', href: '/users', icon: Users },
    { name: 'Database', href: '/database', icon: Database },
    { name: 'Security', href: '/security', icon: Shield },
  ];

  const isCurrentPath = (path: string) => {
    return location.pathname === path;
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-indigo-50/50">
      {/* Mobile sidebar */}
      <div className={`fixed inset-0 z-50 lg:hidden ${sidebarOpen ? 'block' : 'hidden'}`}>
        <div className="fixed inset-0 bg-gray-900/50 backdrop-blur-sm" onClick={() => setSidebarOpen(false)} />
        <div className="fixed inset-y-0 left-0 flex w-80 flex-col bg-white/95 backdrop-blur-xl border-r border-gray-200/50">
          <div className="flex h-16 items-center justify-between px-6 border-b border-gray-200/50">
            <div className="flex items-center">
              <div className="relative">
                <div className="absolute inset-0 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-lg blur opacity-75"></div>
                <Shield className="relative h-8 w-8 text-blue-600" />
              </div>
              <h1 className="ml-3 text-xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                Mesh Probe
              </h1>
            </div>
            <button 
              onClick={() => setSidebarOpen(false)}
              className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
          <nav className="flex-1 space-y-2 px-4 py-6">
            {navigation.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={`group flex items-center rounded-xl px-4 py-3 text-sm font-medium transition-all duration-200 ${
                    isCurrentPath(item.href)
                      ? 'bg-gradient-to-r from-blue-500 to-indigo-500 text-white shadow-lg'
                      : 'text-gray-600 hover:bg-gradient-to-r hover:from-blue-50 hover:to-indigo-50 hover:text-gray-900'
                  }`}
                  onClick={() => setSidebarOpen(false)}
                >
                  <Icon className={`mr-3 h-5 w-5 ${isCurrentPath(item.href) ? 'text-white' : 'text-gray-400 group-hover:text-blue-500'}`} />
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>
      </div>

      {/* Desktop sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-80 lg:flex-col">
        <div className="flex min-h-0 flex-1 flex-col bg-white/80 backdrop-blur-xl border-r border-gray-200/50">
          <div className="flex h-16 items-center px-6 border-b border-gray-200/50">
            <div className="flex items-center">
              <div className="relative">
                <div className="absolute inset-0 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-lg blur opacity-75"></div>
                <Shield className="relative h-8 w-8 text-blue-600" />
              </div>
              <h1 className="ml-3 text-xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
                Mesh Probe
              </h1>
            </div>
          </div>
          <nav className="flex-1 space-y-2 px-4 py-6">
            {navigation.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={`group flex items-center rounded-xl px-4 py-3 text-sm font-medium transition-all duration-200 ${
                    isCurrentPath(item.href)
                      ? 'bg-gradient-to-r from-blue-500 to-indigo-500 text-white shadow-lg transform scale-[1.02]'
                      : 'text-gray-600 hover:bg-gradient-to-r hover:from-blue-50 hover:to-indigo-50 hover:text-gray-900 hover:transform hover:scale-[1.01]'
                  }`}
                >
                  <Icon className={`mr-3 h-5 w-5 ${isCurrentPath(item.href) ? 'text-white' : 'text-gray-400 group-hover:text-blue-500'}`} />
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>
      </div>

      {/* Main content */}
      <div className="lg:pl-80">
        {/* Top bar */}
        <div className="sticky top-0 z-10 bg-white/80 backdrop-blur-xl border-b border-gray-200/50">
          <div className="flex h-16 items-center justify-between px-6 sm:px-8 lg:px-10">
            <button
              onClick={() => setSidebarOpen(true)}
              className="text-gray-500 hover:text-gray-700 lg:hidden p-2 rounded-lg hover:bg-gray-100 transition-colors"
            >
              <Menu className="h-6 w-6" />
            </button>
            
            <div className="flex items-center space-x-4">
              <button className="relative p-2 rounded-lg hover:bg-gray-100 transition-colors">
                <Bell className="h-6 w-6 text-gray-500" />
                <span className="absolute top-1 right-1 h-2 w-2 bg-red-500 rounded-full"></span>
              </button>
              <div className="flex items-center space-x-3">
                <div className="relative">
                  <div className="absolute inset-0 bg-gradient-to-r from-blue-500 to-indigo-500 rounded-full blur opacity-75"></div>
                  <div className="relative h-9 w-9 rounded-full bg-gradient-to-r from-blue-500 to-indigo-500 flex items-center justify-center">
                    <span className="text-white text-sm font-semibold">AD</span>
                  </div>
                </div>
                <div className="hidden sm:block">
                  <span className="text-sm font-semibold bg-gradient-to-r from-gray-700 to-gray-900 bg-clip-text text-transparent">
                    Admin User
                  </span>
                  <p className="text-xs text-gray-500">System Administrator</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Page content */}
        <main className="flex-1">
          <div className="py-8">
            <div className="mx-auto max-w-7xl px-6 sm:px-8 lg:px-10">
              {children}
            </div>
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout;