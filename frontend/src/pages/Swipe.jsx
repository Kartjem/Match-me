import React, { useEffect, useState, useContext } from "react";
import SwipeCard from "../components/SwipeCard";
import { AuthContext } from "../context/AuthContext";
import "./Swipe.css";

function Swipe() {
    const { user } = useContext(AuthContext);
    const [recommendedUsers, setRecommendedUsers] = useState([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [profileComplete, setProfileComplete] = useState(true);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        async function fetchProfileCompletion() {
            try {
                const res = await fetch("http://localhost:8080/me", { credentials: "include" });
                if (!res.ok) {
                    console.error("Error fetching profile data");
                    return;
                }
                const userData = await res.json();
                const isComplete = userData.fname && userData.surname && userData.gender &&
                    userData.city && userData.country && userData.about &&
                    userData.hobbies && userData.interests && userData.birthdate;
                setProfileComplete(!!isComplete);
            } catch (err) {
                console.error("Error checking profile completion:", err);
            }
        }

        async function fetchRecommendations() {
            try {
                const res = await fetch("http://localhost:8080/recommendations", { credentials: "include" });
                if (!res.ok) {
                    console.error("Failed to fetch recommendations");
                    return;
                }

                const recommendedIds = await res.json();
                console.log("Recommended user IDs:", recommendedIds);

                const userDataPromises = recommendedIds.map(async (userId) => {
                    console.log(`Fetching user data for ID: ${userId}`);
                    const userRes = await fetch(`http://localhost:8080/users/${userId}`, { credentials: "include" });

                    if (!userRes.ok) {
                        console.warn(`User ${userId} not found (404)`);
                        return null;
                    }

                    const userData = await userRes.json();
                    console.log("Recommended IDs received:", recommendedIds);
                    console.log(`Fetched user data:`, userData);
                    return { id: userId, ...userData };
                });

                const results = await Promise.all(userDataPromises);
                setRecommendedUsers(results.filter((u) => u !== null));
            } catch (err) {
                console.error("Error fetching recommendations", err);
            } finally {
                setLoading(false);
            }
        }



        async function initializeSwipe() {
            await fetchProfileCompletion();
            if (profileComplete) {
                await fetchRecommendations();
            }
        }

        initializeSwipe();
    }, [profileComplete]);

    const handleLike = async () => {
        if (!recommendedUsers[currentIndex]) return;

        try {
            const res = await fetch("http://localhost:8080/connect", {
                method: "POST",
                credentials: "include",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ targetUserId: recommendedUsers[currentIndex].id }),
            });

            const responseData = await res.json();
            console.log("Like response:", responseData);

            if (!res.ok) {
                console.error("Error connecting:", responseData);
                return;
            }

            setRecommendedUsers((prev) => prev.filter((_, index) => index !== currentIndex));
            setCurrentIndex(0);

        } catch (error) {
            console.error("Error sending connect request", error);
        }
    };


    const handleNope = async () => {
        if (!recommendedUsers[currentIndex]) return;
        try {
            const res = await fetch("http://localhost:8080/recommendations/dismiss", {
                method: "POST",
                credentials: "include",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ dismissedUserId: recommendedUsers[currentIndex].id }),
            });

            if (!res.ok) {
                console.error("Error dismissing recommendation");
                return;
            }

            setCurrentIndex((prevIndex) => prevIndex + 1);
        } catch (error) {
            console.error("Error dismissing recommendation", error);
        }
    };


    if (loading) {
        return <div className="swipe-container"><h2>Loading recommendations...</h2></div>;
    }

    if (!profileComplete) {
        return <div className="swipe-container"><h2>Please complete your profile before swiping.</h2></div>;
    }

    if (currentIndex >= recommendedUsers.length) {
        return <div className="swipe-container"><h2>No more recommendations</h2></div>;
    }

    return (
        <div className="swipe-container">
            <SwipeCard user={recommendedUsers[currentIndex]} />
            <div className="swipe-buttons">
                <button className="nope-btn" onClick={handleNope}>Nope</button>
                <button className="like-btn" onClick={handleLike}>Like</button>
            </div>
        </div>
    );
}

export default Swipe;
