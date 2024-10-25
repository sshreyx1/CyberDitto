import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './Benchmark.css';
import { FaClipboardCheck, FaExclamationTriangle, FaChartBar, FaTools } from 'react-icons/fa';

const Benchmark: React.FC = () => {
    const [selectedBenchmark, setSelectedBenchmark] = useState('windows-server-2019');

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
                                    <select value={selectedBenchmark} onChange={(e) => setSelectedBenchmark(e.target.value)}>
                                        <option value="windows-server-2019">Windows Server 2019</option>
                                        <option value="windows-server-2016">Windows Server 2016</option>
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

                    {/* Compliance Issues Card */}
                    <div className="app-benchmark-card">
                        <div className="app-benchmark-card-header">
                            <FaExclamationTriangle className="card-icon" /> Critical Findings
                        </div>
                        <div className="app-benchmark-card-content">
                            <div className="findings-list">
                                <div className="finding-item critical">
                                    <span className="finding-title">Password Policy Non-Compliant</span>
                                    <p className="finding-description">Minimum password length below recommended value</p>
                                    <span className="finding-severity">High Risk</span>
                                </div>
                                <div className="finding-item warning">
                                    <span className="finding-title">Audit Policy Incomplete</span>
                                    <p className="finding-description">Security event logging not fully configured</p>
                                    <span className="finding-severity">Medium Risk</span>
                                </div>
                            </div>
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
                                    <span className="score">78%</span>
                                    <span className="score-label">Overall Compliance</span>
                                </div>
                                <div className="compliance-breakdown">
                                    <div className="breakdown-item">
                                        <span className="category">Account Policies</span>
                                        <div className="progress-bar">
                                            <div className="progress" style={{width: '85%'}}></div>
                                        </div>
                                        <span className="percentage">85%</span>
                                    </div>
                                    <div className="breakdown-item">
                                        <span className="category">Security Options</span>
                                        <div className="progress-bar">
                                            <div className="progress" style={{width: '70%'}}></div>
                                        </div>
                                        <span className="percentage">70%</span>
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
                            <div className="remediation-list">
                                <div className="remediation-item">
                                    <span className="remediation-title">Update Password Policy</span>
                                    <p className="remediation-steps">Run PowerShell script to update group policy settings</p>
                                    <button className="remediate-button">Apply Fix</button>
                                </div>
                                <div className="remediation-item">
                                    <span className="remediation-title">Configure Audit Policy</span>
                                    <p className="remediation-steps">Enable security event auditing via group policy</p>
                                    <button className="remediate-button">Apply Fix</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Benchmark;