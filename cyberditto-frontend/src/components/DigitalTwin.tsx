import React from 'react';
import Sidebar from './Sidebar';
import './DigitalTwin.css';
import { LineChart, ResponsiveContainer, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';
import { Network, Activity, Cpu, Server, AlertCircle, CheckCircle } from 'lucide-react';

interface PerformanceData {
  name: string;
  cpu: number;
  memory: number;
  network: number;
}

const DigitalTwin: React.FC = () => {
    const performanceData: PerformanceData[] = [
        { name: '00:00', cpu: 45, memory: 60, network: 30 },
        { name: '04:00', cpu: 55, memory: 65, network: 45 },
        { name: '08:00', cpu: 75, memory: 80, network: 65 },
        { name: '12:00', cpu: 85, memory: 85, network: 70 },
        { name: '16:00', cpu: 65, memory: 75, network: 55 },
        { name: '20:00', cpu: 50, memory: 65, network: 40 },
    ];

    return (
        <div className="app-digitaltwin">
            <Sidebar />
            <div className="app-digitaltwin-content">
                <div className="app-digitaltwin-header">
                    <h1>Digital Twin Environment</h1>
                    <p className="app-digitaltwin-subtitle">Monitor and Manage Your Virtual System Replicas</p>
                </div>
                
                <div className="app-digitaltwin-grid">
                    {/* Performance Metrics Card */}
                    <div className="app-digitaltwin-card performance-card">
                        <div className="app-digitaltwin-card-header">
                            <Activity className="card-icon" /> Performance Metrics
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <div className="performance-chart">
                                <ResponsiveContainer width="100%" height={300}>
                                    <LineChart data={performanceData} margin={{ top: 20, right: 30, left: 20, bottom: 20 }}>
                                        <CartesianGrid strokeDasharray="3 3" />
                                        <XAxis dataKey="name" />
                                        <YAxis />
                                        <Tooltip 
                                            contentStyle={{ 
                                                backgroundColor: '#fff',
                                                border: '1px solid #e2e8f0',
                                                borderRadius: '6px',
                                                padding: '8px'
                                            }} 
                                        />
                                        <Legend verticalAlign="top" height={36} />
                                        <Line 
                                            type="monotone" 
                                            dataKey="cpu" 
                                            stroke="#1a73e8" 
                                            name="CPU Usage"
                                            strokeWidth={2}
                                            dot={{ strokeWidth: 2 }}
                                            activeDot={{ r: 6 }}
                                        />
                                        <Line 
                                            type="monotone" 
                                            dataKey="memory" 
                                            stroke="#10b981" 
                                            name="Memory Usage"
                                            strokeWidth={2}
                                            dot={{ strokeWidth: 2 }}
                                            activeDot={{ r: 6 }}
                                        />
                                        <Line 
                                            type="monotone" 
                                            dataKey="network" 
                                            stroke="#6366f1" 
                                            name="Network Usage"
                                            strokeWidth={2}
                                            dot={{ strokeWidth: 2 }}
                                            activeDot={{ r: 6 }}
                                        />
                                    </LineChart>
                                </ResponsiveContainer>
                            </div>
                        </div>
                    </div>

                    {/* Active Digital Twins Card */}
                    <div className="app-digitaltwin-card">
                        <div className="app-digitaltwin-card-header">
                            <Server className="card-icon" /> Active Digital Twins
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <div className="twins-list">
                                <div className="environment-card">
                                    <div className="env-header">
                                        <div className="env-main">
                                            <h3>Production Environment</h3>
                                            <span className="env-status running">Running</span>
                                        </div>
                                        <div className="env-metrics">
                                            <div className="metric">
                                                <Cpu size={16} />
                                                <span>CPU:</span>
                                                <span className="value">75%</span>
                                            </div>
                                            <div className="metric">
                                                <Activity size={16} />
                                                <span>Memory:</span>
                                                <span className="value">80%</span>
                                            </div>
                                            <div className="metric">
                                                <Network size={16} />
                                                <span>Network:</span>
                                                <span className="value">65%</span>
                                            </div>
                                        </div>
                                    </div>
                                    <div className="env-info">
                                        <div className="info-item">
                                            <span className="info-label">5 Endpoints</span>
                                        </div>
                                        <div className="info-item">
                                            <span className="info-label">Windows Server 2019</span>
                                        </div>
                                        <div className="info-item">
                                            <span className="info-label">Last Updated: 2 mins ago</span>
                                        </div>
                                    </div>
                                </div>

                                <div className="environment-card">
                                    <div className="env-header">
                                        <div className="env-main">
                                            <h3>Test Environment</h3>
                                            <span className="env-status stopped">Stopped</span>
                                        </div>
                                        <div className="env-metrics">
                                            <div className="metric inactive">
                                                <Cpu size={16} />
                                                <span>CPU:</span>
                                                <span className="value">0%</span>
                                            </div>
                                            <div className="metric inactive">
                                                <Activity size={16} />
                                                <span>Memory:</span>
                                                <span className="value">0%</span>
                                            </div>
                                            <div className="metric inactive">
                                                <Network size={16} />
                                                <span>Network:</span>
                                                <span className="value">0%</span>
                                            </div>
                                        </div>
                                    </div>
                                    <div className="env-info">
                                        <div className="info-item">
                                            <span className="info-label">3 Endpoints</span>
                                        </div>
                                        <div className="info-item">
                                            <span className="info-label">Windows 10</span>
                                        </div>
                                        <div className="info-item">
                                            <span className="info-label">Stopped: 1 hour ago</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Network Topology Card */}
                    <div className="app-digitaltwin-card">
                        <div className="app-digitaltwin-card-header">
                            <Network className="card-icon" /> Network Topology
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <div className="topology-stats">
                                <div className="topology-stat-item">
                                    <div className="stat-value">8</div>
                                    <div className="stat-label">Total Endpoints</div>
                                </div>
                                <div className="topology-stat-item">
                                    <div className="stat-value">2</div>
                                    <div className="stat-label">Active Networks</div>
                                </div>
                                <div className="topology-stat-item">
                                    <div className="stat-value">99.9%</div>
                                    <div className="stat-label">Uptime</div>
                                </div>
                            </div>
                            <div className="topology-status">
                                <div className="topology-header">Network Health Status</div>
                                <div className="health-items">
                                    <div className="health-item healthy">
                                        <CheckCircle size={16} />
                                        <span>Production Network - Healthy</span>
                                    </div>
                                    <div className="health-item warning">
                                        <AlertCircle size={16} />
                                        <span>Test Network - Inactive</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* System Alerts Card */}
                    <div className="app-digitaltwin-card">
                        <div className="app-digitaltwin-card-header">
                            <AlertCircle className="card-icon" /> System Alerts
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <div className="alerts-list">
                                <div className="alert-item high">
                                    <div className="alert-header">
                                        <AlertCircle size={16} />
                                        <span>High CPU Usage Detected</span>
                                    </div>
                                    <p>Production Environment - Server 2 exceeded 85% CPU utilization</p>
                                    <small>2 minutes ago</small>
                                </div>
                                <div className="alert-item medium">
                                    <div className="alert-header">
                                        <AlertCircle size={16} />
                                        <span>Memory Usage Warning</span>
                                    </div>
                                    <p>Production Environment - Memory usage trending upward</p>
                                    <small>15 minutes ago</small>
                                </div>
                                <div className="alert-item low">
                                    <div className="alert-header">
                                        <AlertCircle size={16} />
                                        <span>Network Latency Detected</span>
                                    </div>
                                    <p>Test Environment - Minor network latency observed</p>
                                    <small>1 hour ago</small>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DigitalTwin;