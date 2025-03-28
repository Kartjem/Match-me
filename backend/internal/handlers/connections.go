package handlers

import (
	"context"
	"encoding/json"
	"log"
	"matchme-backend/internal/db"
	"matchme-backend/internal/utils"
	"net/http"
	"strconv"
)

func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	var connectRequest struct {
		TargetUserID int `json:"targetUserId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&connectRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// check if the target user has already liked the current user
	var existingStatus string
	err = db.Pool.QueryRow(context.Background(), `
        SELECT status FROM connections
        WHERE user_id = $1 AND connected_user_id = $2
    `, connectRequest.TargetUserID, userID).Scan(&existingStatus)

	if err == nil && existingStatus == "pending" {
		// if the target user has also liked the requester, update both to "accepted"
		_, err = db.Pool.Exec(context.Background(), `
            UPDATE connections
            SET status = 'accepted'
            WHERE (user_id = $1 AND connected_user_id = $2) 
               OR (user_id = $2 AND connected_user_id = $1)
        `, userID, connectRequest.TargetUserID)
		if err != nil {
			log.Printf("Database error while accepting connection: %v\n", err)
			http.Error(w, "Failed to accept connection", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Connection accepted!"})
		return
	}

	// insert a new pending connection
	_, err = db.Pool.Exec(context.Background(), `
        INSERT INTO connections (user_id, connected_user_id, status)
        VALUES ($1, $2, 'pending')
        ON CONFLICT (user_id, connected_user_id) DO NOTHING
    `, userID, connectRequest.TargetUserID)
	if err != nil {
		log.Printf("Database error while creating connection: %v\n", err)
		http.Error(w, "Failed to process connection request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection request sent successfully"})
}

func FetchIncomingRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.Pool.Query(context.Background(), `
        SELECT user_id
        FROM connections
        WHERE connected_user_id = $1 AND status = 'pending'
    `, userID)
	if err != nil {
		http.Error(w, "Error fetching connection requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []int
	for rows.Next() {
		var requesterID int
		if err := rows.Scan(&requesterID); err == nil {
			requests = append(requests, requesterID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func RespondToConnectionRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		RequesterID int    `json:"requesterId"`
		Action      string `json:"action"` // "accept" or "reject"
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var newStatus string
	if request.Action == "accept" {
		newStatus = "accepted"
	} else if request.Action == "reject" {
		newStatus = "rejected"
	} else {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(), `
        UPDATE connections
        SET status = $1
        WHERE user_id = $2 AND connected_user_id = $3 AND status = 'pending'
    `, newStatus, request.RequesterID, userID)
	if err != nil {
		http.Error(w, "Error updating connection status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection updated successfully"})
}

func ConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAcceptedConnections(w, r)
	case http.MethodDelete:
		disconnectHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func getAcceptedConnections(w http.ResponseWriter, r *http.Request) {
	userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	rows, err := db.Pool.Query(context.Background(), `
        SELECT connected_user_id
        FROM connections
        WHERE user_id = $1 AND status = 'accepted'
        UNION
        SELECT user_id
        FROM connections
        WHERE connected_user_id = $1 AND status = 'accepted'
    `, userID)
	if err != nil {
		http.Error(w, "Error fetching connections", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var connectedIDs []int
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err == nil {
			connectedIDs = append(connectedIDs, cid)
		}
	}

	if connectedIDs == nil {
		connectedIDs = []int{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectedIDs)
}

func disconnectHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		TargetUserID int `json:"targetUserId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(), `
        DELETE FROM connections
        WHERE (user_id = $1 AND connected_user_id = $2)
           OR (user_id = $2 AND connected_user_id = $1)
    `, userID, request.TargetUserID)
	if err != nil {
		http.Error(w, "Error disconnecting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Disconnected successfully"})
}
