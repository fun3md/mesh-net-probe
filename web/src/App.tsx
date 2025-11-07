import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Dashboard from '@/pages/Dashboard';
import Login from '@/pages/Login';
import Targets from '@/pages/Targets';
import NetworkConfiguration from '@/pages/NetworkConfiguration';
import VisualConfigurator from '@/pages/VisualConfigurator';
import OpenTelemetryConfiguration from '@/pages/OpenTelemetryConfiguration';
import Layout from '@/components/Layout';
import { useAuth } from '@/hooks/useAuth';

const ProtectedRoute: React.FC<{ children: React.ReactElement }> = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  // While auth is being resolved, avoid flicker and do not redirect yet
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950 text-slate-200">
        <div className="animate-pulse text-center">
          <div className="text-xl font-semibold mb-2">Checking authentication...</div>
          <div className="text-sm text-slate-400">Initializing admin web session</div>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
};

export const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        {/* Public login route */}
        <Route path="/login" element={<Login />} />

        {/* Protected application shell */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout>
                <Outlet />
              </Layout>
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="targets" element={<Targets />} />
          <Route path="config" element={<NetworkConfiguration />} />
          <Route path="visual-configurator" element={<VisualConfigurator />} />
          <Route path="otel-config" element={<OpenTelemetryConfiguration />} />
        </Route>

        {/* Fallback: any unknown route -> guarded dashboard */}
        <Route
          path="*"
          element={
            <ProtectedRoute>
              <Layout>
                <Dashboard />
              </Layout>
            </ProtectedRoute>
          }
        />
      </Routes>
    </Router>
  );
};