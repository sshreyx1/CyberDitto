import React from 'react';
import Sidebar from './Sidebar';
import './Dashboard.css';
import { 
    LineChart, Line, BarChart, Bar, PieChart, Pie, 
    ResponsiveContainer, XAxis, YAxis, CartesianGrid, Tooltip 
} from 'recharts';

const Dashboard: React.FC = () => {
    const lineData = [
        { name: 'Page A', uv: 4000, pv: 2400 },
        { name: 'Page B', uv: 3000, pv: 1398 },
        { name: 'Page C', uv: 2000, pv: 9800 },
        { name: 'Page D', uv: 2780, pv: 3908 },
        { name: 'Page E', uv: 1890, pv: 4800 },
        { name: 'Page F', uv: 2390, pv: 3800 },
        { name: 'Page G', uv: 3490, pv: 4300 },
    ];

    const barData = [
        { name: 'CIS Benchmark A', compliance: 85 },
        { name: 'CIS Benchmark B', compliance: 70 },
        { name: 'CIS Benchmark C', compliance: 90 },
    ];

    const pieData = [
        { name: 'Safe', value: 60 },
        { name: 'Minor Issues', value: 30 },
        { name: 'Critical Issues', value: 10 },
    ];

    return (
        <div className="dashboard-container">
            <Sidebar />
            <div className="dashboard-content">
                <h1>Dashboard</h1>
                
                <div className="dashboard-grid">
                    {/* CIS Benchmark Compliance Card */}
                    <div className="card">
                        <div className="card-header">CIS Benchmark Compliance</div>
                        <div className="card-content">
                            <ResponsiveContainer width="100%" height={250}>
                                <BarChart data={barData}>
                                    <XAxis dataKey="name" />
                                    <YAxis />
                                    <Tooltip />
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <Bar dataKey="compliance" fill="#1a73e8" />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    {/* Traffic Analysis Card */}
                    <div className="card">
                        <div className="card-header">Traffic Analysis</div>
                        <div className="card-content">
                            <ResponsiveContainer width="100%" height={250}>
                                <LineChart data={lineData}>
                                    <XAxis dataKey="name" />
                                    <YAxis />
                                    <Tooltip />
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <Line type="monotone" dataKey="pv" stroke="#8884d8" />
                                    <Line type="monotone" dataKey="uv" stroke="#82ca9d" />
                                </LineChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    {/* Security Status Card */}
                    <div className="card">
                        <div className="card-header">Security Status</div>
                        <div className="card-content">
                            <ResponsiveContainer width="100%" height={250}>
                                <PieChart>
                                    <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={80} fill="#1a73e8" label />
                                    <Tooltip />
                                </PieChart>
                            </ResponsiveContainer>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;