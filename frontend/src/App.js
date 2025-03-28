import React, { useContext } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Profile from "./pages/Profile";
import Swipe from "./pages/Swipe";
import Matches from "./pages/Matches";
import Chat from "./pages/Chat";
import NavBar from "./components/NavBar";
import AdminPanel from "./pages/AdminPanel";
import { AuthContext } from "./context/AuthContext";

function App() {
  const { isAuthenticated } = useContext(AuthContext);
  return (
    <Router key={isAuthenticated ? "logged-in" : "logged-out"}>
      <NavBar />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        {/* Protected Routes */}
        <Route path="/profile" element={<Profile />} />
        <Route path="/swipe" element={<Swipe />} />
        <Route path="/matches" element={<Matches />} />
        <Route path="/chat/*" element={<Chat />} />
        <Route path="/admin" element={<AdminPanel />} />
        <Route path="*" element={<Home />} />
      </Routes>
    </Router>
  );
}

export default App;
