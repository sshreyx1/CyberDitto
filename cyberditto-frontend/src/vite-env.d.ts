/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_JWT_SECRET: string;
    readonly VITE_ADMIN_USERNAME: string;
    readonly VITE_ADMIN_PASSWORD: string;
    readonly VITE_SHREYA_USERNAME: string;
    readonly VITE_SHREYA_PASSWORD: string;
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}

declare module '*.svg' {
    import React = require('react');
    export const ReactComponent: React.FC<React.SVGProps<SVGSVGElement>>;
    const src: string;
    export default src;
}