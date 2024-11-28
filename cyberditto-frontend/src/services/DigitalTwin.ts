interface ProcessedDeployment {
    ID?: string;
    id: string;
    Name?: string;
    name: string;
    CreatedAt?: string;
    createdAt: string | null;
    LastActive?: string;
    lastActive: string | null;
    Status?: string;
    status: string;
    Progress?: number;
    progress: number;
    VagrantID?: string;
    vagrantId: string;
    ResourceUsage?: ResourceUsage;
    resourceUsage: ResourceUsage;
    Message?: string;
    message: string;
    Error?: string; 
    error: string;
    ScanResult?: ScanResult;
    scanResult: ScanResult | null;
}

interface SystemInfo {
    osVersion: string;
    cpuModel: string;
    memory: number;
    diskSpace: number;
    lastBootTime: string;
    osName: string;
}

interface NetworkInterface {
    name: string;
    ipAddress: string;
    macAddress: string;
    linkSpeed: string;
    dnsServers: string[];
}

interface NetworkInfo {
    interfaces: NetworkInterface[];
    openPorts: {
        LocalPort: number;
        RemoteAddress: string;
        State: string;
        OwningProcess: number;
    }[];
    dnsServers: string[];
    services: string[];
}

interface SecurityInfo {
    firewall: {
        enabled: boolean;
        rules: number;
        inboundRules: {
            DisplayName: string;
            Action: string;
            Profile: string;
        }[];
        outboundRules: {
            DisplayName: string;
            Action: string;
            Profile: string;
        }[];
    };
    antivirus: {
        windowsDefender: {
            enabled: boolean;
            realTimeProtection: boolean;
            antispyware: boolean;
            behaviorMonitor: boolean;
            ioavProtection: boolean;
            networkProtection: boolean;
        };
    };
    updates: {
        HotFixID: string;
        InstalledOn: string;
    }[];
    uac: {
        enabled: boolean;
        level: {
            ConsentPromptBehaviorAdmin: number;
        };
    };
}

interface ScanResult {
    id?: string;
    systemInfo: SystemInfo;
    networkInfo: NetworkInfo;
    securityInfo: SecurityInfo;
    createdAt?: string;
}

interface ScanResponse {
    scan_id: string;
    requires_elevation: boolean;
}

interface ScanProgress {
    phase: 'idle' | 'scanning' | 'processing' | 'completed' | 'error';
    stage?: string;
    progress: number;
    message?: string;
    error?: string;
    details?: string;
}

interface ResourceUsage {
    cpu: number;
    memory: number;
    disk: number;
}

interface DeploymentStatus {
    status: 'idle' | 'preparing' | 'initializing' | 'deploying' | 'running' | 'stopped' | 'completed' | 'error' | 'saved';
    progress: number;
    message?: string;
    error?: string;
    vagrantId?: string;
    resourceUsage?: ResourceUsage;
}

interface DeploymentInfo {
    ID?: string;
    id?: string;
    Name?: string;
    name?: string;
    CreatedAt?: string;
    createdAt?: string;
    Status?: DeploymentStatus['status'];
    status?: DeploymentStatus['status'];
    Progress?: number;
    progress?: number;
    Message?: string;
    message?: string;
    Error?: string;
    error?: string;
    VagrantID?: string;
    vagrantId?: string;
    ResourceUsage?: ResourceUsage;
    resourceUsage?: ResourceUsage;
    ScanResult?: ScanResult;
    scanResult?: ScanResult;
    LastActive?: string;
    lastActive?: string;
}

interface DeployResponse {
    deployment_id: string;
}

