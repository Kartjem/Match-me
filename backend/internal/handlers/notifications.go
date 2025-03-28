package handlers

import (
	"context"
	"encoding/json"
	"log"
	"matchme-backend/internal/db"
	"matchme-backend/internal/utils"
	"net/http"
)

func FetchNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.Pool.Query(context.Background(), `
        SELECT id, type, message, read, created_at
        FROM notifications
        WHERE user_id = $1
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		log.Printf("Error fetching notifications: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	for rows.Next() {
		var (
			id        int
			notifType string
			message   string
			read      bool
			createdAt string
		)
		if err := rows.Scan(&id, &notifType, &message, &read, &createdAt); err != nil {
			log.Printf("Error scanning notification: %v\n", err)
			continue
		}
		notifications = append(notifications, map[string]interface{}{
			"id":        id,
			"type":      notifType,
			"message":   message,
			"read":      read,
			"createdAt": createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func MarkNotificationAsReadHandler(w http.ResponseWriter, r *http.Request) {
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
		NotificationID int `json:"notificationId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(), `
        UPDATE notifications
        SET read = TRUE
        WHERE id = $1 AND user_id = $2
    `, request.NotificationID, userID)
	if err != nil {
		log.Printf("Error marking notification as read: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
}

func CreateNotification(userID int, notifType, message string) error {
	_, err := db.Pool.Exec(context.Background(), `
        INSERT INTO notifications (user_id, type, message)
        VALUES ($1, $2, $3)
    `, userID, notifType, message)
	return err
}
