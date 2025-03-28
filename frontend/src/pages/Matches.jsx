import React, { useEffect, useState } from "react";
import { fetchConnections, fetchUserMinimal } from "../api/api";
import ProfileModal from "../components/ProfileModal";
import "./Matches.css";

function Matches() {
  const [matchedUsers, setMatchedUsers] = useState([]);
  const [selectedUserId, setSelectedUserId] = useState(null);
  const [modalVisible, setModalVisible] = useState(false);

  useEffect(() => {
    async function fetchMatches() {
      try {
        const res = await fetchConnections({ credentials: "include" });
        if (!res.ok) {
          throw new Error(`Error fetching connections: ${res.status} ${res.statusText}`);
        }
        const connectionIds = await res.json();
        const userDataPromises = connectionIds.map(async (id) => {
          const userRes = await fetchUserMinimal(id);
          if (!userRes.ok) {
            console.warn(`User ${id} not found`);
            return null;
          }
          return await userRes.json();
        });
        setMatchedUsers((await Promise.all(userDataPromises)).filter((u) => u !== null));
      } catch (err) {
        console.error("Error loading matches:", err);
        setMatchedUsers([]);
      }
    }
    fetchMatches();
  }, []);

  const openModal = (userId) => {
    setSelectedUserId(userId);
    setModalVisible(true);
  };

  const closeModal = () => {
    setSelectedUserId(null);
    setModalVisible(false);
  };

  return (
    <div className="matches-container">
      <h2>Your Matches</h2>
      {matchedUsers.length === 0 ? (
        <p>You have no matches yet.</p>
      ) : (
        <div className="matches-grid">
          {matchedUsers.map((userData) => (
            <div
              key={userData.id}
              className="match-card"
              onClick={() => openModal(userData.id)}
            >
              {/* Render image only if profile_picture_url exists */}
              {userData.profile_picture_url && (
                <img src={userData.profile_picture_url} alt={userData.name} />
              )}
              <h3>{userData.name}</h3>
            </div>
          ))}
        </div>
      )}
      {modalVisible && selectedUserId && (
        <ProfileModal userId={selectedUserId} onClose={closeModal} />
      )}
    </div>
  );
}

export default Matches;
