import React from 'react';
import './Sidebar.css';
import { FaChartPie, FaDesktop, FaClipboardCheck, FaUserShield, FaSignOutAlt } from 'react-icons/fa';
import { Link, useNavigate } from 'react-router-dom';

const Sidebar: React.FC = () => {
    const navigate = useNavigate();

    const handleLogout = () => {
        // Remove the auth token and username from local storage
        localStorage.removeItem('authToken');
        localStorage.removeItem('username');
        
        // Redirect to login page
        navigate('/login');
    };

    return (
        <div className="sidebar-container">
            <ul className="sidebar-menu">
                <li className="sidebar-item">
                    <Link to="/dashboard" className="sidebar-link">
                        <FaChartPie className="sidebar-icon" /> Dashboard
                    </Link>
                </li>
                <li className="sidebar-item">
                    <Link to="/digitaltwin" className="sidebar-link">
                        <FaDesktop className="sidebar-icon" /> Digital Twin
                    </Link>
                </li>
                <li className="sidebar-item">
                    <Link to="/benchmark" className="sidebar-link">
                        <FaClipboardCheck className="sidebar-icon" /> CIS Benchmark
                    </Link>
                </li>
                <li className="sidebar-item">
                    <Link to="/emulation" className="sidebar-link">
                        <FaUserShield className="sidebar-icon" /> Adversary Emulation
                    </Link>
                </li>
            </ul>
            <button className="logout-button" onClick={handleLogout}>
                <FaSignOutAlt className="sidebar-icon" /> Log Out
            </button>
        </div>
    );
};

export default Sidebar;
