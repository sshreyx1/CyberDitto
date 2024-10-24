import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './Login.css'; // Adjust path as needed

const Login: React.FC = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();

    const handleLogin = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        // Simple authentication check
        if (username === 'admin' && password === 'password') {
            // Save auth token to local storage
            localStorage.setItem('authToken', 'some-token-value');
            localStorage.setItem('username', username);
            navigate('/dashboard'); // Redirect to the dashboard
        } else {
            setError('Invalid username or password.'); // Show error message
        }
    };

    return (
        <div className="login-container">
            <div className="login-box">
                <div className="login-logo-section">
                </div>
                <form onSubmit={handleLogin}>
                    <div className="login-form-group">
                        <input 
                            type="text" 
                            className="login-input" 
                            placeholder="Username" 
                            value={username}
                            onChange={(e) => setUsername(e.target.value)} 
                            required 
                        />
                    </div>
                    <div className="login-form-group">
                        <input 
                            type="password" 
                            className="login-input" 
                            placeholder="Password" 
                            value={password}
                            onChange={(e) => setPassword(e.target.value)} 
                            required 
                        />
                    </div>
                    {error && <div className="login-error">{error}</div>}
                    <button type="submit" className="login-button">Login</button>
                </form>
                <div className="login-signup-text">
                    Don't have an account? <a href="/signup">Sign up</a>
                </div>
            </div>
        </div>
    );
};

export default Login;