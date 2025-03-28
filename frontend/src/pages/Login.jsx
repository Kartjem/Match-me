import React, { useState, useContext } from "react";
import { useNavigate } from "react-router-dom";
import { AuthContext } from "../context/AuthContext";
import "./Auth.css";

function Login() {
  const { login } = useContext(AuthContext);
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const handleLogin = async (e) => {
    e.preventDefault();
    try {
      const res = await fetch("http://localhost:8080/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ email, password }),
      });
      if (res.ok) {
        const data = await res.json();
        if (data.token && data.token.split(".").length === 3) {
          login(data.user, data.token);
          navigate("/profile");
        } else {
          alert("Invalid token received.");
        }
      } else {
        alert("Invalid email or password");
      }
    } catch (err) {
      console.error("Login error:", err);
      alert("Error logging in. Check console.");
    }
  };

  return (
    <div className="auth-container">
      <h2>Login to Match-Me</h2>
      <form onSubmit={handleLogin} className="auth-form">
        <input
          type="email"
          required
          placeholder="Email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <input
          type="password"
          required
          placeholder="Password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button type="submit" className="primary-btn">Login</button>
      </form>
    </div>
  );
}

export default Login;
