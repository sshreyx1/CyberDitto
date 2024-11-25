declare module 'crypto-js';
declare module 'jwt-decode';

declare global {
    interface Window {
        crypto: Crypto;
    }
}

export {};