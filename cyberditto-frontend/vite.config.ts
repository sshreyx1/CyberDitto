import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';

export default defineConfig({
    plugins: [react()],
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
                secure: false,
                rewrite: (path) => path,
                configure: (proxy) => {
                    proxy.on('error', (err, req, res) => {
                        console.error('Proxy error:', err);
                        if (!res.headersSent) {
                            res.writeHead(503, {
                                'Content-Type': 'application/json',
                                'Connection': 'keep-alive',
                                'Cache-Control': 'no-cache',
                                'Retry-After': '5'
                            });
                            res.end(JSON.stringify({ 
                                error: 'Service temporarily unavailable',
                                retryIn: 5
                            }));
                        }
                    });

                    proxy.on('proxyReq', (proxyReq, req, res) => {
                        proxyReq.removeHeader('Origin');
                        proxyReq.setHeader('Connection', 'keep-alive');
                        proxyReq.setHeader('Cache-Control', 'no-cache');
                    });

                    // Enable keep-alive
                    proxy.on('proxyRes', (proxyRes, req, res) => {
                        proxyRes.headers['connection'] = 'keep-alive';
                        proxyRes.headers['keep-alive'] = 'timeout=5';
                    });
                }
            }
        }
    }
});