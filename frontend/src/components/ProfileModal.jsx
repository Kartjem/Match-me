import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import "./ProfileModal.css";

function ProfileModal({ userId, onClose }) {
  const [profile, setProfile] = useState(null);
  const [online, setOnline] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    async function fetchProfileData() {
      try {
        const [resProfile, resBio] = await Promise.all([
          fetch(`http://localhost:8080/users/${userId}/profile`, {
            credentials: "include",
          }),
          fetch(`http://localhost:8080/users/${userId}/bio`, {
            credentials: "include",
          }),
        ]);
        if (resProfile.ok && resBio.ok) {
          const dataProfile = await resProfile.json();
          const dataBio = await resBio.json();
          setProfile({
            fname: dataProfile.fname,
            surname: dataProfile.surname,
            about: dataProfile.about,
            profile_picture_url: dataProfile.profile_picture_url,
            hobbies: dataBio.hobbies,
          });
        } else {
          console.error("Error fetching profile or bio data");
        }
      } catch (err) {
        console.error("Error fetching profile data:", err);
      }
    }
    fetchProfileData();
  }, [userId]);

  useEffect(() => {
    async function fetchOnlineStatus() {
      try {
        const res = await fetch("http://localhost:8080/users/online-status", {
          credentials: "include",
        });
        if (res.ok) {
          const status = await res.json();
          setOnline(!!status[userId]);
        }
      } catch (err) {
        console.error("Error fetching online status:", err);
      }
    }
    fetchOnlineStatus();
  }, [userId]);

  const handleChat = () => {
    navigate("/chat", { state: { receiverId: userId } });
    onClose();
  };

  const handleDisconnect = async () => {
    if (window.confirm("Are you sure you want to disconnect from this user?")) {
      try {
        const res = await fetch("http://localhost:8080/connections", {
          method: "DELETE",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ targetUserId: userId }),
        });
        if (res.ok) {
          alert("Disconnected successfully.");
          onClose();
        } else {
          const errorText = await res.text();
          alert("Error disconnecting: " + errorText);
        }
      } catch (err) {
        console.error("Error disconnecting:", err);
        alert("Error disconnecting. See console for details.");
      }
    }
  };

  if (!profile) {
    return (
      <div className="profile-modal-overlay" onClick={onClose}>
        <div className="profile-modal" onClick={(e) => e.stopPropagation()}>
          <p>Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="profile-modal-overlay" onClick={onClose}>
      <div className="profile-modal" onClick={(e) => e.stopPropagation()}>
        <button className="close-button" onClick={onClose}>
          ×
        </button>
        <img
          src={
            profile.profile_picture_url ||
            "https://via.placeholder.com/150?text=No+Image"
          }
          alt="Profile"
          className="profile-picture"
        />
        <h2>
          {profile.fname} {profile.surname}
        </h2>
        {online && <div className="online-indicator">Online</div>}
        <p className="bio">{profile.about || "No bio available."}</p>
        <h3>Hobbies</h3>
        {profile.hobbies && Array.isArray(profile.hobbies) && profile.hobbies.length > 0 ? (
          <ul className="hobbies-list">
            {profile.hobbies.map((hobby, index) => (
              <li key={index}>{hobby}</li>
            ))}
          </ul>
        ) : (
          <p>No hobbies listed.</p>
        )}
        <button className="chat-button" onClick={handleChat}>
          Chat
        </button>
        <button className="disconnect-button" onClick={handleDisconnect}>
          Disconnect
        </button>
      </div>
    </div>
  );
}

export default ProfileModal;