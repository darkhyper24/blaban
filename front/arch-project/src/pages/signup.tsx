import React from 'react';
import SignupForm from '../components/signupForm';

const SignupPage: React.FC = () => {
  return (
    <div className="signup-page">
      <div className="signup-container">
        <SignupForm />
      </div>
    </div>
  );
};

export default SignupPage;