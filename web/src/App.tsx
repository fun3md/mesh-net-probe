import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Layout from '@/components/Layout';
import Dashboard from '@/pages/Dashboard';
import Login from '@/pages/Login';
import { useAuth } from '@/hooks/useAuth';
import Targets from '@/pages/Targets';
import VisualConfigurator from '@/pages/VisualConfigurator';
import NetworkConfiguration from '@/pages/NetworkConfiguration';
import OpenTelemetryConfiguration from '@/pages/OpenTelemetryConfiguration';

const App: React.FC = () => {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Login />;
  }

  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/targets" element={<Targets />} />
          <Route path="/visual-configurator" element={<VisualConfigurator />} />
          <Route path="/network-configuration" element={<NetworkConfiguration />} />
          <Route path="/opentelemetry-configuration" element={<OpenTelemetryConfiguration />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </Router>
  );
};

export default App;