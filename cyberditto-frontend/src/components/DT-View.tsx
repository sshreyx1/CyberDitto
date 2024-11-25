import React, { useState, useEffect } from 'react';
import Sidebar from './Sidebar';
import './DT-View.css';
import {
    Server,
    Activity,
    PlayCircle,
    StopCircle,
    BarChart,
    AlertCircle,
    RefreshCcw,
    Loader,
    Command,
    Calendar,
    Clock,
    HardDrive
} from 'lucide-react';
import { digitalTwinApi } from '../services/DigitalTwin';
import type { ProcessedDeployment } from '../services/DigitalTwin';

const LOCAL_STORAGE_KEY = 'digital_twin_deployments';

const DTView: React.FC = () => {
    const [deployments, setDeployments] = useState<ProcessedDeployment[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [refreshing, setRefreshing] = useState(false);
    const [actionInProgress, setActionInProgress] = useState<string | null>(null);

    const loadFromLocalStorage = (): ProcessedDeployment[] => {
        try {
            const stored = localStorage.getItem(LOCAL_STORAGE_KEY);
            if (stored) {
                return JSON.parse(stored);
            }
        } catch (err) {
            console.error('Error loading from local storage:', err);
        }
        return [];
    };

    const saveToLocalStorage = (data: ProcessedDeployment[]) => {
        try {
            localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(data));
        } catch (err) {
            console.error('Save error:', err);
        }
    };

    const updateLocalDeployment = (deploymentId: string, updates: Partial<ProcessedDeployment>) => {
        setDeployments(prevDeployments => {
            const newDeployments = prevDeployments.map(dep =>
                dep.id === deploymentId ? { ...dep, ...updates } : dep
            );
            saveToLocalStorage(newDeployments);
            return newDeployments;
        });
    };

    const loadDeployments = async () => {
        if (refreshing) return;
        setRefreshing(true);

        try {
            const data = await digitalTwinApi.getAllDeployments();
            setDeployments(data);
            saveToLocalStorage(data);
        } catch (err) {
            console.error('Load error:', err);
            const storedDeployments = loadFromLocalStorage();
            if (storedDeployments.length > 0) {
                setDeployments(storedDeployments);
            }
        } finally {
            setRefreshing(false);
            setLoading(false);
        }
    };

    useEffect(() => {
        loadDeployments();
        const interval = setInterval(loadDeployments, 30000);
        return () => clearInterval(interval);
    }, []);

    const handleRefresh = () => {
        if (!refreshing) {
            loadDeployments();
        }
    };

    const handleStartInstance = async (deploymentId: string) => {
        if (!deploymentId) return;

        try {
            setActionInProgress(deploymentId);
            await digitalTwinApi.startInstance(deploymentId);
            updateLocalDeployment(deploymentId, {
                status: 'running',
                progress: 100,
                message: 'Virtual environment is running'
            } as Partial<ProcessedDeployment>);
            await loadDeployments();
        } catch (err) {
            console.error('Error starting instance:', err);
            // Only set UI error for critical failures
            if (err instanceof Error && err.message.includes('not found')) {
                setError('Instance not found');
            }
        } finally {
            setActionInProgress(null);
        }
    };

    const handleStopInstance = async (deploymentId: string) => {
        if (!deploymentId) return;

        try {
            setActionInProgress(deploymentId);
            await digitalTwinApi.stopInstance(deploymentId);
            updateLocalDeployment(deploymentId, {
                status: 'stopped',
                progress: 0,
                message: 'Virtual environment is stopped'
            } as Partial<ProcessedDeployment>);
            await loadDeployments();
        } catch (err) {
            console.error('Error stopping instance:', err);
            // Only set UI error for critical failures
            if (err instanceof Error && err.message.includes('not found')) {
                setError('Instance not found');
            }
        } finally {
            setActionInProgress(null);
        }
    };

    const formatDate = (dateString: string | null) => {
        try {
            if (!dateString) return 'Unknown Date';

            return new Date(dateString).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        } catch (err) {
            console.error('Error formatting date:', err);
            return 'Invalid Date';
        }
    };

    const formatMemoryValue = (bytes: number | undefined) => {
        if (typeof bytes !== 'number' || bytes <= 0) return '0 GB';
        const gb = bytes / (1024 * 1024 * 1024);
        return `${Math.round(gb * 100) / 100} GB`;
    };

    const formatResourcePercent = (value: number | undefined) => {
        if (typeof value !== 'number') return '0%';
        const normalizedValue = Math.max(0, Math.min(100, value));
        return `${Math.round(normalizedValue)}%`;
    };

    const getStatusColor = (status: string) => {
        switch (status?.toLowerCase()) {
            case 'running': return 'dtv-status-running';
            case 'stopped': return 'dtv-status-stopped';
            case 'error': return 'dtv-status-error';
            case 'preparing': return 'dtv-status-preparing';
            case 'deploying': return 'dtv-status-deploying';
            case 'saved': return 'dtv-status-saved';
            case 'initializing': return 'dtv-status-preparing';
            default: return 'dtv-status-idle';
        }
    };

    const isActionable = (status: string) => {
        const statusLower = status?.toLowerCase() || '';
        return ['running', 'stopped', 'saved', 'error'].includes(statusLower);
    };

    const getActionButton = (deployment: ProcessedDeployment) => {
        const isInProgress = actionInProgress === deployment.id;
        const disabled = !isActionable(deployment.status) || isInProgress;

        if (deployment.status === 'running') {
            return (
                <button
                    className="dtv-action-button dtv-stop"
                    onClick={() => handleStopInstance(deployment.id)}
                    disabled={disabled}
                >
                    {isInProgress ? (
                        <><Loader className="dtv-spinning" size={18} />Stopping...</>
                    ) : (
                        <><StopCircle size={18} />Stop Environment</>
                    )}
                </button>
            );
        }

        return (
            <button
                className="dtv-action-button dtv-start"
                onClick={() => handleStartInstance(deployment.id)}
                disabled={disabled}
            >
                {isInProgress ? (
                    <><Loader className="dtv-spinning" size={18} />Starting...</>
                ) : (
                    <><PlayCircle size={18} />Start Environment</>
                )}
            </button>
        );
    };

    return (
        <div className="dtv-wrapper">
            <Sidebar />
            <div className="dtv-content">
                <div className="dtv-header">
                    <div>
                        <h1>Digital Twin Environments</h1>
                        <p>Monitor and manage your virtual system replicas</p>
                    </div>
                    <button
                        className="dtv-refresh-button"
                        onClick={handleRefresh}
                        disabled={refreshing || loading}
                    >
                        <RefreshCcw size={18} className={refreshing ? 'dtv-spinning' : ''} />
                        Refresh
                    </button>
                </div>

                {error && (
                    <div className="dtv-error-message">
                        <AlertCircle />
                        {error}
                    </div>
                )}

                {loading ? (
                    <div className="dtv-loading-container">
                        <div className="dtv-loading-content">
                            <Loader size={40} className="dtv-spinning" />
                            <div className="dtv-loading-text">
                                <h3>Loading Digital Twins</h3>
                                <p>Please wait while we fetch your environments...</p>
                            </div>
                        </div>
                    </div>
                ) : deployments.length === 0 ? (
                    <div className="dtv-no-deployments">
                        <Server size={48} />
                        <h3>No Digital Twins Found</h3>
                        <p>Create your first digital twin by scanning your system.</p>
                    </div>
                ) : (
                    <div className="dtv-twins-grid">
                        {deployments.map(deployment => (
                            <div key={deployment.id} className="dtv-twin-card">
                                <div className="dtv-twin-card-header">
                                    <div className="dtv-twin-title">
                                        <Server className="dtv-twin-icon" />
                                        <h2>{deployment.name}</h2>
                                    </div>
                                    <span className={`dtv-twin-status ${getStatusColor(deployment.status)}`}>
                                        {deployment.status}
                                    </span>
                                </div>

                                <div className="dtv-twin-details">
                                    <div className="dtv-detail-row">
                                        <div className="dtv-vagrant-id">
                                            <Command size={16} />
                                            <span>VM ID: {deployment.vagrantId || 'Not assigned'}</span>
                                        </div>
                                    </div>
                                    <div className="dtv-detail-row">
                                        <span className="dtv-detail-label">
                                            <Calendar size={16} /> Created
                                        </span>
                                        <span className="dtv-detail-value">
                                            {formatDate(deployment.createdAt)}
                                        </span>
                                    </div>
                                    <div className="dtv-detail-row">
                                        <span className="dtv-detail-label">
                                            <Clock size={16} /> Last Active
                                        </span>
                                        <span className="dtv-detail-value">
                                            {formatDate(deployment.lastActive)}
                                        </span>
                                    </div>
                                    {deployment.scanResult && (
                                        <>
                                            <div className="dtv-detail-row">
                                                <span className="dtv-detail-label">
                                                    <HardDrive size={16} /> System
                                                </span>
                                                <span className="dtv-detail-value">
                                                    {deployment.scanResult.systemInfo.osVersion || 'Unknown'}
                                                </span>
                                            </div>
                                            <div className="dtv-detail-row">
                                                <span className="dtv-detail-label">Memory</span>
                                                <span className="dtv-detail-value">
                                                    {formatMemoryValue(deployment.scanResult.systemInfo.memory)}
                                                </span>
                                            </div>
                                        </>
                                    )}
                                </div>

                                <div className="dtv-resource-usage">
                                    <h3>
                                        <Activity className="dtv-resource-icon" />
                                        Resource Usage
                                    </h3>
                                    <div className="dtv-usage-bars">
                                        <div className="dtv-usage-item">
                                            <label>CPU</label>
                                            <div className="dtv-progress-bar">
                                                <div
                                                    className="dtv-progress-fill"
                                                    style={{ width: formatResourcePercent(deployment.resourceUsage.cpu) }}
                                                ></div>
                                            </div>
                                            <span>{formatResourcePercent(deployment.resourceUsage.cpu)}</span>
                                        </div>
                                        <div className="dtv-usage-item">
                                            <label>Memory</label>
                                            <div className="dtv-progress-bar">
                                                <div
                                                    className="dtv-progress-fill"
                                                    style={{ width: formatResourcePercent(deployment.resourceUsage.memory) }}
                                                ></div>
                                            </div>
                                            <span>{formatResourcePercent(deployment.resourceUsage.memory)}</span>
                                        </div>
                                        <div className="dtv-usage-item">
                                            <label>Storage</label>
                                            <div className="dtv-progress-bar">
                                                <div
                                                    className="dtv-progress-fill"
                                                    style={{ width: formatResourcePercent(deployment.resourceUsage.disk) }}
                                                ></div>
                                            </div>
                                            <span>{formatResourcePercent(deployment.resourceUsage.disk)}</span>
                                        </div>
                                    </div>
                                </div>

                                <div className="dtv-twin-actions">
                                    {getActionButton(deployment)}
                                    <button
                                        className="dtv-action-button dtv-monitor"
                                        onClick={() => {/* Add monitoring logic */}}
                                    >
                                        <BarChart size={18} />
                                        View Metrics
                                    </button>
                                </div>

                                {(deployment.message || deployment.error) && (
                                    <div className={`dtv-status-message ${deployment.error ? 'dtv-error' : ''}`}>
                                        {deployment.error ? <AlertCircle size={18} /> : <Activity size={18} />}
                                        <p>{deployment.error || deployment.message}</p>
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

export default DTView;