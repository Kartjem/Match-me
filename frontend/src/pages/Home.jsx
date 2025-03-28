import React from "react";
import { useNavigate } from "react-router-dom";
import "./Home.css";

function Home() {
    const navigate = useNavigate();

    return (
        <div className="home-container">
            <h1>Welcome to Match-Me</h1>
            <p>
                Connect with amazing people near you.
                Get started by creating a profile and discovering new matches!
            </p>
            <div className="home-buttons">
                <button className="primary-btn" onClick={() => navigate("/register")}>
                    Get Started
                </button>
                <button className="secondary-btn" onClick={() => navigate("/login")}>
                    Log In
                </button>
            </div>
        </div>
    );
}

export default Home;
