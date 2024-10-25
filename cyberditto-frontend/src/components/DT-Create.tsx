import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './DT-Create.css';
import { Server, Network, CheckCircle, AlertCircle } from 'lucide-react';

const DTCreate: React.FC = () => {
  const [scanStep, setScanStep] = useState<number>(1);
  const [scanning, setScanning] = useState<boolean>(false);
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
              />
            </div>
            <div className="form-group">
              <label>Operating System</label>
              <select 
                name="operatingSystem"
                value={networkData.operatingSystem}
                onChange={handleInputChange}
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
              />
            </div>
            <div className="form-group">
              <label>Network Topology</label>
              <select
                name="topology"
                value={networkData.topology}
                onChange={handleInputChange}
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
              ) : (
                <div className="scan-complete">
                  <CheckCircle className="section-icon" />
                  <p>Scan complete! Review the results and create your digital twin</p>
                </div>
              )}
            </div>
            <button 
              className={`scan-button ${scanning ? 'scanning' : ''}`}
              onClick={handleScan}
              disabled={scanning}
            >
              {scanning ? 'Scanning...' : scanStep === 1 ? 'Start Scan' : 'Create Digital Twin'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DTCreate;