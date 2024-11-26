// src/services/emulation.ts

// Type definitions
export type EmulationMode = "tactic" | "technique" | "custom";

export interface EmulationRequest {
    deployment_id: string;
    mode: EmulationMode;
    target: string;
    test_numbers?: string;
}

export interface EmulationStatus {
    phase: string;
    progress: number;
    message: string;
    error?: string;
    start_time: string;
    completed_at?: string;
}

export interface TestSummary {
    total_tests: number;
    passed_tests: number;
    failed_tests: number;
    success_rate: number;
}

export interface TestResult {
    timestamp: string;
    status: string;
    tactic: string;
    technique: string;
    technique_name: string;
    test_number: string;
    test_name: string;
    test_guid: string;
    executor_name: string;
    error_message?: string;
    log_file?: string;
}

export interface EmulationResult {
    id: string;
    deploy_id: string;
    status: string;
    summary: TestSummary;
    results: TestResult[];
    created_at: string;
    completed_at: string;
    error?: string;
}

// API Implementation
export class EmulationApi {
    private readonly baseUrl: string = '/api';
    private readonly maxRetries: number = 3;
    private readonly retryDelay: number = 1000;
    private readonly pollInterval: number = 5000;

    constructor() {
        console.log('[EmulationApi] Initialized');
    }

    private logDebug(message: string, context?: any) {
        const timestamp = new Date().toISOString();
        console.log(`[EmulationApi][${timestamp}] ${message}`, context || '');
    }

    private logError(message: string, error?: any) {
        const timestamp = new Date().toISOString();
        console.error(`[EmulationApi][${timestamp}] ERROR: ${message}`, error || '');
    }

    private async retryFetch(url: string, options?: RequestInit): Promise<Response> {
        let lastError: Error | null = null;
        
        for (let attempt = 1; attempt <= this.maxRetries; attempt++) {
            try {
                this.logDebug(`Attempt ${attempt}/${this.maxRetries} - ${options?.method || 'GET'} ${url}`);
                
                const response = await fetch(url, {
                    ...options,
                    headers: {
                        'Content-Type': 'application/json',
                        ...options?.headers,
                    },
                });

                if (!response.ok) {
                    const errorData = await response.json().catch(() => ({ error: response.statusText }));
                    throw new Error(errorData.error || 'Request failed');
                }

                this.logDebug(`Request successful - ${response.status} ${response.statusText}`);
                return response;
            } catch (error) {
                lastError = error instanceof Error ? error : new Error('Unknown error');
                this.logError(`Request failed - Attempt ${attempt}`, lastError);
                
                if (attempt < this.maxRetries) {
                    const delay = this.retryDelay * attempt;
                    this.logDebug(`Retrying in ${delay}ms`);
                    await new Promise(resolve => setTimeout(resolve, delay));
                }
            }
        }

        throw lastError || new Error('Failed after multiple retries');
    }

    private async handleResponse<T>(response: Response): Promise<T> {
        try {
            const data = await response.json() as T;
            this.logDebug('Response parsed successfully', data);
            return data;
        } catch (error) {
            this.logError('Failed to parse response', error);
            throw new Error('Failed to parse response');
        }
    }

    async startEmulation(request: EmulationRequest): Promise<{ execution_id: string }> {
        try {
            this.logDebug('Starting emulation', request);
            
            const response = await this.retryFetch(`${this.baseUrl}/emulation`, {
                method: 'POST',
                body: JSON.stringify(request),
            });
            
            const result = await this.handleResponse<{ execution_id: string }>(response);
            this.logDebug('Emulation started successfully', result);
            return result;
        } catch (error) {
            this.logError('Start emulation error', error);
            throw new Error('Failed to start emulation');
        }
    }

    async getEmulationStatus(executionId: string): Promise<EmulationStatus> {
        try {
            this.logDebug(`Getting status for execution ${executionId}`);
            
            const response = await this.retryFetch(`${this.baseUrl}/emulation/${executionId}/status`);
            const status = await this.handleResponse<EmulationStatus>(response);
            
            this.logDebug(`Status retrieved for ${executionId}`, status);
            return status;
        } catch (error) {
            this.logError(`Get emulation status error for ${executionId}`, error);
            throw new Error('Failed to get emulation status');
        }
    }

    async getEmulationResult(executionId: string): Promise<EmulationResult> {
        try {
            this.logDebug(`Getting results for execution ${executionId}`);
            
            const response = await this.retryFetch(`${this.baseUrl}/emulation/${executionId}/result`);
            const result = await this.handleResponse<EmulationResult>(response);
            
            this.logDebug(`Results retrieved for ${executionId}`, result);
            return result;
        } catch (error) {
            this.logError(`Get emulation result error for ${executionId}`, error);
            throw new Error('Failed to get emulation result');
        }
    }

    async getAllEmulations(): Promise<EmulationResult[]> {
        try {
            this.logDebug('Getting all emulations');
            
            const response = await this.retryFetch(`${this.baseUrl}/emulations`);
            const results = await this.handleResponse<EmulationResult[]>(response);
            
            this.logDebug(`Retrieved ${results.length} emulations`);
            return results;
        } catch (error) {
            this.logError('Get all emulations error', error);
            throw new Error('Failed to get emulations');
        }
    }

    async pollEmulationStatus(
        executionId: string,
        onStatus: (status: EmulationStatus) => void,
        onError: (error: Error) => void,
        maxAttempts = 120
    ): Promise<void> {
        let attempts = 0;
        let interval = this.pollInterval;

        const poll = async () => {
            if (attempts >= maxAttempts) {
                const timeoutError = new Error('Emulation timed out');
                this.logError(`Polling timeout for ${executionId} after ${attempts} attempts`);
                onError(timeoutError);
                return;
            }

            try {
                this.logDebug(`Polling status for ${executionId} - Attempt ${attempts + 1}`);
                
                const status = await this.getEmulationStatus(executionId);
                onStatus(status);

                if (['completed', 'error'].includes(status.phase)) {
                    this.logDebug(`Polling completed for ${executionId} - Final status: ${status.phase}`);
                    return;
                }

                attempts++;
                interval = Math.min(interval * 1.5, 10000);
                this.logDebug(`Scheduling next poll in ${interval}ms`);
                setTimeout(poll, interval);
            } catch (error) {
                this.logError(`Polling error for ${executionId}`, error);
                onError(error instanceof Error ? error : new Error('Polling error'));
            }
        };

        await poll();
    }

    async cancelEmulation(executionId: string): Promise<void> {
        try {
            this.logDebug(`Cancelling emulation ${executionId}`);
            
            await this.retryFetch(`${this.baseUrl}/emulation/${executionId}/cancel`, {
                method: 'POST',
            });
            
            this.logDebug(`Emulation ${executionId} cancelled successfully`);
        } catch (error) {
            this.logError(`Cancel emulation error for ${executionId}`, error);
            throw new Error('Failed to cancel emulation');
        }
    }

    // Helper methods
    formatSuccessRate(rate: number): string {
        return `${Math.round(rate)}%`;
    }

    getPhaseMessage(phase: string): string {
        switch (phase) {
            case 'preparing': return 'Preparing attack simulation environment...';
            case 'copying': return 'Copying attack scripts to VM...';
            case 'executing': return 'Executing attack simulation...';
            case 'collecting': return 'Collecting test results...';
            case 'processing': return 'Processing test results...';
            case 'completed': return 'Attack simulation completed';
            case 'error': return 'Error during attack simulation';
            case 'cancelled': return 'Attack simulation cancelled';
            default: return 'Unknown phase';
        }
    }
}

export const emulationApi = new EmulationApi();