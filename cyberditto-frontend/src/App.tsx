import React from 'react';
import { BrowserRouter as Router, Route, Routes, Navigate, RouteProps } from 'react-router-dom';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import DTCreate from './components/DT-Create';
import DTView from './components/DT-View';
import BenchmarkRun from './components/Benchmark-Run';
import BenchmarkView from './components/Benchmark-View';
import EmulationRun from './components/Emulation-Run';
import EmulationView from './components/Emulation-View';
import DigitalTwin from './components/DigitalTwin';
import Benchmark from './components/Benchmark';
import Emulation from './components/Emulation';

interface ProtectedRouteProps extends Omit<RouteProps, 'element'> {
  element: React.ReactElement;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ element }) => {
  const isAuthenticated = localStorage.getItem('authToken');
  
  if (!isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return element;
};

const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        {/* Public routes */}
        <Route path="/" element={<Login />} />
        <Route path="/login" element={<Login />} />

        {/* Protected routes */}
        <Route path="/dashboard" element={<ProtectedRoute element={<Dashboard />} />} />
        
        {/* Digital Twin routes */}
        <Route path="/digitaltwin" element={<ProtectedRoute element={<DigitalTwin />} />} />
        <Route path="/digitaltwin/create" element={<ProtectedRoute element={<DTCreate />} />} />
        <Route path="/digitaltwin/view" element={<ProtectedRoute element={<DTView />} />} />
        
        {/* Benchmark routes */}
        <Route path="/benchmark" element={<ProtectedRoute element={<Benchmark />} />} />
        <Route path="/benchmark/run" element={<ProtectedRoute element={<BenchmarkRun />} />} />
        <Route path="/benchmark/view" element={<ProtectedRoute element={<BenchmarkView />} />} />
        
        {/* Emulation routes */}
        <Route path="/emulation" element={<ProtectedRoute element={<Emulation />} />} />
        <Route path="/emulation/run" element={<ProtectedRoute element={<EmulationRun />} />} />
        <Route path="/emulation/view" element={<ProtectedRoute element={<EmulationView />} />} />

        {/* Catch all route */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Router>
  );
};

export default App;
