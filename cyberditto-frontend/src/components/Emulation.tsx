import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './Emulation.css';
import { FaPlay, FaHistory, FaShieldAlt, FaChartLine, FaTools } from 'react-icons/fa';

const Emulation: React.FC = () => {
    const [scenario, setScenario] = useState('apt29');

    return (
        <div className="app-emulation">
            <Sidebar />
            <div className="app-emulation-content">
                <div className="app-emulation-header">
                    <h1>Adversary Emulation</h1>
                    <p className="app-emulation-subtitle">Caldera-Based Security Testing</p>
                </div>

                <div className="app-emulation-grid">
                    {/* Scenario Selection Card */}
                    <div className="app-emulation-card">
                        <div className="app-emulation-card-header">
                            <FaPlay className="card-icon" /> Attack Scenario
                        </div>
                        <div className="app-emulation-card-content">
                            <form className="scenario-form">
                                <div className="form-group">
                                    <label>Environment Target</label>
                                    <select>
                                        <option>Production Digital Twin</option>
                                        <option>Test Environment</option>
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label>Adversary Profile</label>
                                    <select value={scenario} onChange={(e) => setScenario(e.target.value)}>
                                        <option value="apt29">APT29 (Cozy Bear)</option>
                                        <option value="apt3">APT3 (Gothic Panda)</option>
                                        <option value="fin6">FIN6</option>
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label>Attack Intensity</label>
                                    <select>
                                        <option>Moderate</option>
                                        <option>Aggressive</option>
                                        <option>Stealth</option>
                                    </select>
                                </div>
                                <button className="run-scenario-button">Launch Attack Simulation</button>
                            </form>
                        </div>
                    </div>

                    {/* Active Operations Card */}
                    <div className="app-emulation-card">
                        <div className="app-emulation-card-header">
                            <FaHistory className="card-icon" /> Active Operations
                        </div>
                        <div className="app-emulation-card-content">
                            <div className="operations-list">
                                <div className="operation-item active">
                                    <span className="operation-name">APT29 Simulation</span>
                                    <span className="operation-status">In Progress</span>
                                    <div className="operation-details">
                                        <small>Started: 10 minutes ago • Progress: 45%</small>
                                    </div>
                                </div>
                                <div className="operation-item completed">
                                    <span className="operation-name">FIN6 Assessment</span>
                                    <span className="operation-status">Completed</span>
                                    <div className="operation-details">
                                        <small>Duration: 45 minutes • Success Rate: 65%</small>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Defense Analysis Card */}
                    <div className="app-emulation-card">
                        <div className="app-emulation-card-header">
                            <FaShieldAlt className="card-icon" /> Defense Effectiveness
                        </div>
                        <div className="app-emulation-card-content">
                            <div className="defense-metrics">
                                <div className="metric-group">
                                    <h3>Detection Rate</h3>
                                    <div className="circular-progress">
                                        <span className="percentage">75%</span>
                                    </div>
                                    <p>Attack techniques detected</p>
                                </div>
                                <div className="defense-breakdown">
                                    <div className="defense-item">
                                        <span>Initial Access Prevention</span>
                                        <div className="progress-bar">
                                            <div className="progress" style={{ width: '80%' }}></div>
                                        </div>
                                    </div>
                                    <div className="defense-item">
                                        <span>Lateral Movement Detection</span>
                                        <div className="progress-bar">
                                            <div className="progress" style={{ width: '65%' }}></div>
                                        </div>
                                    </div>
                                    <div className="defense-item">
                                        <span>Data Exfiltration Prevention</span>
                                        <div className="progress-bar">
                                            <div className="progress" style={{ width: '90%' }}></div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Attack Metrics Card */}
                    <div className="app-emulation-card">
                        <div className="app-emulation-card-header">
                            <FaChartLine className="card-icon" /> Attack Success Metrics
                        </div>
                        <div className="app-emulation-card-content">
                            <div className="attack-metrics">
                                <div className="metric-item">
                                    <h4>Successful Techniques</h4>
                                    <div className="technique-list">
                                        <div className="technique-item success">
                                            <span className="technique-name">T1078 - Valid Accounts</span>
                                            <span className="technique-status">Successful</span>
                                        </div>
                                        <div className="technique-item failed">
                                            <span className="technique-name">T1110 - Brute Force</span>
                                            <span className="technique-status">Blocked</span>
                                        </div>
                                        <div className="technique-item success">
                                            <span className="technique-name">T1003 - OS Credential Dumping</span>
                                            <span className="technique-status">Successful</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="attack-summary">
                                    <div className="summary-item">
                                        <span className="label">Total Techniques</span>
                                        <span className="value">15</span>
                                    </div>
                                    <div className="summary-item">
                                        <span className="label">Successful</span>
                                        <span className="value success">8</span>
                                    </div>
                                    <div className="summary-item">
                                        <span className="label">Blocked</span>
                                        <span className="value blocked">7</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* MITRE Coverage Card */}
                    <div className="app-emulation-card">
                        <div className="app-emulation-card-header">
                            <FaShieldAlt className="card-icon" /> MITRE ATT&CK Coverage
                        </div>
                        <div className="app-emulation-card-content">
                            <div className="mitre-coverage">
                                <div className="tactics-list">
                                    <div className="tactic-item">
                                        <span className="tactic-name">Initial Access</span>
                                        <div className="techniques-progress">
                                            <div className="progress" style={{ width: '75%' }}></div>
                                            <span className="count">6/8</span>
                                        </div>
                                    </div>
                                    <div className="tactic-item">
                                        <span className="tactic-name">Execution</span>
                                        <div className="techniques-progress">
                                            <div className="progress" style={{ width: '60%' }}></div>
                                            <span className="count">3/5</span>
                                        </div>
                                    </div>
                                    <div className="tactic-item">
                                        <span className="tactic-name">Persistence</span>
                                        <div className="techniques-progress">
                                            <div className="progress" style={{ width: '80%' }}></div>
                                            <span className="count">4/5</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="coverage-summary">
                                    <div className="total-coverage">
                                        <span className="label">Total Coverage</span>
                                        <span className="value">72%</span>
                                    </div>
                                    <button className="view-details-button">View Full Matrix</button>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Remediation Actions Card */}
                    <div className="app-emulation-card">
                        <div className="app-emulation-card-header">
                            <FaTools className="card-icon" /> Remediation Actions
                        </div>
                        <div className="app-emulation-card-content">
                            <div className="remediation-list">
                                <div className="remediation-item">
                                    <span className="remediation-title">Patch Vulnerable Service</span>
                                    <p className="remediation-steps">Update exposed RDP service to latest security patch</p>
                                    <button className="remediate-button">Apply Fix</button>
                                </div>
                                <div className="remediation-item">
                                    <span className="remediation-title">Strengthen Access Controls</span>
                                    <p className="remediation-steps">Implement Network Level Authentication for RDP</p>
                                    <button className="remediate-button">Apply Fix</button>
                                </div>
                                <div className="remediation-item">
                                    <span className="remediation-title">Configure EDR Alerts</span>
                                    <p className="remediation-steps">Update EDR rules to detect credential dumping attempts</p>
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

export default Emulation;