const processDeployment = (deployment: DeploymentInfo): ProcessedDeployment => {
    const defaultResourceUsage: ResourceUsage = {
        cpu: 0,
        memory: 0,
        disk: 0
    };
 
    const createdAtTime = deployment.CreatedAt || deployment.createdAt || null;
    const lastActiveTime = deployment.LastActive || deployment.lastActive || createdAtTime;
 
    return {
        ID: deployment.ID,
        id: deployment.ID || deployment.id || '',
        Name: deployment.Name,
        name: deployment.Name || deployment.name ||
            `CyberDitto_${deployment.VagrantID || deployment.vagrantId}`,
        CreatedAt: deployment.CreatedAt,
        createdAt: createdAtTime,
        LastActive: deployment.LastActive,
        lastActive: lastActiveTime,
        Status: deployment.Status,
        status: (deployment.Status || deployment.status || 'error').toLowerCase(),
        Progress: deployment.Progress,
        progress: deployment.Progress || deployment.progress || 0,
        VagrantID: deployment.VagrantID,
        vagrantId: deployment.VagrantID || deployment.vagrantId || '',
        ResourceUsage: deployment.ResourceUsage,
        resourceUsage: deployment.ResourceUsage || deployment.resourceUsage || defaultResourceUsage,
        Message: deployment.Message,
        message: deployment.Message || deployment.message || getDefaultMessage(deployment.Status || deployment.status),
        Error: deployment.Error,
        error: deployment.Error || deployment.error || '',
        ScanResult: deployment.ScanResult,
        scanResult: deployment.ScanResult || deployment.scanResult || null
    };
};

const getDefaultMessage = (status: string = ''): string => {
    switch (status.toLowerCase()) {
        case 'running': return 'Virtual environment is running';
        case 'stopped': return 'Virtual environment is stopped';
        case 'error': return 'Error in virtual environment';
        case 'saved': return 'Virtual environment is saved';
        case 'deploying': return 'Deploying virtual environment...';
        case 'preparing': return 'Preparing deployment...';
        case 'initializing': return 'Initializing environment...';
        default: return 'Unknown status';
    }
};

class DigitalTwinApi {
    private readonly baseUrl: string = '/api';
    private readonly maxRetries: number = 2;
    private readonly retryDelay: number = 1000;
    private readonly pollInterval: number = 2000;
    private activePolls: Set<string> = new Set();
    private deploymentStore: Map<string, ProcessedDeployment> = new Map();

    private async handleResponse<T>(response: Response): Promise<T> {
        if (!response.ok) {
            let errorMessage: string;
            try {
                const errorData = await response.json();
                errorMessage = errorData.error || response.statusText;
            } catch {
                errorMessage = response.statusText;
            }

            if (response.status === 404) {
                throw new Error('Resource not found');
            } else if (response.status === 500) {
                return {
                    status: 'error',
                    progress: 0,
                    message: 'Service temporarily unavailable',
                    error: errorMessage
                } as T;
            }
            throw new Error(errorMessage || 'Unknown error occurred');
        }

        try {
            return await response.json() as T;
        } catch (error) {
            console.error('JSON parse error:', error);
            throw new Error('Invalid response format from server');
        }
    }

    private async retryFetch(url: string, options?: RequestInit): Promise<Response> {
        let attempts = 0;
        const maxAttempts = 3;
        const baseDelay = 1000;

        while (attempts < maxAttempts) {
            try {
                const response = await fetch(url, {
                    ...options,
                    headers: {
                        ...options?.headers,
                        'Accept': 'application/json',
                        'Connection': 'keep-alive'
                    }
                });

                if (response.ok || response.status !== 503) {
                    return response;
                }

                attempts++;
                if (attempts < maxAttempts) {
                    await new Promise(resolve => setTimeout(resolve, baseDelay * attempts));
                }
            } catch (error) {
                attempts++;
                if (attempts === maxAttempts) {
                    throw error;
                }
                await new Promise(resolve => setTimeout(resolve, baseDelay * attempts));
            }
        }

        throw new Error('Failed to connect to server after multiple attempts');
    }

    async startScan(): Promise<{ scanId: string }> {
        const response = await this.retryFetch(`${this.baseUrl}/scan`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        });

