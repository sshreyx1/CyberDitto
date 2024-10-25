import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './Emulation-View.css';
import { 
  Target, 
  Shield, 
  AlertTriangle, 
  CheckCircle2, 
  XCircle,
  ArrowRight,
  Clock,
  Download,
  ChevronDown,
  BarChart2,
  FileText,
  Activity
} from 'lucide-react';

interface EmulationResult {
  id: string;
  scenarioName: string;
  targetEnvironment: string;
  date: string;
  duration: string;
  status: 'completed' | 'failed' | 'aborted';
  successRate: number;
  techniques: {
    total: number;
    successful: number;
    failed: number;
    blocked: number;
  };
  findings: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  techniques_detail: {
    id: string;
    name: string;
    status: 'success' | 'failed' | 'blocked';
    description: string;
    duration: string;
  }[];
}

const EmulationView: React.FC = () => {
  const [expandedResult, setExpandedResult] = useState<string | null>(null);
  const [timeFilter, setTimeFilter] = useState('all');

  const results: EmulationResult[] = [
    {
      id: '1',
      scenarioName: 'Credential Access and Lateral Movement',
      targetEnvironment: 'Production Environment',
      date: '2024-02-20',
      duration: '45 minutes',
      status: 'completed',
      successRate: 75,
      techniques: {
        total: 12,
        successful: 9,
        failed: 2,
        blocked: 1
      },
      findings: {
        critical: 2,
        high: 3,
        medium: 4,
        low: 5
      },
      techniques_detail: [
        {
          id: 'T1003',
          name: 'OS Credential Dumping',
          status: 'success',
          description: 'Successfully extracted password hashes from memory',
          duration: '5m 30s'
        },
        {
          id: 'T1021',
          name: 'Remote Services',
          status: 'blocked',
          description: 'Attempt to use remote services was blocked by security controls',
          duration: '3m 15s'
        },
        {
          id: 'T1078',
          name: 'Valid Accounts',
          status: 'success',
          description: 'Successfully used compromised credentials for access',
          duration: '4m 45s'
        }
      ]
    },
    {
      id: '2',
      scenarioName: 'Data Exfiltration',
      targetEnvironment: 'Test Environment',
      date: '2024-02-19',
      duration: '30 minutes',
      status: 'completed',
      successRate: 60,
      techniques: {
        total: 8,
        successful: 5,
        failed: 2,
        blocked: 1
      },
      findings: {
        critical: 1,
        high: 2,
        medium: 3,
        low: 2
      },
      techniques_detail: [
        {
          id: 'T1048',
          name: 'Exfiltration Over Alternative Protocol',
          status: 'success',
          description: 'Successfully exfiltrated data over DNS',
          duration: '6m 20s'
        },
        {
          id: 'T1567',
          name: 'Exfiltration Over Web Service',
          status: 'failed',
          description: 'Failed to exfiltrate data over HTTP',
          duration: '4m 10s'
        }
      ]
    }
  ];

  const toggleExpand = (id: string) => {
    setExpandedResult(expandedResult === id ? null : id);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'status-success';
      case 'failed':
        return 'status-error';
      case 'blocked':
        return 'status-blocked';
      default:
        return 'status-default';
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
    <div className="emulation-view">
      <Sidebar />
      <div className="emulation-view-content">
        <div className="emulation-view-header">
          <div className="header-content">
            <h1>Attack Simulation Results</h1>
            <p>View and analyze past attack emulation results</p>
          </div>
          <div className="header-actions">
            <select 
              className="time-filter"
              value={timeFilter}
              onChange={(e) => setTimeFilter(e.target.value)}
            >
              <option value="all">All Time</option>
              <option value="week">Last 7 Days</option>
              <option value="month">Last 30 Days</option>
              <option value="quarter">Last Quarter</option>
            </select>
          </div>
        </div>

        <div className="results-summary">
          <div className="summary-card">
            <Target className="summary-icon" />
            <div className="summary-content">
              <span className="summary-value">{results.length}</span>
              <span className="summary-label">Total Simulations</span>
            </div>
          </div>
          <div className="summary-card">
            <CheckCircle2 className="summary-icon" />
            <div className="summary-content">
              <span className="summary-value">
                {Math.round(results.reduce((acc, curr) => acc + curr.successRate, 0) / results.length)}%
              </span>
              <span className="summary-label">Average Success Rate</span>
            </div>
          </div>
          <div className="summary-card">
            <AlertTriangle className="summary-icon" />
            <div className="summary-content">
              <span className="summary-value">
                {results.reduce((acc, curr) => acc + curr.findings.critical + curr.findings.high, 0)}
              </span>
              <span className="summary-label">Critical/High Findings</span>
            </div>
          </div>
        </div>

        <div className="results-list">
          {results.map(result => (
            <div key={result.id} className="result-card">
              <div 
                className="result-header"
                onClick={() => toggleExpand(result.id)}
              >
                <div className="result-main">
                  <div className="result-info">
                    <h3>{result.scenarioName}</h3>
                    <span className="target-env">{result.targetEnvironment}</span>
                  </div>
                  <div className="result-metrics">
                    <span className={`completion-rate ${result.successRate >= 70 ? 'high' : 
                      result.successRate >= 40 ? 'medium' : 'low'}`}>
                      {result.successRate}% Success
                    </span>
                  </div>
                </div>
                <div className="result-meta">
                  <span className="meta-item">
                    <Clock size={16} />
                    {formatDate(result.date)}
                  </span>
                  <span className={`status-badge ${getStatusColor(result.status)}`}>
                    {result.status}
                  </span>
                  <ChevronDown className={`expand-icon ${expandedResult === result.id ? 'rotated' : ''}`} />
                </div>
              </div>

              {expandedResult === result.id && (
                <div className="result-details">
                  <div className="stats-grid">
                    <div className="stat-card">
                      <h4>Techniques Executed</h4>
                      <div className="technique-stats">
                        <div className="stat-item">
                          <span className="stat-label">Successful</span>
                          <span className="stat-value success">{result.techniques.successful}</span>
                        </div>
                        <div className="stat-item">
                          <span className="stat-label">Failed</span>
                          <span className="stat-value failed">{result.techniques.failed}</span>
                        </div>
                        <div className="stat-item">
                          <span className="stat-label">Blocked</span>
                          <span className="stat-value blocked">{result.techniques.blocked}</span>
                        </div>
                      </div>
                    </div>
                    <div className="stat-card">
                      <h4>Security Findings</h4>
                      <div className="findings-stats">
                        <div className="finding-item critical">
                          <span className="finding-count">{result.findings.critical}</span>
                          <span className="finding-label">Critical</span>
                        </div>
                        <div className="finding-item high">
                          <span className="finding-count">{result.findings.high}</span>
                          <span className="finding-label">High</span>
                        </div>
                        <div className="finding-item medium">
                          <span className="finding-count">{result.findings.medium}</span>
                          <span className="finding-label">Medium</span>
                        </div>
                        <div className="finding-item low">
                          <span className="finding-count">{result.findings.low}</span>
                          <span className="finding-label">Low</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="techniques-detail">
                    <h4>Executed Techniques</h4>
                    <div className="techniques-list">
                      {result.techniques_detail.map(technique => (
                        <div key={technique.id} className="technique-item">
                          <div className="technique-header">
                            <div className="technique-info">
                              <span className="technique-id">{technique.id}</span>
                              <h5>{technique.name}</h5>
                            </div>
                            <span className={`technique-status ${getStatusColor(technique.status)}`}>
                              {technique.status}
                            </span>
                          </div>
                          <p className="technique-description">{technique.description}</p>
                          <div className="technique-meta">
                            <span>
                              <Clock size={14} />
                              Duration: {technique.duration}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="result-actions">
                    <button className="action-button">
                      <BarChart2 size={18} />
                      View Analysis
                    </button>
                    <button className="action-button">
                      <FileText size={18} />
                      Full Report
                    </button>
                    <button className="action-button">
                      <Download size={18} />
                      Export Data
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default EmulationView;