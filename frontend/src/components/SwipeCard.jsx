import React, { useEffect, useState } from 'react';
import './SwipeCard.css';

function SwipeCard({ user }) {
    const [bioData, setBioData] = useState(null);

    useEffect(() => {
        async function fetchBio() {
            try {
                const res = await fetch(`http://localhost:8080/users/${user.id}/bio`, {
                    credentials: 'include'
                });
                if (res.ok) {
                    const data = await res.json();
                    setBioData(data);
                }
            } catch (err) {
                console.error('Error fetching user bio', err);
            }
        }
        fetchBio();
    }, [user.id]);

    return (
        <div className="swipe-card">
            {/* if we have a photo url, show it, else fallback */}
            {user.photo ? (
                <img src={user.photo} alt={user.name} className="swipe-card-img" />
            ) : (
                <div className="swipe-card-placeholder">No Image</div>
            )}
            <div className="swipe-card-info">
                <h2>{user.name}</h2>
                {bioData && (
                    <>
                        <p><strong>City:</strong> {bioData.city}</p>
                        <p><strong>Interests:</strong> {(bioData.interests || []).join(', ')}</p>
                        <p><strong>Hobbies:</strong> {bioData.hobbies}</p>
                        {/* etc. */}
                    </>
                )}
            </div>
        </div>
    );
}

export default SwipeCard;
