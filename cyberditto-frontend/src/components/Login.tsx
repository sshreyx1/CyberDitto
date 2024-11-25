import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { authService } from '../services/AuthService';
import './Login.css';

const Login: React.FC = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();

    useEffect(() => {
        if (authService.isAuthenticated()) {
            const from = location.state?.from?.pathname || '/dashboard';
            navigate(from, { replace: true });
        }
    }, [navigate, location]);

    const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            const success = authService.login(username, password);
            if (success) {
                const from = location.state?.from?.pathname || '/dashboard';
                navigate(from, { replace: true });
            } else {
                setError('Invalid username or password.');
            }
        } catch (err) {
            setError('An error occurred during login. Please try again.');
            console.error('Login error:', err);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="auth-login-container">
            <div className="auth-login-box">
                <div className="auth-login-logo-section">
                    <h1 className="auth-login-title">Welcome Back</h1>
                </div>
                <form onSubmit={handleLogin} noValidate>
                    <div className="auth-login-form-group">
                        <input 
                            type="text" 
                            className="auth-login-input" 
                            placeholder="Username" 
                            value={username}
                            onChange={(e) => setUsername(e.target.value)} 
                            required 
                            autoComplete="username"
                            disabled={isLoading}
                            aria-label="Username"
                        />
                    </div>
                    <div className="auth-login-form-group">
                        <input 
                            type="password" 
                            className="auth-login-input" 
                            placeholder="Password" 
                            value={password}
                            onChange={(e) => setPassword(e.target.value)} 
                            required 
                            autoComplete="current-password"
                            disabled={isLoading}
                            aria-label="Password"
                        />
                    </div>
                    {error && (
                        <div className="auth-login-error" role="alert">
                            {error}
                        </div>
                    )}
                    <button 
                        type="submit" 
                        className="auth-login-button"
                        disabled={isLoading}
                    >
                        {isLoading ? 'Logging in...' : 'Login'}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default Login;