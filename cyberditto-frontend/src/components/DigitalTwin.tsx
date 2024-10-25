import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './DigitalTwin.css';
import { FaServer, FaNetworkWired, FaWindowMaximize, FaCheckCircle } from 'react-icons/fa';

const DigitalTwin: React.FC = () => {
    const [deploymentStatus, setDeploymentStatus] = useState('idle'); // idle, deploying, completed, failed

    return (
        <div className="app-digitaltwin">
            <Sidebar />
            <div className="app-digitaltwin-content">
                <div className="app-digitaltwin-header">
                    <h1>Digital Twin Environment</h1>
                    <p className="app-digitaltwin-subtitle">Create and Manage Virtual System Replicas</p>
                </div>
                
                <div className="app-digitaltwin-grid">
                    {/* System Configuration Card */}
                    <div className="app-digitaltwin-card">
                        <div className="app-digitaltwin-card-header">
                            <FaServer className="card-icon" /> System Configuration
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <form className="config-form">
                                <div className="form-group">
                                    <label>Environment Name</label>
                                    <input type="text" placeholder="Production Environment" />
                                </div>
                                <div className="form-group">
                                    <label>Number of Endpoints</label>
                                    <input type="number" placeholder="Enter number of endpoints" />
                                </div>
                                <div className="form-group">
                                    <label>Operating System</label>
                                    <select>
                                        <option>Windows Server 2019</option>
                                        <option>Windows Server 2016</option>
                                        <option>Windows 10</option>
                                    </select>
                                </div>
                                <button className="deploy-button">Create Digital Twin</button>
                            </form>
                        </div>
                    </div>

                    {/* Network Configuration Card */}
                    <div className="app-digitaltwin-card">
                        <div className="app-digitaltwin-card-header">
                            <FaNetworkWired className="card-icon" /> Network Settings
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <form className="config-form">
                                <div className="form-group">
                                    <label>Network Type</label>
                                    <select>
                                        <option>Corporate LAN</option>
                                        <option>DMZ</option>
                                        <option>Custom</option>
                                    </select>
                                </div>
                                <div className="form-group">
                                    <label>IP Range</label>
                                    <input type="text" placeholder="192.168.1.0/24" />
                                </div>
                                <div className="form-group">
                                    <label>Network Topology</label>
                                    <select>
                                        <option>Star</option>
                                        <option>Mesh</option>
                                        <option>Ring</option>
                                    </select>
                                </div>
                            </form>
                        </div>
                    </div>

                    {/* Deployment Status Card */}
                    <div className="app-digitaltwin-card">
                        <div className="app-digitaltwin-card-header">
                            <FaCheckCircle className="card-icon" /> Deployment Status
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <div className="status-container">
                                <div className={`status-item ${deploymentStatus === 'completed' ? 'completed' : ''}`}>
                                    <span className="status-label">Environment Setup</span>
                                    <span className="status-indicator"></span>
                                </div>
                                <div className={`status-item ${deploymentStatus === 'completed' ? 'completed' : ''}`}>
                                    <span className="status-label">Network Configuration</span>
                                    <span className="status-indicator"></span>
                                </div>
                                <div className={`status-item ${deploymentStatus === 'completed' ? 'completed' : ''}`}>
                                    <span className="status-label">Endpoint Replication</span>
                                    <span className="status-indicator"></span>
                                </div>
                                <div className={`status-item ${deploymentStatus === 'completed' ? 'completed' : ''}`}>
                                    <span className="status-label">Validation</span>
                                    <span className="status-indicator"></span>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Active Twins Card */}
                    <div className="app-digitaltwin-card">
                        <div className="app-digitaltwin-card-header">
                            <FaWindowMaximize className="card-icon" /> Active Digital Twins
                        </div>
                        <div className="app-digitaltwin-card-content">
                            <div className="twins-list">
                                <div className="twin-item">
                                    <span className="twin-name">Production Environment</span>
                                    <span className="twin-status active">Running</span>
                                    <div className="twin-details">
                                        <small>5 Endpoints • Windows Server 2019</small>
                                    </div>
                                </div>
                                <div className="twin-item">
                                    <span className="twin-name">Test Environment</span>
                                    <span className="twin-status stopped">Stopped</span>
                                    <div className="twin-details">
                                        <small>3 Endpoints • Windows 10</small>
                                    </div>
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