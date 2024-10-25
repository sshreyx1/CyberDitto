import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './DT-View.css';
import { Server, Activity, PlayCircle, StopCircle, BarChart } from 'lucide-react';

interface DigitalTwin {
  id: string;
  name: string;
  status: 'running' | 'stopped' | 'error';
  endpoints: number;
  os: string;
  networkType: string;
  ipRange: string;
  createdAt: string;
  lastActive: string;
  resourceUsage: {
    cpu: number;
    memory: number;
    storage: number;
  };
}

const DTView: React.FC = () => {
  const [twins] = useState<DigitalTwin[]>([
    {
      id: '1',
      name: 'Production Environment',
      status: 'running',
      endpoints: 5,
      os: 'Windows Server 2019',
      networkType: 'Corporate LAN',
      ipRange: '192.168.1.0/24',
      createdAt: '2024-02-15',
      lastActive: '2024-02-20',
      resourceUsage: {
        cpu: 45,
        memory: 60,
        storage: 35
      }
    },
    {
      id: '2',
      name: 'Test Environment',
      status: 'stopped',
      endpoints: 3,
      os: 'Windows Server 2016',
      networkType: 'DMZ',
      ipRange: '192.168.2.0/24',
      createdAt: '2024-02-10',
      lastActive: '2024-02-18',
      resourceUsage: {
        cpu: 0,
        memory: 0,
        storage: 30
      }
    }
  ]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'status-running';
      case 'stopped':
        return 'status-stopped';
      case 'error':
        return 'status-error';
      default:
        return '';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  return (
    <div className="dt-view-wrapper">
      <Sidebar />
      <div className="dt-view-content">
        <div className="dt-view-header">
          <h1>Digital Twin Environments</h1>
          <p>Monitor and manage your virtual system replicas</p>
        </div>

        <div className="twins-grid">
          {twins.map(twin => (
            <div key={twin.id} className="twin-card">
              <div className="twin-card-header">
                <div className="twin-title">
                  <Server className="twin-icon" />
                  <h2>{twin.name}</h2>
                </div>
                <span className={`twin-status ${getStatusColor(twin.status)}`}>
                  {twin.status}
                </span>
              </div>

              <div className="twin-details">
                <div className="detail-row">
                  <span className="detail-label">Endpoints:</span>
                  <span className="detail-value">{twin.endpoints}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Operating System:</span>
                  <span className="detail-value">{twin.os}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Network Type:</span>
                  <span className="detail-value">{twin.networkType}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">IP Range:</span>
                  <span className="detail-value">{twin.ipRange}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Created:</span>
                  <span className="detail-value">{formatDate(twin.createdAt)}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Last Active:</span>
                  <span className="detail-value">{formatDate(twin.lastActive)}</span>
                </div>
              </div>

              <div className="resource-usage">
                <h3><Activity className="resource-icon" /> Resource Usage</h3>
                <div className="usage-bars">
                  <div className="usage-item">
                    <label>CPU</label>
                    <div className="progress-bar">
                      <div 
                        className="progress-fill"
                        style={{ width: `${twin.resourceUsage.cpu}%` }}
                      ></div>
                    </div>
                    <span>{twin.resourceUsage.cpu}%</span>
                  </div>
                  <div className="usage-item">
                    <label>Memory</label>
                    <div className="progress-bar">
                      <div 
                        className="progress-fill"
                        style={{ width: `${twin.resourceUsage.memory}%` }}
                      ></div>
                    </div>
                    <span>{twin.resourceUsage.memory}%</span>
                  </div>
                  <div className="usage-item">
                    <label>Storage</label>
                    <div className="progress-bar">
                      <div 
                        className="progress-fill"
                        style={{ width: `${twin.resourceUsage.storage}%` }}
                      ></div>
                    </div>
                    <span>{twin.resourceUsage.storage}%</span>
                  </div>
                </div>
              </div>

              <div className="twin-actions">
                {twin.status === 'running' ? (
                  <button className="action-button stop">
                    <StopCircle size={18} />
                    Stop Environment
                  </button>
                ) : (
                  <button className="action-button start">
                    <PlayCircle size={18} />
                    Start Environment
                  </button>
                )}
                <button className="action-button monitor">
                  <BarChart size={18} />
                  View Metrics
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default DTView;