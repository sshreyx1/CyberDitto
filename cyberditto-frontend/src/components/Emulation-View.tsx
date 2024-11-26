// src/components/Emulation-View.tsx

import React, { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import Sidebar from './Sidebar';
import './Emulation-View.css';
import { 
  Target, 
  Shield, 
  AlertTriangle, 
  CheckCircle2, 
  XCircle,
  Clock,
  Download,
  ChevronDown,
  BarChart2,
  FileText,
  Activity
} from 'lucide-react';
import { 
  emulationApi,
  EmulationResult,
  TestResult
} from '../services/Emulation';

const EmulationView: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [expandedResult, setExpandedResult] = useState<string | null>(null);
  const [timeFilter, setTimeFilter] = useState('all');
  const [results, setResults] = useState<EmulationResult[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // If there's a specific execution ID in URL params, expand that result
    const executionId = searchParams.get('id');
    if (executionId) {
      setExpandedResult(executionId);
    }

    // Load emulation results
    loadResults();
  }, [searchParams]);

  const loadResults = async () => {
    try {
      const allResults = await emulationApi.getAllEmulations();
      setResults(allResults);
    } catch (error) {
      console.error('Failed to load emulation results:', error);
    } finally {
      setLoading(false);
    }
  };

  const toggleExpand = (id: string) => {
    setExpandedResult(expandedResult === id ? null : id);
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'pass':
      case 'completed':
        return 'status-success';
      case 'fail':
      case 'error':
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

  const calculateSuccessRate = (result: EmulationResult) => {
    if (result.summary) {
      return result.summary.success_rate;
    }
    return 0;
  };

  if (loading) {
    return (
      <div className="emulation-view">
        <Sidebar />
        <div className="emulation-view-content">
          <div className="loading">
            <Activity className="spin" />
            Loading results...
          </div>
        </div>
      </div>
    );
  }

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
                {Math.round(results.reduce((acc, curr) => acc + calculateSuccessRate(curr), 0) / results.length)}%
              </span>
              <span className="summary-label">Average Success Rate</span>
            </div>
          </div>
          <div className="summary-card">
            <AlertTriangle className="summary-icon" />
            <div className="summary-content">
              <span className="summary-value">
                {results.reduce((acc, curr) => acc + curr.summary.passed_tests, 0)}
              </span>
              <span className="summary-label">Total Passed Tests</span>
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
                    <h3>Emulation {result.id}</h3>
                    <span className="target-env">Deployment ID: {result.deploy_id}</span>
                  </div>
                  <div className="result-metrics">
                    <span className={`completion-rate ${
                      result.summary.success_rate >= 70 ? 'high' : 
                      result.summary.success_rate >= 40 ? 'medium' : 'low'
                    }`}>
                      {result.summary.success_rate}% Success
                    </span>
                  </div>
                </div>
                <div className="result-meta">
                  <span className="meta-item">
                    <Clock size={16} />
                    {formatDate(result.created_at)}
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
                      <h4>Test Summary</h4>
                      <div className="technique-stats">
                        <div className="stat-item">
                          <span className="stat-label">Passed</span>
                          <span className="stat-value success">{result.summary.passed_tests}</span>
                        </div>
                        <div className="stat-item">
                          <span className="stat-label">Failed</span>
                          <span className="stat-value failed">{result.summary.failed_tests}</span>
                        </div>
                        <div className="stat-item">
                          <span className="stat-label">Total</span>
                          <span className="stat-value">{result.summary.total_tests}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="techniques-detail">
                    <h4>Executed Tests</h4>
                    <div className="techniques-list">
                      {result.results.map((test, index) => (
                        <div key={index} className="technique-item">
                          <div className="technique-header">
                            <div className="technique-info">
                              <span className="technique-id">{test.technique}</span>
                              <h5>{test.technique_name}</h5>
                            </div>
                            <span className={`technique-status ${getStatusColor(test.status)}`}>
                              {test.status}
                            </span>
                          </div>
                          {test.error_message && (
                            <p className="technique-description error">{test.error_message}</p>
                          )}
                          <div className="technique-meta">
                            <span>
                              <Clock size={14} />
                              Executed at: {formatDate(test.timestamp)}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="result-actions">
                    <button
                      className="action-button"
                      onClick={() => {
                        // Download results as JSON
                        const dataStr = "data:text/json;charset=utf-8," + 
                          encodeURIComponent(JSON.stringify(result, null, 2));
                        const downloadAnchor = document.createElement('a');
                        downloadAnchor.setAttribute("href", dataStr);
                        downloadAnchor.setAttribute("download", `emulation_${result.id}_results.json`);
                        document.body.appendChild(downloadAnchor);
                        downloadAnchor.click();
                        downloadAnchor.remove();
                      }}
                    >
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

export default EmulationView;