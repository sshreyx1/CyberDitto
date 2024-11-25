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
import type { ScanStatus, DeploymentStatus, ScanResult } from '../services/DigitalTwin';
import Sidebar from './Sidebar';
import './DT-Create.css';

interface ScanState {
    scanId: string;
    status: ScanStatus['phase'];
    progress: number;
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
    const POLLING_INTERVAL = 2000;

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
        setScanState(prev => ({ ...prev, status: 'scanning', progress: 0 }));
        try {
            const response = await fetch('/api/scan', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' }
            });
            const data = await response.json();
            setScanState(prev => ({ ...prev, scanId: data.scan_id }));
            pollScanStatus(data.scan_id);
        } catch (error) {
            setScanState(prev => ({
                ...prev,
                status: 'error',
                error: error instanceof Error ? error.message : 'Failed to start scan'
            }));
        }
    };

    const pollScanStatus = React.useCallback(async (scanId: string) => {
        const checkStatus = async () => {
            try {
                const response = await fetch(`/api/scan/${scanId}/status`);
                const status = await response.json();

                setScanState(prev => ({
                    ...prev,
                    status: status.phase,
                    progress: status.progress,
                    error: status.error || ''
                }));

                if (status.phase === 'completed') {
                    const resultResponse = await fetch(`/api/scan/${scanId}/result`);
                    const result = await resultResponse.json();
                    setScanState(prev => ({ ...prev, result }));
                    return true;
                } else if (status.phase === 'error') {
                    return true;
                }
                return false;
            } catch (error) {
                console.error('Scan status check error:', error);
                return true;
            }
        };

        const interval = setInterval(async () => {
            const shouldStop = await checkStatus();
            if (shouldStop) {
                clearInterval(interval);
            }
        }, POLLING_INTERVAL);

        return () => clearInterval(interval);
    }, []);

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
            const response = await fetch('/api/deploy', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ scan_id: scanState.scanId })
            });
            const deployData = await response.json();
            const deploymentId = deployData.deployment_id;

            setDeployState(prev => ({
                ...prev,
                deploymentId,
                message: 'Initializing virtual environment...'
            }));

            // Poll deployment status
            const checkStatus = async () => {
                try {
                    const statusResponse = await fetch(`/api/deploy/${deploymentId}/status`);
                    const status = await statusResponse.json();

                    setDeployState(prev => ({
                        ...prev,
                        status: status.status,
                        progress: status.progress,
                        error: status.error || '',
                        vagrantId: status.vagrantId || '',
                        message: status.message || '',
                        success: status.status === 'running'
                    }));

                    return ['running', 'error', 'stopped'].includes(status.status);
                } catch (error) {
                    setDeployState(prev => ({
                        ...prev,
                        status: 'error',
                        error: error instanceof Error ? error.message : 'Failed to check deployment status',
                        success: false
                    }));
                    return true;
                }
            };

            const interval = setInterval(async () => {
                const shouldStop = await checkStatus();
                if (shouldStop) {
                    clearInterval(interval);
                }
            }, POLLING_INTERVAL);

            return () => clearInterval(interval);

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
                                        <p>
                                            Memory:{' '}
                                            {Math.round(
                                                scanState.result.systemInfo.memory /
                                                (1024 * 1024 * 1024)
                                            )}{' '}
                                            GB
                                        </p>
                                    </div>

                                    <div className="result-section">
                                        <h4>Network Info</h4>
                                        <p>
                                            Interfaces:{' '}
                                            {scanState.result.networkInfo.interfaces.length}
                                        </p>
                                        <p>
                                            Open Ports:{' '}
                                            {scanState.result.networkInfo.openPorts.length}
                                        </p>
                                    </div>

                                    <button
                                        className="dt-button"
                                        onClick={deployDigitalTwin}
                                        disabled={deployState.status === 'deploying' || deployState.success || deployState.status !== 'idle'} // Modified condition
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