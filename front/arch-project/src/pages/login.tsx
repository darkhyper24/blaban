import React from 'react';
import LoginForm from '../components/loginForm';
import '../styles/auth.css';

const LoginPage: React.FC = () => {
  return (
    <div className="login-page">
      <div className="login-container">
        <LoginForm />
      </div>
    </div>
  );
};

export default LoginPage;