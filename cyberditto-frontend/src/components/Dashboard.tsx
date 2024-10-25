import React from 'react';
import Sidebar from './Sidebar';
import './Dashboard.css';
import { 
    LineChart, Line, BarChart, Bar, PieChart, Pie, Cell,
    ResponsiveContainer, XAxis, YAxis, CartesianGrid, Tooltip,
    Legend, RadarChart, Radar, PolarGrid, PolarAngleAxis, PolarRadiusAxis
} from 'recharts';

const Dashboard: React.FC = () => {
    const benchmarkData = [
        { name: 'Account Policies', compliance: 85 },
        { name: 'Local Policies', compliance: 70 },
        { name: 'Event Log', compliance: 90 },
        { name: 'System Services', compliance: 75 },
        { name: 'Security Options', compliance: 88 },
    ];

    const trafficData = [
        { name: 'Jan', inbound: 4000, outbound: 2400 },
        { name: 'Feb', inbound: 3000, outbound: 1398 },
        { name: 'Mar', inbound: 2000, outbound: 9800 },
        { name: 'Apr', inbound: 2780, outbound: 3908 },
        { name: 'May', inbound: 1890, outbound: 4800 },
        { name: 'Jun', inbound: 2390, outbound: 3800 },
    ];

    const securityStatusData = [
        { name: 'Secure', value: 60, color: '#10b981' },
        { name: 'At Risk', value: 30, color: '#f59e0b' },
        { name: 'Critical', value: 10, color: '#ef4444' },
    ];

    const mitreData = [
        { subject: 'Initial Access', A: 85, fullMark: 100 },
        { subject: 'Execution', A: 65, fullMark: 100 },
        { subject: 'Persistence', A: 75, fullMark: 100 },
        { subject: 'Privilege Escalation', A: 80, fullMark: 100 },
        { subject: 'Defense Evasion', A: 70, fullMark: 100 },
        { subject: 'Lateral Movement', A: 85, fullMark: 100 },
    ];

    const emulationData = [
        { name: 'APT29', success: 12, failure: 28 },
        { name: 'APT3', success: 8, failure: 32 },
        { name: 'FIN6', success: 15, failure: 25 },
        { name: 'APT41', success: 10, failure: 30 },
    ];

    const endpointSecurityData = [
        { name: 'Firewall', score: 90 },
        { name: 'Antivirus', score: 85 },
        { name: 'Patch Level', score: 75 },
        { name: 'Config', score: 88 },
        { name: 'Access Control', score: 92 },
    ];

    const CustomTooltip = ({ active, payload, label }: any) => {
        if (active && payload && payload.length) {
            return (
                <div style={{ 
                    backgroundColor: '#fff',
                    padding: '8px 12px',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
                    fontFamily: "'Poppins', sans-serif"
                }}>
                    <p style={{ margin: '0 0 4px 0', fontWeight: 500 }}>{`${label}`}</p>
                    {payload.map((pld: any, index: number) => (
                        <p key={index} style={{ 
                            margin: '4px 0', 
                            color: pld.color || pld.stroke || '#666',
                            fontSize: '0.875rem'
                        }}>
                            {`${pld.name}: ${pld.value}${pld.name === 'compliance' || pld.name === 'score' ? '%' : ''}`}
                        </p>
                    ))}
                </div>
            );
        }
        return null;
    };

    return (
        <div className="app-dashboard">
            <Sidebar />
            <div className="app-dashboard-content">
                <div className="app-dashboard-header">
                    <h1>Security Operations Dashboard</h1>
                    <p className="app-dashboard-subtitle">CyberDitto Digital Twin Analysis</p>
                </div>
                
                <div className="app-dashboard-grid">
                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">
                            CIS Benchmark Compliance
                        </div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <BarChart data={benchmarkData}>
                                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                    <XAxis dataKey="name" angle={-45} textAnchor="end" height={80} />
                                    <YAxis domain={[0, 100]} />
                                    <Tooltip content={<CustomTooltip />} />
                                    <Bar dataKey="compliance" fill="#1a73e8" radius={[4, 4, 0, 0]} />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">
                            MITRE ATT&CK Coverage
                        </div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <RadarChart cx="50%" cy="50%" outerRadius="80%" data={mitreData}>
                                    <PolarGrid />
                                    <PolarAngleAxis dataKey="subject" />
                                    <PolarRadiusAxis angle={30} domain={[0, 100]} />
                                    <Radar name="Coverage" dataKey="A" stroke="#1a73e8" 
                                        fill="#1a73e8" fillOpacity={0.6} />
                                </RadarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">
                            Security Posture
                        </div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <PieChart>
                                    <Pie 
                                        data={securityStatusData} 
                                        dataKey="value" 
                                        nameKey="name" 
                                        cx="50%" 
                                        cy="50%" 
                                        innerRadius={60}
                                        outerRadius={80} 
                                        label
                                    >
                                        {securityStatusData.map((entry, index) => (
                                            <Cell key={`cell-${index}`} fill={entry.color} />
                                        ))}
                                    </Pie>
                                    <Tooltip content={<CustomTooltip />} />
                                    <Legend />
                                </PieChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">
                            Adversary Emulation Results
                        </div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <BarChart data={emulationData}>
                                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                    <XAxis dataKey="name" />
                                    <YAxis />
                                    <Tooltip content={<CustomTooltip />} />
                                    <Legend />
                                    <Bar dataKey="success" stackId="a" fill="#ef4444" />
                                    <Bar dataKey="failure" stackId="a" fill="#10b981" />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">
                            Network Traffic Analysis
                        </div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <LineChart data={trafficData}>
                                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                    <XAxis dataKey="name" />
                                    <YAxis />
                                    <Tooltip content={<CustomTooltip />} />
                                    <Legend />
                                    <Line type="monotone" dataKey="inbound" stroke="#8884d8" />
                                    <Line type="monotone" dataKey="outbound" stroke="#82ca9d" />
                                </LineChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">
                            Endpoint Security Scores
                        </div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <BarChart data={endpointSecurityData} layout="vertical">
                                    <CartesianGrid strokeDasharray="3 3" horizontal={false} />
                                    <XAxis type="number" domain={[0, 100]} />
                                    <YAxis dataKey="name" type="category" width={100} />
                                    <Tooltip content={<CustomTooltip />} />
                                    <Bar dataKey="score" fill="#1a73e8" radius={[0, 4, 4, 0]} />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;