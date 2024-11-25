import React from 'react';
import { 
    BrowserRouter as Router, 
    Route, 
    Routes, 
    Navigate
} from 'react-router-dom';
import { 
    UNSAFE_NavigationContext,
    UNSAFE_LocationContext,
    UNSAFE_RouteContext 
} from 'react-router-dom';
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
import { authService } from './services/AuthService';

interface ProtectedRouteProps {
    children: React.ReactNode;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
    const isAuthenticated = authService.isAuthenticated();
    React.useEffect(() => {
        if (!isAuthenticated) {
            authService.logout();
        }
    }, [isAuthenticated]);

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }
    return <>{children}</>;
};

const App: React.FC = () => {
    return (
        <Router future={{ v7_startTransition: true }}>
            <Routes>
                {/* Public routes */}
                <Route path="/" element={<Navigate to="/login" replace />} />
                <Route path="/login" element={<Login />} />

                {/* Protected routes */}
                <Route path="/dashboard" element={
                    <ProtectedRoute>
                        <Dashboard />
                    </ProtectedRoute>
                } />
                
                {/* Digital Twin routes */}
                <Route path="/digitaltwin" element={
                    <ProtectedRoute>
                        <DigitalTwin />
                    </ProtectedRoute>
                } />
                <Route path="/digitaltwin/create" element={
                    <ProtectedRoute>
                        <DTCreate />
                    </ProtectedRoute>
                } />
                <Route path="/digitaltwin/view" element={
                    <ProtectedRoute>
                        <DTView />
                    </ProtectedRoute>
                } />
                
                {/* Benchmark routes */}
                <Route path="/benchmark" element={
                    <ProtectedRoute>
                        <Benchmark />
                    </ProtectedRoute>
                } />
                <Route path="/benchmark/run" element={
                    <ProtectedRoute>
                        <BenchmarkRun />
                    </ProtectedRoute>
                } />
                <Route path="/benchmark/view" element={
                    <ProtectedRoute>
                        <BenchmarkView />
                    </ProtectedRoute>
                } />
                
                {/* Emulation routes */}
                <Route path="/emulation" element={
                    <ProtectedRoute>
                        <Emulation />
                    </ProtectedRoute>
                } />
                <Route path="/emulation/run" element={
                    <ProtectedRoute>
                        <EmulationRun />
                    </ProtectedRoute>
                } />
                <Route path="/emulation/view" element={
                    <ProtectedRoute>
                        <EmulationView />
                    </ProtectedRoute>
                } />

                {/* Catch all route */}
                <Route path="*" element={<Navigate to="/login" replace />} />
            </Routes>
        </Router>
    );
};

export default App;