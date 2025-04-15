// src/components/NavBar.tsx
import { Link } from 'react-router-dom';
import { useAuth } from '../context/authContext';

export default function NavBar() {
  const { isAuthenticated, logout } = useAuth();
  
  return (
    <nav className="navbar">
      <div className="logo">
        <Link to="/">Blaban</Link>
      </div>
      
      <div className="nav-links">
        {isAuthenticated ? (
          <>
            <Link to="/profile">My Profile</Link>
            <button onClick={logout}>Logout</button>
          </>
        ) : (
          <>
            <Link to="/login">Login</Link>
            <Link to="/signup">Sign Up</Link>
          </>
        )}
      </div>
    </nav>
  );
}