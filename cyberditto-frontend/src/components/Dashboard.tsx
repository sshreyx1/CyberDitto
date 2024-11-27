import React, { useState, useEffect } from 'react';
import Sidebar from './Sidebar';
import './Dashboard.css';
import {
    BarChart, Bar, PieChart, Pie, Cell,
    ResponsiveContainer, XAxis, YAxis, CartesianGrid, Tooltip,
    Legend, RadarChart, Radar, PolarGrid, PolarAngleAxis, PolarRadiusAxis
} from 'recharts';

const Dashboard: React.FC = () => {
    const [dashboardData, setDashboardData] = useState<any>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchDashboardData = async () => {
            try {
                const response = await fetch('/api/dashboard-data');
                if (!response.ok) {
                    throw new Error('Failed to fetch dashboard data');
                }
                const data = await response.json();
                console.log("Raw API Response:", data); // Log the raw response
                setDashboardData(data);
            } catch (err: any) {
                console.error("Error fetching dashboard data:", err); // Debugging log
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchDashboardData();
    }, []);

    const handleGeneratePDF = () => {
        window.open('http://localhost:8080/api/generate-pdf', '_blank');
    };

    if (loading) return <p>Loading...</p>;
    if (error) return <p>Error: {error}</p>;

    const {
        benchmarkStatuses = {}, // Correct spelling matches the API
        aggregateMetrics = { averageCompliance: 0, totalSuccess: 0, totalFailure: 0 },
    } = dashboardData || {};

    const cisBenchmarkData = benchmarkStatuses["CIS Benchmark"] || [];
    const mitreAttackData = benchmarkStatuses["MITRE ATT&CK"] || [];

    console.log("CIS Benchmark Data:", cisBenchmarkData); // Should now show the correct array
    console.log("MITRE ATT&CK Data:", mitreAttackData);   // Should now show the correct array

    return (
        <div className="app-dashboard">
            <Sidebar />
            <div className="app-dashboard-content">
                <div className="app-dashboard-header">
                    <h1>Security Operations Dashboard</h1>
                    <p className="app-dashboard-subtitle">Adversary Emulation Analysis</p>
                    <button
                        onClick={handleGeneratePDF}
                        className="app-dashboard-pdf-button"
                    >
                        Generate PDF Report
                    </button>
                </div>

                {/* Dashboard Grid */}
                <div className="app-dashboard-grid">
                    {/* Pass/Fail Chart */}
                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">CIS Benchmark Compliance (Pass/Fail)</div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <BarChart
                                    data={[
                                        {
                                            category: "Pass",
                                            count: cisBenchmarkData.filter((item) => item.compliance === 100).length,
                                        },
                                        {
                                            category: "Fail",
                                            count: cisBenchmarkData.filter((item) => item.compliance < 100).length,
                                        },
                                    ]}
                                    margin={{ top: 20, right: 30, left: 20, bottom: 50 }}
                                >
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <XAxis dataKey="category" />
                                    <YAxis />
                                    <Tooltip />
                                    <Bar dataKey="count" fill="#1a73e8" />
                                </BarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    {/* MITRE ATT&CK Coverage */}
                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">Atomic Coverage</div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <RadarChart cx="50%" cy="50%" outerRadius="80%" data={mitreAttackData}>
                                    <PolarGrid />
                                    <PolarAngleAxis dataKey="name" />
                                    <PolarRadiusAxis domain={[0, 100]} />
                                    <Radar name="Coverage" dataKey="compliance" stroke="#1a73e8" fill="#1a73e8" fillOpacity={0.6} />
                                    <Legend />
                                </RadarChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    {/* Security Posture */}
                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">Security Posture</div>
                        <div className="app-dashboard-card-content">
                            <ResponsiveContainer width="100%" height={300}>
                                <PieChart>
                                    <Pie
                                        dataKey="value"
                                        data={[
                                            { name: "Secure", value: 60, color: "#4caf50" },
                                            { name: "At Risk", value: 30, color: "#ffc107" },
                                            { name: "Critical", value: 10, color: "#f44336" },
                                        ]}
                                        cx="50%"
                                        cy="50%"
                                        outerRadius={100}
                                    >
                                        {["#4caf50", "#ffc107", "#f44336"].map((color, index) => (
                                            <Cell key={index} fill={color} />
                                        ))}
                                    </Pie>
                                    <Tooltip />
                                    <Legend />
                                </PieChart>
                            </ResponsiveContainer>
                        </div>
                    </div>

                    {/* Summary */}
                    <div className="app-dashboard-card">
                        <div className="app-dashboard-card-header">Summary</div>
                        <div className="app-dashboard-card-content">
                            <p>Average Compliance: {aggregateMetrics.averageCompliance.toFixed(2)}%</p>
                            <p>Total Success: {aggregateMetrics.totalSuccess}</p>
                            <p>Total Failure: {aggregateMetrics.totalFailure}</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
