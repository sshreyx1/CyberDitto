export const AUTH_CONFIG = {
    JWT_SECRET: import.meta.env.VITE_JWT_SECRET || 'fallback-secret-key',
    USERS: [
        {
            id: '1',
            username: import.meta.env.VITE_ADMIN_USERNAME || 'admin',
            password: import.meta.env.VITE_ADMIN_PASSWORD || 'CyberDit0o',
            role: 'user'
        },
        {
            id: '2',
            username: import.meta.env.VITE_SHREYA_USERNAME || 'shreya',
            password: import.meta.env.VITE_SHREYA_PASSWORD || 'Gpa4.oplz',
            role: 'user'
        }
    ]
};

// Type declarations for environment variables
declare global {
    interface ImportMetaEnv {
        readonly VITE_JWT_SECRET: string;
        readonly VITE_ADMIN_USERNAME: string;
        readonly VITE_ADMIN_PASSWORD: string;
        readonly VITE_SHREYA_USERNAME: string;
        readonly VITE_SHREYA_PASSWORD: string;
    }
}