import React from "react";

function AdminPanel() {
    const API_URL = "http://localhost:8080/admin";
    const ADMIN_SECRET = "supersecureadminpassword";

    const callAPI = async (endpoint) => {
        const res = await fetch(`${API_URL}/${endpoint}`, {
            method: "POST",
            headers: { "X-Admin-Secret": ADMIN_SECRET },
        });
        const data = await res.json();
        alert(data.message || "Action completed");
    };

    return (
        <div className="admin-panel">
            <h2>Admin Controls</h2>
            <button onClick={() => callAPI("load-fake-users")}>Load Fake Users</button>
            <button onClick={() => callAPI("reset-database")}>Reset Database</button>
        </div>
    );
}

export default AdminPanel;
