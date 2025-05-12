import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/authContext';
import '../styles/Navbar.css';

const NavBar: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <Link to="/" className="navbar-logo-link">
          <img src="/src/assets/blabanlogo.png" alt="Blaban Logo" className="navbar-logo" />
        </Link>
      </div>
      <div className="navbar-menu">
        {isAuthenticated ? (
          <>
            <span className="welcome-text">Welcome, {user?.name}</span>
            <button onClick={handleLogout} className="auth-button logout-button">
              Logout
            </button>
          </>
        ) : (
          <Link to="/login" className="auth-button login-button">
            Login
          </Link>
        )}
      </div>
    </nav>
  );
};

export default NavBar;