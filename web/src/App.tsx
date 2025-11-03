import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Layout from '@/components/Layout';
import Dashboard from '@/pages/Dashboard';
import Probes from '@/pages/Probes';
import Measurements from '@/pages/Measurements';
import Configuration from '@/pages/Configuration';
import Users from '@/pages/Users';
import Database from '@/pages/Database';
import Security from '@/pages/Security';
import Login from '@/pages/Login';
import { useAuth } from '@/hooks/useAuth';

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
          <Route path="/probes" element={<Probes />} />
          <Route path="/measurements" element={<Measurements />} />
          <Route path="/config" element={<Configuration />} />
          <Route path="/users" element={<Users />} />
          <Route path="/database" element={<Database />} />
          <Route path="/security" element={<Security />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </Router>
  );
};

export default App;