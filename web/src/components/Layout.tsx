import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  Target,
  FileJson,
  Network,
  Settings,
  Bell,
  UserCircle,
  Share2
} from 'lucide-react';

interface LayoutProps {
  children: React.ReactNode;
}

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Targets', href: '/targets', icon: Target },
  { name: 'Visual Configurator', href: '/visual-configurator', icon: FileJson },
  { name: 'Network Config', href: '/network-configuration', icon: Network },
  { name: 'OpenTelemetry Config', href: '/opentelemetry-configuration', icon: Settings },
];

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="flex h-screen bg-background-light dark:bg-background-dark text-text-light-primary dark:text-text-dark-primary">
      {/* Sidebar */}
      <aside className="flex w-16 flex-col items-center bg-panel-light dark:bg-panel-dark border-r border-border-light dark:border-border-dark py-4">
        <div className="text-primary flex size-12 shrink-0 items-center justify-center text-3xl">
          <Share2 />
        </div>
        <nav className="flex flex-col items-center gap-4 mt-8">
          {navigation.map((item) => (
            <NavLink
              key={item.name}
              to={item.href}
              className={({ isActive }) =>
                `flex h-12 w-12 items-center justify-center rounded-lg transition-colors ${
                  isActive
                    ? 'bg-primary/20 text-primary'
                    : 'text-text-light-secondary dark:text-text-dark-secondary hover:bg-primary/10 hover:text-primary'
                }`
              }
            >
              <item.icon className="h-6 w-6" />
            </NavLink>
          ))}
        </nav>
        <div className="mt-auto flex flex-col items-center gap-4">
            <button className="flex h-12 w-12 items-center justify-center rounded-lg text-text-light-secondary dark:text-text-dark-secondary hover:bg-primary/10 hover:text-primary">
                <Bell className="h-6 w-6" />
            </button>
            <button className="flex h-12 w-12 items-center justify-center rounded-full text-text-light-secondary dark:text-text-dark-secondary text-2xl">
                <UserCircle className="h-8 w-8" />
            </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
    </div>
  );
};

export default Layout;