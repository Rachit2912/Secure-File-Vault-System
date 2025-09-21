// src/pages/SignupPage.tsx
import React from "react";
import { SignupForm } from "../../components/forms/SignupForm";
import { Link } from "react-router-dom";

const SignupPage: React.FC = () => {
  return (
    <div style={{ padding: "2rem" }}>
      <h1>Sign Up</h1>
      {/* signup form component : */}
      <SignupForm />
      {/* link for login page :  */}
      <p>
        Already have an account? <Link to="/login">Log in</Link>
      </p>
    </div>
  );
};

export default SignupPage;
