import React, { useContext } from "react";
import { Link } from "react-router-dom";
import { AuthContext } from "../context/AuthContext";
import "./NavBar.css";

function NavBar() {
  const { isAuthenticated, logout } = useContext(AuthContext);

  return (
    <nav className="navbar">
      <div className="navbar-brand">Match-Me</div>
      <ul className="navbar-links">
        <li>
          <Link to="/">Home</Link>
        </li>
        {isAuthenticated ? (
          <>
            <li>
              <Link to="/swipe">Discover</Link>
            </li>
            <li>
              <Link to="/matches">Matches</Link>
            </li>
            <li>
              <Link to="/chat">Chat</Link>
            </li>
            <li>
              <Link to="/profile">My Profile</Link>
            </li>
            <li>
              <button className="logout-btn" onClick={logout}>
                Logout
              </button>
            </li>
          </>
        ) : (
          <>
            <li>
              <Link to="/login">Login</Link>
            </li>
            <li>
              <Link to="/register">Sign Up</Link>
            </li>
          </>
        )}
      </ul>
    </nav>
  );
}

export default NavBar;
