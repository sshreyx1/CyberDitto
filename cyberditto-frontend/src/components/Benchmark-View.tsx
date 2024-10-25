import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './Benchmark-View.css';
import { 
  BarChart2, 
  Shield, 
  AlertTriangle, 
  CheckCircle2, 
  ChevronDown, 
  Download,
  Calendar,
  Server,
  XCircle
} from 'lucide-react';

interface BenchmarkResult {
  id: string;
  date: string;
  endpointName: string;
  profile: string;
  complianceScore: number;
  status: 'completed' | 'failed';
  findings: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  controls: {
    passed: number;
    failed: number;
    notApplicable: number;
    total: number;
  };
}

const BenchmarkView: React.FC = () => {
  const [expandedResult, setExpandedResult] = useState<string | null>(null);
  const [timeFilter, setTimeFilter] = useState('all');

  const results: BenchmarkResult[] = [
    {
      id: '1',
      date: '2024-02-20',
      endpointName: 'PROD-DC-01',
      profile: 'CIS Windows Server 2019 Level 1',
      complianceScore: 85,
      status: 'completed',
      findings: {
        critical: 2,
        high: 5,
        medium: 8,
        low: 12
      },
      controls: {
        passed: 156,
        failed: 24,
        notApplicable: 10,
        total: 190
      }
    },
    {
      id: '2',
      date: '2024-02-19',
      endpointName: 'PROD-DC-02',
      profile: 'CIS Windows Server 2019 Level 1',
      complianceScore: 78,
      status: 'completed',
      findings: {
        critical: 3,
        high: 7,
        medium: 10,
        low: 15
      },
      controls: {
        passed: 140,
        failed: 40,
        notApplicable: 10,
        total: 190
      }
    },
    {
      id: '3',
      date: '2024-02-18',
      endpointName: 'PROD-APP-01',
      profile: 'CIS Windows Server 2016 Level 1',
      complianceScore: 92,
      status: 'completed',
      findings: {
        critical: 1,
        high: 3,
        medium: 5,
        low: 8
      },
      controls: {
        passed: 145,
        failed: 15,
        notApplicable: 5,
        total: 165
      }
    }
  ];

  const toggleExpand = (id: string) => {
    setExpandedResult(expandedResult === id ? null : id);
  };

  const getScoreColor = (score: number) => {
    if (score >= 90) return 'score-high';
    if (score >= 70) return 'score-medium';
    return 'score-low';
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  return (
    <div className="benchmark-view">
      <Sidebar />
      <div className="benchmark-view-content">
        <div className="benchmark-view-header">
          <div className="header-content">
            <h1>Benchmark Results</h1>
            <p>View and analyze CIS benchmark results across your endpoints</p>
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
            <div className="summary-icon">
              <Server />
            </div>
            <div className="summary-content">
              <span className="summary-value">{results.length}</span>
              <span className="summary-label">Total Scans</span>
            </div>
          </div>
          <div className="summary-card">
            <div className="summary-icon">
              <CheckCircle2 />
            </div>
            <div className="summary-content">
              <span className="summary-value">
                {Math.round(results.reduce((acc, curr) => acc + curr.complianceScore, 0) / results.length)}%
              </span>
              <span className="summary-label">Average Compliance</span>
            </div>
          </div>
          <div className="summary-card">
            <div className="summary-icon">
              <AlertTriangle />
            </div>
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
                    <h3>{result.endpointName}</h3>
                    <span className="result-profile">{result.profile}</span>
                  </div>
                  <div className={`compliance-score ${getScoreColor(result.complianceScore)}`}>
                    {result.complianceScore}%
                  </div>
                </div>
                <div className="result-meta">
                  <div className="meta-item">
                    <Calendar size={16} />
                    {formatDate(result.date)}
                  </div>
                  <ChevronDown className={`expand-icon ${expandedResult === result.id ? 'rotated' : ''}`} />
                </div>
              </div>

              {expandedResult === result.id && (
                <div className="result-details">
                  <div className="findings-grid">
                    <div className="finding-card critical">
                      <span className="finding-count">{result.findings.critical}</span>
                      <span className="finding-label">Critical</span>
                    </div>
                    <div className="finding-card high">
                      <span className="finding-count">{result.findings.high}</span>
                      <span className="finding-label">High</span>
                    </div>
                    <div className="finding-card medium">
                      <span className="finding-count">{result.findings.medium}</span>
                      <span className="finding-label">Medium</span>
                    </div>
                    <div className="finding-card low">
                      <span className="finding-count">{result.findings.low}</span>
                      <span className="finding-label">Low</span>
                    </div>
                  </div>

                  <div className="controls-summary">
                    <h4>Controls Summary</h4>
                    <div className="controls-grid">
                      <div className="control-item passed">
                        <span className="control-count">{result.controls.passed}</span>
                        <span className="control-label">Passed</span>
                      </div>
                      <div className="control-item failed">
                        <span className="control-count">{result.controls.failed}</span>
                        <span className="control-label">Failed</span>
                      </div>
                      <div className="control-item na">
                        <span className="control-count">{result.controls.notApplicable}</span>
                        <span className="control-label">N/A</span>
                      </div>
                    </div>
                  </div>

                  <div className="result-actions">
                    <button className="action-button">
                      <BarChart2 size={18} />
                      Detailed Report
                    </button>
                    <button className="action-button">
                      <Download size={18} />
                      Export Results
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

export default BenchmarkView;