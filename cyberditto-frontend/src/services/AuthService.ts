import { jwtDecode } from "jwt-decode";
import * as CryptoJS from 'crypto-js';
import { AUTH_CONFIG } from '../config/config';

export interface User {
    id: string;
    username: string;
    role: string;
}

interface DecodedToken {
    user: User;
    exp: number;
}

class AuthService {
    private static instance: AuthService;
    private tokenRefreshTimer: ReturnType<typeof setInterval> | null = null;

    private constructor() {
        const token = this.getToken();
        if (token) {
            this.setupTokenRefresh();
        }
    }

    public static getInstance(): AuthService {
        if (!AuthService.instance) {
            AuthService.instance = new AuthService();
        }
        return AuthService.instance;
    }

    login(username: string, password: string): boolean {
        const user = AUTH_CONFIG.USERS.find(
            u => u.username === username && u.password === password
        );

        if (user) {
            const userData: User = {
                id: user.id,
                username: user.username,
                role: user.role
            };

            const token = this.generateToken(userData);
            localStorage.setItem('authToken', token);
            localStorage.setItem('user', JSON.stringify(userData));
            
            this.setupTokenRefresh();
            return true;
        }
        return false;
    }

    logout(): void {
        localStorage.removeItem('authToken');
        localStorage.removeItem('user');
        if (this.tokenRefreshTimer) {
            clearInterval(this.tokenRefreshTimer);
            this.tokenRefreshTimer = null;
        }
    }

    isAuthenticated(): boolean {
        const token = this.getToken();
        if (!token) return false;

        try {
            const decoded = this.verifyToken(token);
            return decoded !== null && decoded.exp > Date.now() / 1000;
        } catch (error) {
            this.logout();
            return false;
        }
    }

    getToken(): string | null {
        return localStorage.getItem('authToken');
    }

    getCurrentUser(): User | null {
        const userStr = localStorage.getItem('user');
        if (!userStr) return null;
        
        try {
            return JSON.parse(userStr) as User;
        } catch {
            return null;
        }
    }

    private generateToken(user: User): string {
        const header = {
            alg: 'HS256',
            typ: 'JWT'
        };

        const payload = {
            user,
            exp: Math.floor(Date.now() / 1000) + (12 * 60 * 60), // 12 hours
            iat: Math.floor(Date.now() / 1000)
        };

        const headerBase64 = btoa(JSON.stringify(header));
        const payloadBase64 = btoa(JSON.stringify(payload));

        const signature = CryptoJS.HmacSHA256(
            `${headerBase64}.${payloadBase64}`,
            AUTH_CONFIG.JWT_SECRET
        ).toString(CryptoJS.enc.Base64);

        return `${headerBase64}.${payloadBase64}.${signature}`;
    }

    private verifyToken(token: string): DecodedToken | null {
        try {
            const decoded = jwtDecode<DecodedToken>(token);
            const parts = token.split('.');
            if (parts.length !== 3) return null;

            const signature = CryptoJS.HmacSHA256(
                `${parts[0]}.${parts[1]}`,
                AUTH_CONFIG.JWT_SECRET
            ).toString(CryptoJS.enc.Base64);

            if (parts[2] !== signature) return null;

            return decoded;
        } catch (error) {
            return null;
        }
    }

    private setupTokenRefresh(): void {
        if (this.tokenRefreshTimer) {
            clearInterval(this.tokenRefreshTimer);
        }

        this.tokenRefreshTimer = setInterval(() => {
            const user = this.getCurrentUser();
            if (user && this.isAuthenticated()) {
                const newToken = this.generateToken(user);
                localStorage.setItem('authToken', newToken);
            } else {
                this.logout();
            }
        }, 11 * 60 * 60 * 1000);
    }
}

export const authService = AuthService.getInstance();