import React, { useState, useEffect } from 'react';
import Sidebar from './Sidebar';
import './Benchmark.css';
import { FaClipboardCheck, FaExclamationTriangle, FaChartBar, FaTools } from 'react-icons/fa';

const Benchmark: React.FC = () => {
    const [selectedBenchmark, setSelectedBenchmark] = useState('windows-server-2019');
    const [benchmarkData, setBenchmarkData] = useState<any>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchBenchmarkData = async () => {
            try {
                const response = await fetch('/api/dashboard-data');
                if (!response.ok) {
                    throw new Error('Failed to fetch benchmark data');
                }
                const data = await response.json();
                console.log('Raw Benchmark API Response:', data);
                setBenchmarkData(data);
            } catch (err: any) {
                console.error('Error fetching benchmark data:', err);
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchBenchmarkData();
    }, []);

    if (loading) return <div>Loading benchmark data...</div>;
    if (error) return <div>Error: {error}</div>;

    // Use aggregateMetrics from fetched benchmarkData
    const aggregateMetrics = benchmarkData?.aggregateMetrics || {
        averageCompliance: 0,
        totalSuccess: 0,
        totalFailure: 0,
    };

    return (
        <div className="app-benchmark">
            <Sidebar />
            <div className="app-benchmark-content">
                <div className="app-benchmark-header">
                    <h1>CIS Benchmark Assessment</h1>
                    <p className="app-benchmark-subtitle">Windows Security Configuration Analysis</p>
                </div>

                <div className="app-benchmark-grid">
                    {/* Benchmark Selection Card */}
                    <div className="app-benchmark-card">
                        <div className="app-benchmark-card-header">
                            <FaClipboardCheck className="card-icon" /> Run Benchmark
                        </div>
                        <div className="app-benchmark-card-content">
                            <form className="benchmark-form">
                                <div className="form-group">
                                    <label>Select Environment</label>
                                    <select>
                                        <option>Production Environment</option>
                                        <option>Test Environment</option>
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label>Benchmark Type</label>
                                    <select
                                        value={selectedBenchmark}
                                        onChange={(e) => setSelectedBenchmark(e.target.value)}
                                    >
                                        <option value="windows-enterprise-10">
                                            Windows Enterprise 10
                                        </option>
                                        <option value="windows-10">Windows 10</option>
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label>Assessment Level</label>
                                    <select>
                                        <option>Level 1 - Basic</option>
                                        <option>Level 2 - Advanced</option>
                                    </select>
                                </div>
                                <button className="run-benchmark-button">Start Assessment</button>
                            </form>
                        </div>
                    </div>

                    {/* Compliance Score Card */}
                    <div className="app-benchmark-card">
                        <div className="app-benchmark-card-header">
                            <FaChartBar className="card-icon" /> Compliance Overview
                        </div>
                        <div className="app-benchmark-card-content">
                            <div className="compliance-summary">
                                <div className="compliance-score">
                                    <span className="score">
                                        {aggregateMetrics.averageCompliance.toFixed(0)}%
                                    </span>
                                    <span className="score-label">Overall Compliance</span>
                                </div>
                                <div className="compliance-breakdown">
                                    <div className="breakdown-item">
                                        <span className="category">Account Policies</span>
                                        <div className="progress-bar">
                                            <div
                                                className="progress"
                                                style={{
                                                    width: `${aggregateMetrics.averageCompliance.toFixed(0)}%`,
                                                }}
                                            ></div>
                                        </div>
                                        <span className="percentage">
                                            {aggregateMetrics.averageCompliance.toFixed(0)}%
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Remediation Card */}
                    <div className="app-benchmark-card">
                        <div className="app-benchmark-card-header">
                            <FaTools className="card-icon" /> Remediation Actions
                        </div>
                        <div className="app-benchmark-card-content">
                            {benchmarkData?.remediationActions?.length > 0 ? (
                                <div className="remediation-list">
                                    {benchmarkData.remediationActions.map(
                                        (action: any, index: number) => (
                                            <div className="remediation-item" key={index}>
                                                <span className="remediation-title">
                                                    {action.title || 'Unknown Issue'}
                                                </span>
                                                <p className="remediation-steps">
                                                    {action.steps ||
                                                        'No steps provided for remediation.'}
                                                </p>
                                                <button className="remediate-button">
                                                    {action.buttonText || 'Apply Fix'}
                                                </button>
                                            </div>
                                        )
                                    )}
                                </div>
                            ) : (
                                <div className="no-remediation-actions">
                                    No remediation actions available at this time.
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Benchmark;
