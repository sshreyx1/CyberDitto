import React, { useState } from 'react';
import Sidebar from './Sidebar';
import './Emulation-Run.css';
import { 
  Shield, 
  Target, 
  Play, 
  AlertTriangle,
  Server,
  ChevronDown,
  CheckCircle2,
  XCircle,
  Activity,
  Eye
} from 'lucide-react';

interface AttackScenario {
  id: string;
  name: string;
  description: string;
  techniques: string[];
  estimatedTime: string;
  riskLevel: 'high' | 'medium' | 'low';
  targetOS: string[];
}

interface DigitalTwin {
  id: string;
  name: string;
  status: 'running' | 'stopped';
  endpoints: number;
  os: string;
}

const EmulationRun: React.FC = () => {
  const [selectedTwin, setSelectedTwin] = useState<string>('');
  const [selectedScenario, setSelectedScenario] = useState<string>('');
  const [isRunning, setIsRunning] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);
  const [showAdvanced, setShowAdvanced] = useState(false);

  const digitalTwins: DigitalTwin[] = [
    {
      id: '1',
      name: 'Production Environment',
      status: 'running',
      endpoints: 5,
      os: 'Windows Server 2019'
    },
    {
      id: '2',
      name: 'Test Environment',
      status: 'running',
      endpoints: 3,
      os: 'Windows Server 2016'
    }
  ];

  const attackScenarios: AttackScenario[] = [
    {
      id: '1',
      name: 'Credential Access and Lateral Movement',
      description: 'Simulates an attacker attempting to harvest credentials and move laterally through the network',
      techniques: ['T1003', 'T1021', 'T1078'],
      estimatedTime: '30-45 minutes',
      riskLevel: 'high',
      targetOS: ['Windows Server 2019', 'Windows Server 2016']
    },
    {
      id: '2',
      name: 'Data Exfiltration',
      description: 'Simulates data theft attempts using various exfiltration techniques',
      techniques: ['T1048', 'T1567', 'T1020'],
      estimatedTime: '20-30 minutes',
      riskLevel: 'medium',
      targetOS: ['Windows Server 2019', 'Windows Server 2016']
    },
    {
      id: '3',
      name: 'Persistence Mechanisms',
      description: 'Tests various persistence techniques used by attackers to maintain access',
      techniques: ['T1547', 'T1053', 'T1136'],
      estimatedTime: '25-35 minutes',
      riskLevel: 'high',
      targetOS: ['Windows Server 2019']
    }
  ];

  const executionSteps = [
    'Environment Validation',
    'Preparing Attack Vectors',
    'Executing Techniques',
    'Monitoring Responses',
    'Collecting Results'
  ];

  const startEmulation = () => {
    setIsRunning(true);
    // Simulate progression through steps
    let step = 0;
    const interval = setInterval(() => {
      if (step < executionSteps.length - 1) {
        step++;
        setCurrentStep(step);
      } else {
        clearInterval(interval);
      }
    }, 3000);
  };

  return (
    <div className="emulation-run">
      <Sidebar />
      <div className="emulation-run-content">
        <div className="emulation-run-header">
          <h1>Run Attack Emulation</h1>
          <p>Select environment and attack scenario to begin simulation</p>
        </div>

        <div className="emulation-grid">
          {/* Digital Twin Selection */}
          <div className="emulation-section">
            <h2><Server className="section-icon" /> Select Digital Twin</h2>
            <div className="twin-list">
              {digitalTwins.map(twin => (
                <div 
                  key={twin.id}
                  className={`twin-item ${selectedTwin === twin.id ? 'selected' : ''} ${
                    twin.status === 'stopped' ? 'disabled' : ''
                  }`}
                  onClick={() => twin.status === 'running' && setSelectedTwin(twin.id)}
                >
                  <div className="twin-info">
                    <h3>{twin.name}</h3>
                    <p>{twin.endpoints} endpoints • {twin.os}</p>
                  </div>
                  <span className={`twin-status ${twin.status}`}>
                    {twin.status}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Attack Scenario Selection */}
          <div className="emulation-section">
            <h2><Target className="section-icon" /> Select Attack Scenario</h2>
            <div className="scenario-list">
              {attackScenarios.map(scenario => (
                <div 
                  key={scenario.id}
                  className={`scenario-item ${selectedScenario === scenario.id ? 'selected' : ''}`}
                  onClick={() => setSelectedScenario(scenario.id)}
                >
                  <div className="scenario-header">
                    <h3>{scenario.name}</h3>
                    <span className={`risk-level ${scenario.riskLevel}`}>
                      {scenario.riskLevel}
                    </span>
                  </div>
                  <p>{scenario.description}</p>
                  <div className="scenario-meta">
                    <span>
                      <Activity size={16} /> {scenario.estimatedTime}
                    </span>
                    <span>
                      <Shield size={16} /> {scenario.techniques.length} techniques
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Advanced Options */}
          <div className="emulation-section">
            <div
              className="section-header-toggle"
              onClick={() => setShowAdvanced(!showAdvanced)}
            >
              <h2><AlertTriangle className="section-icon" /> Advanced Options</h2>
              <ChevronDown className={`toggle-icon ${showAdvanced ? 'rotated' : ''}`} />
            </div>
            {showAdvanced && (
              <div className="advanced-options">
                <div className="option-group">
                  <label>
                    <input type="checkbox" /> Enable Cleanup After Execution
                  </label>
                  <label>
                    <input type="checkbox" /> Collect Extended Telemetry
                  </label>
                  <label>
                    <input type="checkbox" /> Generate Detailed Logs
                  </label>
                </div>
                <div className="execution-params">
                  <div className="param-group">
                    <label>Execution Timeout</label>
                    <select>
                      <option>30 minutes</option>
                      <option>1 hour</option>
                      <option>2 hours</option>
                    </select>
                  </div>
                  <div className="param-group">
                    <label>Execution Mode</label>
                    <select>
                      <option>Sequential</option>
                      <option>Parallel</option>
                    </select>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Execution Status */}
          <div className="emulation-section">
            <h2><Activity className="section-icon" /> Execution Status</h2>
            <div className="execution-status">
              {!isRunning ? (
                <div className="status-message">
                  <AlertTriangle size={40} />
                  <p>Select environment and scenario to begin emulation</p>
                </div>
              ) : (
                <div className="execution-progress">
                  {executionSteps.map((step, index) => (
                    <div 
                      key={index}
                      className={`progress-step ${
                        index === currentStep ? 'current' : 
                        index < currentStep ? 'completed' : ''
                      }`}
                    >
                      <div className="step-indicator">
                        {index < currentStep ? (
                          <CheckCircle2 size={20} />
                        ) : index === currentStep ? (
                          <Activity size={20} />
                        ) : (
                          <div className="step-number">{index + 1}</div>
                        )}
                      </div>
                      <span>{step}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
            <button 
              className="start-emulation"
              disabled={!selectedTwin || !selectedScenario || isRunning}
              onClick={startEmulation}
            >
              {isRunning ? (
                <>
                  <Activity size={18} /> Emulation in Progress...
                </>
              ) : (
                <>
                  <Play size={18} /> Start Emulation
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EmulationRun;