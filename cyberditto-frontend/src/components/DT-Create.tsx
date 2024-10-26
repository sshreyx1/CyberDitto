import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './DT-Create.css';
import { Server, Network, CheckCircle, AlertCircle } from 'lucide-react';

const DTCreate: React.FC = () => {
  const [scanStep, setScanStep] = useState<number>(1);
  const [scanning, setScanning] = useState<boolean>(false);
  const [deploying, setDeploying] = useState<boolean>(false);
  const [deploymentStep, setDeploymentStep] = useState<number>(0);
  const [deploymentComplete, setDeploymentComplete] = useState<boolean>(false);
  const [networkData, setNetworkData] = useState({
    environmentName: '',
    endpointCount: '',
    operatingSystem: 'Windows Server 2019',
    networkType: 'Corporate LAN',
    ipRange: '',
    topology: 'Star'
  });

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setNetworkData({
      ...networkData,
      [e.target.name]: e.target.value
    });
  };

  const handleScan = async () => {
    setScanning(true);
    // Simulate scanning process
    setTimeout(() => {
      setScanStep(2);
      setScanning(false);
    }, 3000);
  };

  const handleDeploy = async () => {
    setDeploying(true);
    
    // Simulate deployment steps
    const steps = [
      'Environment Setup',
      'Network Configuration',
      'Endpoint Replication',
      'Validation'
    ];

    for (let i = 0; i < steps.length; i++) {
      await new Promise(resolve => setTimeout(resolve, 2000));
      setDeploymentStep(i + 1);
    }

    setTimeout(() => {
      setDeploying(false);
      setDeploymentComplete(true);
    }, 1000);
  };

  const handleReset = () => {
    setScanStep(1);
    setDeploymentStep(0);
    setDeploymentComplete(false);
    setNetworkData({
      environmentName: '',
      endpointCount: '',
      operatingSystem: 'Windows Server 2019',
      networkType: 'Corporate LAN',
      ipRange: '',
      topology: 'Star'
    });
  };

  return (
    <div className="dt-create-wrapper">
      <Sidebar />
      <div className="dt-create-content">
        <div className="dt-create-header">
          <h1>Create Digital Twin</h1>
          <p>Scan and create a virtual replica of your system architecture</p>
        </div>

        <div className="dt-create-grid">
          <div className="dt-create-section">
            <h2><Server className="section-icon" /> System Configuration</h2>
            <div className="form-group">
              <label>Environment Name</label>
              <input
                type="text"
                name="environmentName"
                value={networkData.environmentName}
                onChange={handleInputChange}
                placeholder="e.g., Production Environment"
                disabled={deploymentComplete}
              />
            </div>
            <div className="form-group">
              <label>Number of Endpoints</label>
              <input
                type="number"
                name="endpointCount"
                value={networkData.endpointCount}
                onChange={handleInputChange}
                placeholder="Enter the number of endpoints"
                disabled={deploymentComplete}
              />
            </div>
            <div className="form-group">
              <label>Operating System</label>
              <select 
                name="operatingSystem"
                value={networkData.operatingSystem}
                onChange={handleInputChange}
                disabled={deploymentComplete}
              >
                <option>Windows Server 2019</option>
                <option>Windows Server 2016</option>
                <option>Windows 10</option>
              </select>
            </div>
          </div>

          <div className="dt-create-section">
            <h2><Network className="section-icon" /> Network Settings</h2>
            <div className="form-group">
              <label>Network Type</label>
              <select
                name="networkType"
                value={networkData.networkType}
                onChange={handleInputChange}
                disabled={deploymentComplete}
              >
                <option>Corporate LAN</option>
                <option>DMZ</option>
                <option>Custom</option>
              </select>
            </div>
            <div className="form-group">
              <label>IP Range</label>
              <input
                type="text"
                name="ipRange"
                value={networkData.ipRange}
                onChange={handleInputChange}
                placeholder="e.g., 192.168.1.0/24"
                disabled={deploymentComplete}
              />
            </div>
            <div className="form-group">
              <label>Network Topology</label>
              <select
                name="topology"
                value={networkData.topology}
                onChange={handleInputChange}
                disabled={deploymentComplete}
              >
                <option>Star</option>
                <option>Mesh</option>
                <option>Ring</option>
              </select>
            </div>
          </div>

          <div className="dt-create-section">
            <h2><CheckCircle className="section-icon" /> Scanning Progress</h2>
            <div className="scanning-status">
              {scanning ? (
                <div className="scanning-progress">
                  <div className="spinner"></div>
                  <p>Scanning network and endpoints...</p>
                </div>
              ) : scanStep === 1 ? (
                <div className="scan-instructions">
                  <AlertCircle className="section-icon" />
                  <p>Configure the settings above and click Start Scan to begin the discovery process</p>
                </div>
              ) : deploymentComplete ? (
                <div className="scan-complete">
                  <CheckCircle className="section-icon" />
                  <p>Digital Twin deployed successfully!</p>
                </div>
              ) : (
                <div className="scan-complete">
                  <CheckCircle className="section-icon" />
                  <p>Scan complete! Review the results and create your digital twin</p>
                </div>
              )}
            </div>
            {deploymentComplete ? (
              <button 
                className="scan-button reset"
                onClick={handleReset}
              >
                Create New Digital Twin
              </button>
            ) : (
              <button 
                className={`scan-button ${scanning || deploying ? 'scanning' : ''}`}
                onClick={scanStep === 1 ? handleScan : handleDeploy}
                disabled={scanning || deploying}
              >
                {scanning ? 'Scanning...' : 
                 deploying ? 'Deploying...' : 
                 scanStep === 1 ? 'Start Scan' : 'Deploy Digital Twin'}
              </button>
            )}
          </div>

          {/* Deployment Status Section - Now stays visible after completion */}
          {(scanStep === 2 || deploying || deploymentComplete) && (
            <div className="dt-create-section">
              <h2><CheckCircle className="section-icon" /> Deployment Status</h2>
              <div className="status-container">
                <div className={`status-item ${deploymentStep >= 1 ? 'completed' : ''}`}>
                  <span className="status-label">Environment Setup</span>
                  <span className="status-indicator"></span>
                </div>
                <div className={`status-item ${deploymentStep >= 2 ? 'completed' : ''}`}>
                  <span className="status-label">Network Configuration</span>
                  <span className="status-indicator"></span>
                </div>
                <div className={`status-item ${deploymentStep >= 3 ? 'completed' : ''}`}>
                  <span className="status-label">Endpoint Replication</span>
                  <span className="status-indicator"></span>
                </div>
                <div className={`status-item ${deploymentStep >= 4 ? 'completed' : ''}`}>
                  <span className="status-label">Validation</span>
                  <span className="status-indicator"></span>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default DTCreate;