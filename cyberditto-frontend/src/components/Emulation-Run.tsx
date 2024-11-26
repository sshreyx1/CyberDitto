// src/components/Emulation-Run.tsx

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import Sidebar from './Sidebar';
import './Emulation-Run.css';
import { 
    Shield, 
    Target, 
    Play, 
    AlertCircle,
    Server,
    ChevronDown,
    CheckCircle2,
    Activity,
    Loader
} from 'lucide-react';
import { emulationApi, EmulationMode, EmulationStatus } from '../services/Emulation';
import { digitalTwinApi } from '../services/DigitalTwin';
import type { ProcessedDeployment } from '../services/DigitalTwin';

const EmulationRun: React.FC = () => {
    const navigate = useNavigate();
    const [selectedTwin, setSelectedTwin] = useState<string>('');
    const [mode, setMode] = useState<EmulationMode>('tactic');
    const [target, setTarget] = useState<string>('');
    const [testNumbers, setTestNumbers] = useState<string>('');
    const [isRunning, setIsRunning] = useState(false);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [digitalTwins, setDigitalTwins] = useState<ProcessedDeployment[]>([]);
    const [emulationStatus, setEmulationStatus] = useState<EmulationStatus | null>(null);

    useEffect(() => {
        loadDigitalTwins();
        // Poll for updates every 30 seconds
        const interval = setInterval(loadDigitalTwins, 30000);
        return () => clearInterval(interval);
    }, []);

    const loadDigitalTwins = async () => {
        try {
            setError(null);
            const deployments = await digitalTwinApi.getAllDeployments();
            // Filter to only show running VMs
            const runningDeployments = deployments.filter(d => 
                d.status.toLowerCase() === 'running'
            );
            setDigitalTwins(runningDeployments);
        } catch (error) {
            console.error('Failed to load digital twins:', error);
            setError('Failed to load available environments');
        } finally {
            setLoading(false);
        }
    };

    const handleModeChange = (newMode: EmulationMode) => {
        setMode(newMode);
        setTarget('');
        setTestNumbers('');
    };

    const startEmulation = async () => {
        try {
            setIsRunning(true);
            setError(null);

            const { execution_id } = await emulationApi.startEmulation({
                deployment_id: selectedTwin,
                mode,
                target,
                test_numbers: testNumbers || undefined
            });

            await emulationApi.pollEmulationStatus(
                execution_id,
                (status) => {
                    setEmulationStatus(status);
                    if (status.phase === 'completed') {
                        navigate(`/emulation/view?id=${execution_id}`);
                    }
                },
                (error) => {
                    setError('Emulation failed: ' + error.message);
                    setIsRunning(false);
                }
            );
        } catch (error) {
            setError('Failed to start emulation: ' + (error instanceof Error ? error.message : 'Unknown error'));
            setIsRunning(false);
        }
    };

    if (loading) {
        return (
            <div className="emulation-run">
                <Sidebar />
                <div className="emulation-run-content">
                    <div className="emulation-loading">
                        <Loader className="spin" size={40} />
                        <h3>Loading Environments</h3>
                        <p>Please wait while we fetch available systems...</p>
                    </div>
                </div>
            </div>
        );
    }

    if (digitalTwins.length === 0) {
        return (
            <div className="emulation-run">
                <Sidebar />
                <div className="emulation-run-content">
                    <div className="emulation-empty">
                        <Server size={48} />
                        <h3>No Running Environments Available</h3>
                        <p>Start a digital twin environment before running attack simulations.</p>
                        <button 
                            className="emulation-view-button"
                            onClick={() => navigate('/digitaltwin/view')}
                        >
                            View Digital Twins
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="emulation-run">
            <Sidebar />
            <div className="emulation-run-content">
                <div className="emulation-run-header">
                    <h1>Run Attack Emulation</h1>
                    <p>Select environment and attack parameters to begin simulation</p>
                </div>

                {error && (
                    <div className="emulation-error">
                        <AlertCircle size={20} />
                        <span>{error}</span>
                    </div>
                )}

                <div className="emulation-grid">
                    {/* Digital Twin Selection */}
                    <div className="emulation-section">
                        <h2><Server className="section-icon" /> Select Environment</h2>
                        <div className="twin-list">
                            {digitalTwins.map(twin => (
                                <div 
                                    key={twin.id}
                                    className={`twin-item ${selectedTwin === twin.id ? 'selected' : ''}`}
                                    onClick={() => setSelectedTwin(twin.id)}
                                >
                                    <div className="twin-info">
                                        <h3>{twin.name}</h3>
                                        <p>
                                            CPU: {digitalTwinApi.formatResourceUsage(twin.resourceUsage.cpu)} | 
                                            Memory: {digitalTwinApi.formatResourceUsage(twin.resourceUsage.memory)}
                                        </p>
                                    </div>
                                    <span className="twin-status running">
                                        Running
                                    </span>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Attack Configuration */}
                    <div className="emulation-section">
                        <h2><Target className="section-icon" /> Configure Attack</h2>
                        <div className="attack-config">
                            <div className="config-group">
                                <label>Attack Mode</label>
                                <select 
                                    value={mode} 
                                    onChange={(e) => handleModeChange(e.target.value as EmulationMode)}
                                    disabled={isRunning}
                                >
                                    <option value="tactic">By Tactic</option>
                                    <option value="technique">By Technique</option>
                                    <option value="custom">Custom Chain</option>
                                </select>
                            </div>
                            
                            <div className="config-group">
                                <label>
                                    {mode === 'tactic' ? 'Tactic Name' : 
                                     mode === 'technique' ? 'Technique ID' : 
                                     'Technique Chain'}
                                </label>
                                <input 
                                    type="text" 
                                    value={target}
                                    onChange={(e) => setTarget(e.target.value)}
                                    placeholder={
                                        mode === 'tactic' ? 'e.g., discovery' : 
                                        mode === 'technique' ? 'e.g., T1033' : 
                                        'e.g., T1033,T1087'
                                    }
                                    disabled={isRunning}
                                />
                            </div>

                            {mode === 'technique' && (
                                <div className="config-group">
                                    <label>Test Numbers (Optional)</label>
                                    <input 
                                        type="text" 
                                        value={testNumbers}
                                        onChange={(e) => setTestNumbers(e.target.value)}
                                        placeholder="e.g., 1,2,3"
                                        disabled={isRunning}
                                    />
                                </div>
                            )}
                        </div>
                    </div>

                    {/* Execution Status */}
                    <div className="emulation-section">
                        <h2><Activity className="section-icon" /> Execution Status</h2>
                        <div className="execution-status">
                            {!isRunning ? (
                                <div className="status-message">
                                    <AlertCircle size={40} />
                                    <p>Select environment and configure attack parameters</p>
                                </div>
                            ) : (
                                <div className="execution-progress">
                                    {['preparing', 'copying', 'executing', 'collecting'].map((phase, index) => (
                                        <div 
                                            key={phase}
                                            className={`progress-step ${
                                                emulationStatus?.phase === phase ? 'current' : 
                                                (emulationStatus?.progress || 0) > index * 25 ? 'completed' : ''
                                            }`}
                                        >
                                            <div className="step-indicator">
                                                {(emulationStatus?.progress || 0) > index * 25 ? 
                                                    <CheckCircle2 size={20} /> : 
                                                    emulationStatus?.phase === phase ?
                                                    <Activity size={20} /> :
                                                    <span className="step-number">{index + 1}</span>
                                                }
                                            </div>
                                            <span>{phase.charAt(0).toUpperCase() + phase.slice(1)}</span>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>

                        <button 
                            className="start-emulation"
                            disabled={!selectedTwin || !target || isRunning}
                            onClick={startEmulation}
                        >
                            {isRunning ? (
                                <>
                                    <Loader className="spin" size={18} />
                                    Emulation in Progress...
                                </>
                            ) : (
                                <>
                                    <Play size={18} />
                                    Start Emulation
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