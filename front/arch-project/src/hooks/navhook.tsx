import { useNavigate } from 'react-router-dom';

function LoginForm() {
  const navigate = useNavigate();
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    // Login logic here
    
    // After successful login
    navigate('/'); // Navigate to home page
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {/* form fields */}
      <button type="submit">Login</button>
      <button type="button" onClick={() => navigate('/signup')}>
        Need an account?
      </button>
    </form>
  );
}