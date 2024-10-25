import React, { useState, useEffect } from 'react';
import './Sidebar.css';
import { 
  FaChartPie, 
  FaDesktop, 
  FaClipboardCheck, 
  FaUserShield, 
  FaSignOutAlt,
  FaChevronDown,
  FaPlus,
  FaEye
} from 'react-icons/fa';
import { Link, useNavigate, useLocation } from 'react-router-dom';

const Sidebar: React.FC = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const [expandedMenu, setExpandedMenu] = useState<string | null>(null);

    // Automatically expand menu based on current path
    useEffect(() => {
        if (location.pathname.includes('/digitaltwin')) {
            setExpandedMenu('digitaltwin');
        } else if (location.pathname.includes('/benchmark')) {
            setExpandedMenu('benchmark');
        } else if (location.pathname.includes('/emulation')) {
            setExpandedMenu('emulation');
        }
    }, [location.pathname]);

    const handleLogout = () => {
        localStorage.removeItem('authToken');
        localStorage.removeItem('username');
        navigate('/login');
    };

    const isActive = (path: string) => {
        return location.pathname === path ? 'app-sidebar-link active' : 'app-sidebar-link';
    };

    const toggleSubMenu = (menu: string, e: React.MouseEvent) => {
        // Only toggle if clicking the chevron
        if ((e.target as HTMLElement).closest('.app-sidebar-chevron')) {
            e.preventDefault();
            setExpandedMenu(expandedMenu === menu ? null : menu);
        }
    };

    const isSubActive = (path: string) => {
        return location.pathname === path ? 'app-sidebar-sublink active' : 'app-sidebar-sublink';
    };

    return (
        <div className="app-sidebar">
            <ul className="app-sidebar-menu">
                <li className="app-sidebar-item">
                    <Link to="/dashboard" className={isActive('/dashboard')}>
                        <FaChartPie className="app-sidebar-icon" />
                        <span>Dashboard</span>
                    </Link>
                </li>

                {/* Digital Twin Section */}
                <li className="app-sidebar-item">
                    <Link 
                        to="/digitaltwin" 
                        className={`app-sidebar-link ${location.pathname.includes('/digitaltwin') ? 'active' : ''}`}
                        onClick={(e) => toggleSubMenu('digitaltwin', e)}
                    >
                        <div className="app-sidebar-link-content">
                            <FaDesktop className="app-sidebar-icon" />
                            <span>Digital Twin</span>
                        </div>
                        <FaChevronDown className={`app-sidebar-chevron ${expandedMenu === 'digitaltwin' ? 'rotated' : ''}`} />
                    </Link>
                    {expandedMenu === 'digitaltwin' && (
                        <ul className="app-sidebar-submenu">
                            <li>
                                <Link to="/digitaltwin/create" className={isSubActive('/digitaltwin/create')}>
                                    <FaPlus className="app-sidebar-subicon" />
                                    <span>Create</span>
                                </Link>
                            </li>
                            <li>
                                <Link to="/digitaltwin/view" className={isSubActive('/digitaltwin/view')}>
                                    <FaEye className="app-sidebar-subicon" />
                                    <span>View</span>
                                </Link>
                            </li>
                        </ul>
                    )}
                </li>

                {/* Benchmark Section */}
                <li className="app-sidebar-item">
                    <Link 
                        to="/benchmark"
                        className={`app-sidebar-link ${location.pathname.includes('/benchmark') ? 'active' : ''}`}
                        onClick={(e) => toggleSubMenu('benchmark', e)}
                    >
                        <div className="app-sidebar-link-content">
                            <FaClipboardCheck className="app-sidebar-icon" />
                            <span>CIS Benchmark</span>
                        </div>
                        <FaChevronDown className={`app-sidebar-chevron ${expandedMenu === 'benchmark' ? 'rotated' : ''}`} />
                    </Link>
                    {expandedMenu === 'benchmark' && (
                        <ul className="app-sidebar-submenu">
                            <li>
                                <Link to="/benchmark/run" className={isSubActive('/benchmark/run')}>
                                    <FaPlus className="app-sidebar-subicon" />
                                    <span>Run</span>
                                </Link>
                            </li>
                            <li>
                                <Link to="/benchmark/view" className={isSubActive('/benchmark/view')}>
                                    <FaEye className="app-sidebar-subicon" />
                                    <span>View</span>
                                </Link>
                            </li>
                        </ul>
                    )}
                </li>

                {/* Emulation Section */}
                <li className="app-sidebar-item">
                    <Link 
                        to="/emulation"
                        className={`app-sidebar-link ${location.pathname.includes('/emulation') ? 'active' : ''}`}
                        onClick={(e) => toggleSubMenu('emulation', e)}
                    >
                        <div className="app-sidebar-link-content">
                            <FaUserShield className="app-sidebar-icon" />
                            <span>Adversary Emulation</span>
                        </div>
                        <FaChevronDown className={`app-sidebar-chevron ${expandedMenu === 'emulation' ? 'rotated' : ''}`} />
                    </Link>
                    {expandedMenu === 'emulation' && (
                        <ul className="app-sidebar-submenu">
                            <li>
                                <Link to="/emulation/run" className={isSubActive('/emulation/run')}>
                                    <FaPlus className="app-sidebar-subicon" />
                                    <span>Run</span>
                                </Link>
                            </li>
                            <li>
                                <Link to="/emulation/view" className={isSubActive('/emulation/view')}>
                                    <FaEye className="app-sidebar-subicon" />
                                    <span>View</span>
                                </Link>
                            </li>
                        </ul>
                    )}
                </li>
            </ul>
            <button className="app-sidebar-logout" onClick={handleLogout}>
                <FaSignOutAlt className="app-sidebar-icon" />
                <span>Log Out</span>
            </button>
        </div>
    );
};

export default Sidebar;