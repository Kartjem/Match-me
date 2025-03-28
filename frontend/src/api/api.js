const BASE_URL = 'http://localhost:8080';

// register with email + password
export async function registerUser(email, password) {
    const response = await fetch(`${BASE_URL}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
    });
    return response;
}

// log in
export async function loginUser(email, password) {
    const response = await fetch(`${BASE_URL}/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, password }),
    });
    return response;
}

// log out
export async function logoutUser() {
    const response = await fetch(`${BASE_URL}/logout`, {
        method: 'POST',
        credentials: 'include',
    });
    return response;
}

// get my user data
export async function fetchMe() {
    const response = await fetch(`${BASE_URL}/me`, {
        credentials: 'include',
    });
    return response;
}

// update profile
export async function updateProfile(profileData) {
    const response = await fetch(`${BASE_URL}/update-profile`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(profileData),
    });
    return response;
}

// fetch up to 10 recommended IDs
export async function fetchRecommendations() {
    const response = await fetch(`${BASE_URL}/recommendations`, {
        credentials: 'include',
    });
    return response;
}

// dismiss a recommendation
export async function dismissRecommendation(dismissedUserId) {
    const response = await fetch(`${BASE_URL}/recommendations/dismiss`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dismissedUserId }),
    });
    return response;
}

// connect a user
export async function connectUser(targetUserId) {
    const response = await fetch(`${BASE_URL}/connect`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ targetUserId }),
    });
    return response;
}

// get accepted connections
export async function fetchConnections() {
    const response = await fetch(`${BASE_URL}/connections`, {
        credentials: 'include',
    });
    return response;
}

// disconnect a user
export async function disconnectUser(targetUserId) {
    const response = await fetch(`${BASE_URL}/connections`, {
        method: 'DELETE',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ targetUserId }),
    });
    return response;
}

// minimal user info
export async function fetchUserMinimal(userId) {
    const response = await fetch(`${BASE_URL}/users/${userId}`, {
        credentials: 'include',
    });
    return response;
}

// full bio
export async function fetchUserBio(userId) {
    const response = await fetch(`${BASE_URL}/users/${userId}/bio`, {
        credentials: 'include',
    });
    return response;
}

export async function fetchUnreadMessages() {
    const response = await fetch("http://localhost:8080/chats/unread", {
        method: "GET",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
    });
    if (!response.ok) {
        console.error("Failed to fetch unread messages", await response.text());
        return {};
    }
    return response.json();
}

export async function fetchOnlineStatus() {
    const response = await fetch("http://localhost:8080/users/online-status", {
        method: "GET",
        credentials: "include",
        headers: {
            "Content-Type": "application/json",
        },
    });
    if (!response.ok) {
        console.error("Failed to fetch online status", await response.text());
        return {};
    }
    return response.json();
}

