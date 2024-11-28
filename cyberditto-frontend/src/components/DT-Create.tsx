import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Server,
    Scan,
    AlertCircle,
    CheckCircle,
    Activity,
    Loader,
    ArrowRight
} from 'lucide-react';
import { digitalTwinApi } from '../services/DigitalTwin';
import type { 
    ScanProgress, 
    DeploymentStatus, 
    ScanResult
} from '../services/DigitalTwin';
import Sidebar from './Sidebar';
import './DT-Create.css';

interface ScanState {
    scanId: string;
    status: 'idle' | 'scanning' | 'processing' | 'completed' | 'error';
    progress: number;
    stage?: string;
    message?: string;
    error: string;
    result: ScanResult | null;
}

interface DeployState {
    deploymentId: string;
    status: DeploymentStatus['status'];
    progress: number;
    error: string;
    vagrantId: string;
    success: boolean;
    message: string;
}

const DTCreate: React.FC = () => {
    const navigate = useNavigate();

    const [scanState, setScanState] = React.useState<ScanState>({
        scanId: '',
        status: 'idle',
        progress: 0,
        error: '',
        result: null
    });

    const [deployState, setDeployState] = React.useState<DeployState>({
        deploymentId: '',
        status: 'idle',
        progress: 0,
        error: '',
        vagrantId: '',
        success: false,
        message: ''
    });

    const startScan = async () => {
        setScanState(prev => ({
            ...prev,
            status: 'scanning',
            progress: 0,
            error: '',
            message: 'Starting security scan...'
        }));

        try {
            const { scanId } = await digitalTwinApi.startScan();
            setScanState(prev => ({ ...prev, scanId }));

            await digitalTwinApi.pollScan(
                scanId,
                (progress) => {
                    setScanState(prev => ({
                        ...prev,
                        status: progress.phase,
                        progress: progress.progress,
                        stage: progress.stage,
                        message: progress.message,
                        error: progress.error || ''
                    }));

                    if (progress.phase === 'completed') {
                        digitalTwinApi.getScanResult(scanId)
                            .then(result => {
                                setScanState(prev => ({
                                    ...prev,
                                    result,
                                    progress: 100,
                                    message: 'Scan completed successfully'
                                }));
                            })
                            .catch(error => {
                                setScanState(prev => ({
                                    ...prev,
                                    error: 'Failed to fetch scan results'
                                }));
                            });
                    }
                },
                (error) => {
                    setScanState(prev => ({
                        ...prev,
                        status: 'error',
                        error: error.message,
                        progress: 0
                    }));
                }
            );
        } catch (error) {
            setScanState(prev => ({
                ...prev,
                status: 'error',
                error: error instanceof Error ? error.message : 'Failed to start scan',
                progress: 0
            }));
        }
    };

    const deployDigitalTwin = async () => {
        if (!scanState.scanId) return;

        setDeployState(prev => ({
            ...prev,
            status: 'deploying',
            progress: 0,
            error: '',
            success: false,
            message: 'Starting deployment...'
        }));

        try {
            const deploymentId = await digitalTwinApi.deployDigitalTwin(scanState.scanId);
            setDeployState(prev => ({
                ...prev,
                deploymentId,
                message: 'Initializing virtual environment...'
            }));

            await digitalTwinApi.pollDeploymentStatus(
                deploymentId,
                (status) => {
                    setDeployState(prev => ({
                        ...prev,
                        status: status.status,
                        progress: status.progress,
                        error: status.error || '',
                        vagrantId: status.vagrantId || '',
                        message: status.message || 'Processing...',
                        success: status.status === 'running'
                    }));
                },
                (error) => {
                    setDeployState(prev => ({
                        ...prev,
                        status: 'error',
                        error: error.message,
                        success: false
                    }));
                }
            );
        } catch (error) {
            setDeployState(prev => ({
                ...prev,
                status: 'error',
                error: error instanceof Error ? error.message : 'Failed to start deployment',
                success: false
            }));
        }
    };

    const handleViewDeployments = () => {
        navigate('/digitaltwin/view');
    };

    return (
        <div className="dt-create-wrapper">
            <Sidebar />
            <div className="dt-create-content">
                <div className="dt-create-header">
                    <h1>Create Digital Twin</h1>
                    <p>Create a virtual replica of your system environment</p>
                </div>

                <div className="dt-grid">
                    <div className="dt-card">
                        <div className="card-header">
                            <Scan className="icon" />
                            <h2>System Scan</h2>
                        </div>

                        <button
                            className="dt-button"
                            onClick={startScan}
                            disabled={
                                scanState.status === 'scanning' ||
                                deployState.status === 'deploying' ||
                                deployState.status === 'preparing' ||
                                deployState.status === 'initializing'
                            }
                        >
                            {scanState.status === 'scanning' ? (
                                <>
                                    <Loader className="spin" />
                                    Scanning...
                                </>
                            ) : (
                                <>
                                    <Scan />
                                    Start Scan
                                </>
                            )}
                        </button>

                        {scanState.status !== 'idle' && (
                            <div className="progress-section">
                                <div className="progress-container">
                                    <div className="progress-bar">
                                        <div
                                            className="progress-fill"
                                            style={{ width: `${scanState.progress}%` }}
                                        />
                                    </div>
                                    <span className="progress-text">
                                        {Math.round(scanState.progress)}%
                                    </span>
                                </div>
                                {(scanState.stage || scanState.message) && (
                                    <div className="progress-message">
                                        <Activity className="spin" />
                                        <span>{scanState.stage ? `${scanState.stage} - ${scanState.message}` : scanState.message}</span>
                                    </div>
                                )}
                            </div>
                        )}

                        {scanState.error && (
                            <div className="error-message">
                                <AlertCircle />
                                {scanState.error}
                            </div>
                        )}

                        {scanState.result && (
                            <div className="scan-results">
                                <h3>Scan Results</h3>
                                <div className="result-grid">
                                    <div className="result-section">
                                        <h4>System Info</h4>
                                        <p>OS: {scanState.result.systemInfo.osVersion}</p>
                                        <p>CPU: {scanState.result.systemInfo.cpuModel}</p>
                                        <p>Memory: {digitalTwinApi.formatMemory(scanState.result.systemInfo.memory)}</p>
                                        <p>OS: {scanState.result.systemInfo.osName}</p>
                                    </div>

                                    <div className="result-section">
                                        <h4>Security Info</h4>
                                        <p>Firewall: {scanState.result.securityInfo.firewall.enabled ? 'Enabled' : 'Disabled'}</p>
                                        <p>Firewall Rules: {scanState.result.securityInfo.firewall.rules}</p>
                                        <p>Windows Defender: {scanState.result.securityInfo.antivirus.windowsDefender.enabled ? 'Active' : 'Inactive'}</p>
                                        <p>Updates: {scanState.result.securityInfo.updates.length} recent updates</p>
                                    </div>

                                    <button
                                        className="dt-button"
                                        onClick={deployDigitalTwin}
                                        disabled={deployState.status === 'deploying' || deployState.success || deployState.status !== 'idle'}
                                    >
                                        {deployState.status === 'deploying' ? (
                                            <>
                                                <Loader className="spin" />
                                                Deploying...
                                            </>
                                        ) : (
                                            <>
                                                <Server />
                                                Deploy Digital Twin
                                            </>
                                        )}
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>

                    {(deployState.status !== 'idle') && (
                        <div className="dt-card">
                            <div className="card-header">
                                <Server className="icon" />
                                <h2>Deployment Status</h2>
                            </div>

                            <div className="progress-section">
                                <div className="progress-container">
                                    <div className="progress-bar">
                                        <div
                                            className="progress-fill"
                                            style={{ width: `${deployState.progress}%` }}
                                        />
                                    </div>
                                    <span className="progress-text">
                                        {Math.round(deployState.progress)}%
                                    </span>
                                </div>
                                <div className="progress-message">
                                    <Activity className="spin" />
                                    <span>{deployState.message}</span>
                                </div>
                            </div>

                            {deployState.error && (
                                <div className="error-message">
                                    <AlertCircle />
                                    {deployState.error}
                                </div>
                            )}

                            {deployState.success && (
                                <div className="success-message">
                                    <div style={{ display: 'flex', alignItems: 'center', marginBottom: '0.5rem' }}>
                                        <CheckCircle className="CheckCircle" />
                                        <p>Digital Twin deployed successfully!</p>
                                    </div>
                                    <button
                                        className="dt-button view-button"
                                        onClick={handleViewDeployments}
                                    >
                                        View Digital Twins
                                        <ArrowRight className="ArrowRight" />
                                    </button>
                                </div>
                            )}

                            {deployState.status === 'deploying' && (
                                <div className="status-message">
                                    <Activity />
                                    <p>{deployState.message || 'Setting up your virtual environment...'}</p>
                                    <p className="status-detail">This may take a few minutes</p>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default DTCreate;