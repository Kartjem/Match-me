import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./Auth.css";

function Register() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const handleRegister = async (e) => {
    e.preventDefault();
    try {
      const res = await fetch("http://localhost:8080/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (res.ok) {
        alert("Registration successful. Please log in.");
        navigate("/login");
      } else {
        const errorText = await res.text();
        alert("Registration error: " + errorText);
      }
    } catch (err) {
      console.error(err);
      alert("Error registering. Check console.");
    }
  };

  return (
    <div className="auth-container">
      <h2>Create Your Account</h2>
      <form onSubmit={handleRegister} className="auth-form">
        <input
          type="email"
          required
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <input
          type="password"
          required
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button type="submit" className="primary-btn">Sign Up</button>
      </form>
    </div>
  );
}

export default Register;