        const data = await this.handleResponse<ScanResponse>(response);
        return { scanId: data.scan_id };
    }

    async getScanProgress(scanId: string): Promise<ScanProgress> {
        const response = await this.retryFetch(`${this.baseUrl}/scan/${scanId}/status`);
        const data = await this.handleResponse<any>(response);
        
        return {
            phase: data.phase || 'error',
            stage: data.stage,
            progress: data.progress || 0,
            message: data.status || data.message,
            error: data.error
        };
    }

    async getScanResult(scanId: string): Promise<ScanResult> {
        const response = await this.retryFetch(`${this.baseUrl}/scan/${scanId}/result`);
        return await this.handleResponse<ScanResult>(response);
    }

    async pollScan(
        scanId: string,
        onProgress: (progress: ScanProgress) => void,
        onError: (error: Error) => void
    ): Promise<void> {
        let lastProgress = -1;
        let lastStage = '';

        const poll = async () => {
            try {
                const progress = await this.getScanProgress(scanId);
                
                // Only update if progress has changed
                if (progress.progress !== lastProgress || progress.stage !== lastStage) {
                    lastProgress = progress.progress;
                    lastStage = progress.stage || '';
                    onProgress(progress);
                }

                if (progress.phase === 'completed') {
                    try {
                        const result = await this.getScanResult(scanId);
                        onProgress({
                            phase: 'completed',
                            progress: 100,
                            message: 'Scan completed successfully',
                            stage: 'Complete'
                        });
                    } catch (error) {
                        onError(error instanceof Error ? error : new Error('Failed to get scan results'));
                    }
                    return true;
                } else if (progress.phase === 'error' || progress.error) {
                    onError(new Error(progress.error || 'Scan failed'));
                    return true;
                }
                return false;
            } catch (error) {
                onError(error instanceof Error ? error : new Error('Failed to check scan status'));
                return true;
            }
        };

        return new Promise((resolve) => {
            const interval = setInterval(async () => {
                const shouldStop = await poll();
                if (shouldStop) {
                    clearInterval(interval);
                    resolve();
                }
            }, this.pollInterval);
        });
    }

    async getAllDeployments(): Promise<ProcessedDeployment[]> {
        try {
            const response = await this.retryFetch(`${this.baseUrl}/deployments`);
            const data = await this.handleResponse<DeploymentInfo[]>(response);

            if (!Array.isArray(data)) {
                console.warn('Server returned non-array response:', data);
                return [];
            }

            const processedDeployments = data.map(deployment => {
                const processed = processDeployment(deployment);
                this.deploymentStore.set(processed.id || '', processed);
                return processed;
            });

            return processedDeployments;
        } catch (error) {
            console.error('Get all deployments error:', error);
            return Array.from(this.deploymentStore.values());
        }
    }

    async getDeploymentStatus(deploymentId: string): Promise<DeploymentStatus> {
        if (this.activePolls.has(deploymentId)) {
            return {
                status: 'deploying',
                progress: 50,
                message: 'Waiting for status update...',
                resourceUsage: {
                    cpu: 0,
                    memory: 0,
                    disk: 0
                }
            };
        }

        try {
            this.activePolls.add(deploymentId);
            const response = await this.retryFetch(`${this.baseUrl}/deploy/${deploymentId}/status`);
            const data = await this.handleResponse<DeploymentStatus>(response);

            const status = {
                ...data,
                status: data.status || 'error',
                progress: data.progress || 0,
                message: data.message || getDefaultMessage(data.status),
                resourceUsage: {
                    cpu: data.resourceUsage?.cpu ?? 0,
                    memory: data.resourceUsage?.memory ?? 0,
                    disk: data.resourceUsage?.disk ?? 0
                }
            };

            if (this.deploymentStore.has(deploymentId)) {
                const storedDeployment = this.deploymentStore.get(deploymentId);
                if (storedDeployment) {
                    this.deploymentStore.set(deploymentId, {
                        ...storedDeployment,
                        status: status.status,
                        progress: status.progress,
                        message: status.message,
                        resourceUsage: status.resourceUsage
                    });
                }
            }

            return status;
        } catch (error) {
            return {
                status: 'error',
                progress: 0,
                message: error instanceof Error ? error.message : 'Unknown error',
                error: 'Failed to get deployment status',
                resourceUsage: { cpu: 0, memory: 0, disk: 0 }
            };
        } finally {
            this.activePolls.delete(deploymentId);
        }
    }

    async deployDigitalTwin(scanId: string): Promise<string> {
        const response = await this.retryFetch(`${this.baseUrl}/deploy`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ scan_id: scanId })
        });

        const data = await this.handleResponse<DeployResponse>(response);
        return data.deployment_id;
    }

    async startInstance(deploymentId: string): Promise<void> {
        try {
            const response = await this.retryFetch(`${this.baseUrl}/deploy/${deploymentId}/start`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' }
            });
            await this.handleResponse<void>(response);

            if (this.deploymentStore.has(deploymentId)) {
                const deployment = this.deploymentStore.get(deploymentId);
                if (deployment) {
                    this.deploymentStore.set(deploymentId, {
                        ...deployment,
                        status: 'running',
                        progress: 100,
                        message: 'Virtual environment is running'
                    });
                }
            }
        } catch (error) {
            console.error('Start instance error:', error);
            throw new Error('Failed to start instance');
        }
    }

    async stopInstance(deploymentId: string): Promise<void> {
        try {
            const response = await this.retryFetch(`${this.baseUrl}/deploy/${deploymentId}/stop`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' }
            });
            await this.handleResponse<void>(response);

            if (this.deploymentStore.has(deploymentId)) {
                const deployment = this.deploymentStore.get(deploymentId);
                if (deployment) {
                    this.deploymentStore.set(deploymentId, {
                        ...deployment,
                        status: 'stopped',
                        progress: 0,
                        message: 'Virtual environment is stopped'
                    });
                }
            }
        } catch (error) {
            console.error('Stop instance error:', error);
            throw new Error('Failed to stop instance');
        }
    }

    async deleteInstance(deploymentId: string): Promise<void> {
        try {
            const response = await this.retryFetch(`${this.baseUrl}/deploy/${deploymentId}`, {
                method: 'DELETE'
            });
            await this.handleResponse<void>(response);
            this.deploymentStore.delete(deploymentId);
        } catch (error) {
            console.error('Delete instance error:', error);
            throw new Error('Failed to delete instance');
        }
    }

    async pollDeploymentStatus(
        deploymentId: string,
        onStatus: (status: DeploymentStatus) => void,
        onError: (error: Error) => void,
        maxAttempts = 120
    ): Promise<void> {
        let attempts = 0;
        let lastStatus = '';
        let lastProgress = -1;

        const poll = async () => {
            if (attempts >= maxAttempts) {
                onError(new Error('Deployment timed out'));
                return true;
            }

            try {
                const status = await this.getDeploymentStatus(deploymentId);
                
                // Only trigger callback if status has changed
                if (status.status !== lastStatus || status.progress !== lastProgress) {
                    lastStatus = status.status;
                    lastProgress = status.progress;
                    onStatus(status);
                }

                if (['running', 'error', 'stopped'].includes(status.status)) {
                    return true;
                }

                attempts++;
                return false;
            } catch (error) {
                onError(error instanceof Error ? error : new Error('Polling error'));
                return true;
            }
        };

        return new Promise((resolve) => {
            const interval = setInterval(async () => {
                const shouldStop = await poll();
                if (shouldStop) {
                    clearInterval(interval);
                    resolve();
                }
            }, this.pollInterval);
        });
    }

    formatResourceUsage(usage: number): string {
        return `${Math.round(Math.max(0, Math.min(100, usage)))}%`;
    }

    formatMemory(bytes: number): string {
        const gb = Math.max(0, bytes) / (1024 * 1024 * 1024);
        return `${Math.round(gb * 100) / 100} GB`;
    }

    formatDiskSpace(bytes: number): string {
        const gb = Math.max(0, bytes) / (1024 * 1024 * 1024);
        return `${Math.round(gb * 100) / 100} GB`;
    }

    isDeploymentActive(status: DeploymentStatus['status']): boolean {
        return ['running', 'deploying', 'preparing', 'initializing'].includes(status.toLowerCase());
    }

    canStartInstance(status: DeploymentStatus['status']): boolean {
        return ['stopped', 'saved', 'error'].includes(status.toLowerCase());
    }

    canStopInstance(status: DeploymentStatus['status']): boolean {
        return ['running'].includes(status.toLowerCase());
    }

    canDeleteInstance(status: DeploymentStatus['status']): boolean {
        return ['stopped', 'error', 'saved'].includes(status.toLowerCase());
    }

    isDeploymentFailed(status: DeploymentStatus['status']): boolean {
        return status.toLowerCase() === 'error';
    }

    isDeploymentCompleted(status: DeploymentStatus['status']): boolean {
        return ['running', 'stopped', 'saved'].includes(status.toLowerCase());
    }
}

export const digitalTwinApi = new DigitalTwinApi();

export type {
    ProcessedDeployment,
    SystemInfo,
    NetworkInterface,
    NetworkInfo,
    SecurityInfo,
    ScanResult,
    ScanProgress,
    ResourceUsage,
    DeploymentStatus,
    DeploymentInfo,
    ScanResponse,
    DeployResponse
};