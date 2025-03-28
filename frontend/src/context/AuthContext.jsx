import React, { createContext, useState, useEffect } from "react";

export const AuthContext = createContext(null);

function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const checkSession = async () => {
    try {
      const res = await fetch("http://localhost:8080/me", {
        method: "GET",
        credentials: "include",
      });
      if (res.ok) {
        const data = await res.json();
        setUser(data);
      } else {
        setUser(null);
      }
    } catch (err) {
      console.error("Error checking session:", err);
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkSession();
  }, []);

  const login = (userData, token) => {
    setUser(userData);
    localStorage.setItem("authToken", token);
  };

  const logout = async () => {
    try {
      await fetch("http://localhost:8080/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch (err) {
      console.error("Logout error:", err);
    } finally {
      setUser(null);
      localStorage.removeItem("authToken");
    }
  };

  return (
    <AuthContext.Provider
      value={{ user, loading, isAuthenticated: !!user, login, logout, checkSession }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export default AuthProvider;
