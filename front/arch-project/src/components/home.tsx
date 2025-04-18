import React from 'react';
import { useAuth } from '../context/authContext';
import '../styles/home.css';

const Home: React.FC = () => {
  const { user, isAuthenticated } = useAuth();

  return (
    <div className="home-page">
      <div className="home-container">
        <h1>Welcome to Blaban</h1>
        
        {isAuthenticated ? (
          <div className="authenticated-content">
            <p>Hello, {user?.name}! You are successfully logged in.</p>
            <div className="app-description">
              <h2>What is Blaban?</h2>
              <p>Blaban is your new favorite application for managing tasks and projects.</p>
              <p>Start using it today and see how it improves your productivity!</p>
            </div>
          </div>
        ) : (
          <div className="unauthenticated-content">
            <p>Please log in to access all features.</p>
            <div className="auth-buttons">
              <a href="/login" className="login-link">Login</a>
              <a href="/signup" className="signup-link">Sign Up</a>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Home;