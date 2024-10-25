import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './Login.css';

const Login: React.FC = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleLogin = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        // Simple authentication check
        if (username === 'admin' && password === 'password') {
            localStorage.setItem('authToken', 'some-token-value');
            localStorage.setItem('username', username);
            navigate('/dashboard');
        } else {
            setError('Invalid username or password.');
        }
    };

    return (
        <div className="auth-login-container">
            <div className="auth-login-box">
                <div className="auth-login-logo-section">
                    {/* Add your logo here if needed */}
                </div>
                <form onSubmit={handleLogin}>
                    <div className="auth-login-form-group">
                        <input 
                            type="text" 
                            className="auth-login-input" 
                            placeholder="Username" 
                            value={username}
                            onChange={(e) => setUsername(e.target.value)} 
                            required 
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
                        />
                    </div>
                    {error && <div className="auth-login-error">{error}</div>}
                    <button type="submit" className="auth-login-button">Login</button>
                </form>
                <div className="auth-login-signup-text">
                    Don't have an account? <a href="/signup">Sign up</a>
                </div>
            </div>
        </div>
    );
};

export default Login;