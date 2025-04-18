import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom'; // Add this import
import { useAuth } from '../context/authContext';

const LoginForm: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  
  const { login, loginWithGoogle } = useAuth();
  const navigate = useNavigate(); // Add this hook
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);
    
    try {
      await login(email, password);
      // Add navigation after successful login
      navigate('/'); // Navigate to home page
    } catch (err: any) {
      setError(err.message || 'Failed to login');
    } finally {
      setIsLoading(false);
    }
  };
  
  const handleGoogleLogin = async () => {
    try {
      await loginWithGoogle();
      // No navigate here because Google OAuth will redirect
    } catch (err: any) {
      setError(err.message || 'Failed to login with Google');
    }
  };
  
  return (
    <div className="login-form">
      <h2>Login</h2>
      
      {error && <div className="error-message">{error}</div>}
      
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        
        <div className="form-actions">
          <button type="submit" disabled={isLoading}>
            {isLoading ? 'Logging in...' : 'Login'}
          </button>
        </div>
      </form>
      
      <div className="social-login">
        <button onClick={handleGoogleLogin} className="google-btn">
          Login with Google
        </button>
      </div>
      
      <div className="form-footer">
        <p>
          Don't have an account? <Link to="/signup">Sign Up</Link>
        </p>
      </div>
    </div>
  );
};

export default LoginForm;