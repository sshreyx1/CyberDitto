import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './Benchmark-Run.css';
import { 
  CheckCircle2, 
  Server, 
  Shield, 
  AlertCircle,
  Play,
  Clock,
  CheckSquare
} from 'lucide-react';

interface Endpoint {
  id: string;
  name: string;
  os: string;
  ip: string;
  status: 'online' | 'offline';
}

interface BenchmarkProfile {
  id: string;
  name: string;
  description: string;
  controls: number;
  severity: 'high' | 'medium' | 'low';
  estimatedTime: string;
}

const BenchmarkRun: React.FC = () => {
  const [selectedEndpoints, setSelectedEndpoints] = useState<string[]>([]);
  const [selectedProfile, setSelectedProfile] = useState<string>('');
  const [isRunning, setIsRunning] = useState(false);
  const [progress, setProgress] = useState(0);

  const endpoints: Endpoint[] = [
    { id: '1', name: 'PROD-DC-01', os: 'Windows Server 2019', ip: '192.168.1.10', status: 'online' },
    { id: '2', name: 'PROD-DC-02', os: 'Windows Server 2019', ip: '192.168.1.11', status: 'online' },
    { id: '3', name: 'PROD-APP-01', os: 'Windows Server 2016', ip: '192.168.1.20', status: 'offline' },
    { id: '4', name: 'PROD-APP-02', os: 'Windows Server 2016', ip: '192.168.1.21', status: 'online' }
  ];

  const benchmarkProfiles: BenchmarkProfile[] = [
    {
      id: '1',
      name: 'CIS Windows Server 2019 Level 1',
      description: 'Basic security configuration for Windows Server 2019',
      controls: 180,
      severity: 'high',
      estimatedTime: '45-60 minutes'
    },
    {
      id: '2',
      name: 'CIS Windows Server 2019 Level 2',
      description: 'Advanced security configuration for Windows Server 2019',
      controls: 240,
      severity: 'high',
      estimatedTime: '60-90 minutes'
    },
    {
      id: '3',
      name: 'CIS Windows Server 2016 Level 1',
      description: 'Basic security configuration for Windows Server 2016',
      controls: 160,
      severity: 'medium',
      estimatedTime: '40-50 minutes'
    }
  ];

  const toggleEndpoint = (id: string) => {
    setSelectedEndpoints(prev => 
      prev.includes(id) 
        ? prev.filter(epId => epId !== id)
        : [...prev, id]
    );
  };

  const handleProfileSelect = (id: string) => {
    setSelectedProfile(id);
  };

  const startBenchmark = () => {
    setIsRunning(true);
    // Simulate benchmark progress
    let currentProgress = 0;
    const interval = setInterval(() => {
      currentProgress += 1;
      setProgress(currentProgress);
      if (currentProgress >= 100) {
        clearInterval(interval);
        setIsRunning(false);
      }
    }, 300);
  };

  return (
    <div className="benchmark-run">
      <Sidebar />
      <div className="benchmark-run-content">
        <div className="benchmark-run-header">
          <h1>Run CIS Benchmark</h1>
          <p>Select endpoints and benchmark profile to evaluate security compliance</p>
        </div>

        <div className="benchmark-grid">
          {/* Endpoint Selection */}
          <div className="benchmark-section">
            <h2><Server className="section-icon" /> Select Endpoints</h2>
            <div className="endpoints-list">
              {endpoints.map(endpoint => (
                <div 
                  key={endpoint.id} 
                  className={`endpoint-item ${endpoint.status === 'offline' ? 'offline' : ''}`}
                  onClick={() => endpoint.status === 'online' && toggleEndpoint(endpoint.id)}
                >
                  <div className="endpoint-info">
                    <CheckSquare 
                      className={`checkbox ${selectedEndpoints.includes(endpoint.id) ? 'checked' : ''}`}
                    />
                    <div>
                      <h3>{endpoint.name}</h3>
                      <p>{endpoint.os} • {endpoint.ip}</p>
                    </div>
                  </div>
                  <span className={`endpoint-status ${endpoint.status}`}>
                    {endpoint.status}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Benchmark Profile Selection */}
          <div className="benchmark-section">
            <h2><Shield className="section-icon" /> Select Benchmark Profile</h2>
            <div className="profile-list">
              {benchmarkProfiles.map(profile => (
                <div 
                  key={profile.id}
                  className={`profile-item ${selectedProfile === profile.id ? 'selected' : ''}`}
                  onClick={() => handleProfileSelect(profile.id)}
                >
                  <div className="profile-header">
                    <h3>{profile.name}</h3>
                    <span className={`severity ${profile.severity}`}>
                      {profile.severity}
                    </span>
                  </div>
                  <p>{profile.description}</p>
                  <div className="profile-meta">
                    <span><CheckCircle2 size={16} /> {profile.controls} controls</span>
                    <span><Clock size={16} /> {profile.estimatedTime}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Benchmark Progress */}
          <div className="benchmark-section">
            <h2><AlertCircle className="section-icon" /> Benchmark Status</h2>
            <div className="benchmark-status">
              {!isRunning && progress === 0 ? (
                <div className="status-message">
                  <AlertCircle size={40} />
                  <p>Select endpoints and profile to start the benchmark</p>
                </div>
              ) : (
                <>
                  <div className="progress-info">
                    <span>Running CIS Benchmark...</span>
                    <span>{progress}%</span>
                  </div>
                  <div className="progress-bar">
                    <div 
                      className="progress-fill"
                      style={{ width: `${progress}%` }}
                    ></div>
                  </div>
                  <div className="selected-summary">
                    <p>{selectedEndpoints.length} endpoints selected</p>
                    <p>{benchmarkProfiles.find(p => p.id === selectedProfile)?.name}</p>
                  </div>
                </>
              )}
            </div>
            <button 
              className="start-benchmark"
              disabled={selectedEndpoints.length === 0 || !selectedProfile || isRunning}
              onClick={startBenchmark}
            >
              <Play size={18} />
              {isRunning ? 'Running Benchmark...' : 'Start Benchmark'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default BenchmarkRun;