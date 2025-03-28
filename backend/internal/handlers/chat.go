package handlers

import (
	"context"
	"encoding/json"
	"log"
	"matchme-backend/internal/db"
	"matchme-backend/internal/models"
	"matchme-backend/internal/utils"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type       string `json:"type"`
	UserID     string `json:"user_id,omitempty"`
	Token      string `json:"token,omitempty"`
	SenderID   int    `json:"sender_id,omitempty"`
	ReceiverID int    `json:"receiver_id,omitempty"`
	Content    string `json:"content,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
	Status     string `json:"status,omitempty"`
}

const (
	pongWait   = 120 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 15 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients      = make(map[int]*websocket.Conn)
	onlineUsers  = make(map[int]bool)
	clientsMutex = sync.Mutex{}
)

func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		http.Error(w, "Unable to establish WebSocket connection", http.StatusInternalServerError)
		return
	}

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var connectMsg WSMessage
	if err := conn.ReadJSON(&connectMsg); err != nil || connectMsg.Type != "connect" {
		log.Println("Failed to read initial connect message or invalid type")
		conn.Close()
		return
	}

	log.Printf("Received token in WS connect message: %s\n", connectMsg.Token)

	// validate the token
	userIDStr, err := utils.ExtractUserIDFromToken(connectMsg.Token)
	if err != nil {
		log.Println("WebSocket authentication failed:", err)
		conn.Close()
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Println("Invalid user ID from token:", err)
		conn.Close()
		return
	}

	clientsMutex.Lock()
	clients[userID] = conn
	onlineUsers[userID] = true
	clientsMutex.Unlock()

	log.Printf("User %d connected via WebSocket\n", userID)

	// send periodic ping messages
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Error sending ping:", err)
				return
			}
		}
	}()

	// on connection send any undelivered messages
	undelivered, err := fetchUndeliveredMessages(userID)
	if err == nil {
		for _, msg := range undelivered {
			sendWSMessage(conn, WSMessage{
				Type:       "message",
				SenderID:   msg.SenderID,
				ReceiverID: msg.ReceiverID,
				Content:    msg.Message,
				Timestamp:  msg.CreatedAt,
			})
			markMessageAsDelivered(msg.ID)
		}
	} else {
		log.Println("Error fetching undelivered messages:", err)
	}

	// listen for incoming WebSocket messages
	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket read error for user %d: %v\n", userID, err)
			break
		}

		switch msg.Type {
		case "message":
			if msg.SenderID != userID {
				log.Printf("User %d attempted to send message with mismatched sender ID\n", userID)
				continue
			}
			msg.Timestamp = time.Now().Format(time.RFC3339)
			savedMsg, err := saveChatMessage(msg)
			if err != nil {
				log.Printf("Error saving message from user %d: %v\n", userID, err)
				continue
			}
			clientsMutex.Lock()
			receiverConn, online := clients[msg.ReceiverID]
			clientsMutex.Unlock()
			if online {
				sendWSMessage(receiverConn, WSMessage{
					Type:       "message",
					SenderID:   msg.SenderID,
					ReceiverID: msg.ReceiverID,
					Content:    msg.Content,
					Timestamp:  msg.Timestamp,
				})
				markMessageAsDelivered(savedMsg.ID)
				sendWSMessage(conn, WSMessage{Type: "delivered"})
			} else {
				log.Printf("User %d is offline; message stored for later delivery\n", msg.ReceiverID)
			}
		case "typing":
			clientsMutex.Lock()
			if receiverConn, ok := clients[msg.ReceiverID]; ok {
				sendWSMessage(receiverConn, WSMessage{
					Type:       "typing",
					SenderID:   msg.SenderID,
					ReceiverID: msg.ReceiverID,
					Status:     msg.Status,
				})
			}
			clientsMutex.Unlock()
		case "disconnect":
			log.Printf("User %d requested disconnect\n", userID)
			goto DISCONNECT
		default:
			log.Printf("Unknown WS message type from user %d: %s\n", userID, msg.Type)
		}
	}

DISCONNECT:
	clientsMutex.Lock()
	delete(clients, userID)
	delete(onlineUsers, userID)
	clientsMutex.Unlock()
	conn.Close()
	log.Printf("User %d disconnected from WebSocket\n", userID)
}

// writes a JSON message to the WebSocket
func sendWSMessage(conn *websocket.Conn, msg WSMessage) {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := conn.WriteJSON(msg); err != nil {
		log.Println("Error sending WS message:", err)
	}
}

func saveChatMessage(msg WSMessage) (models.Chat, error) {
	var chatRecord models.Chat
	query := `
		INSERT INTO chats (sender_id, receiver_id, message, created_at, delivered)
		VALUES ($1, $2, $3, $4, false)
		RETURNING id, sender_id, receiver_id, message, TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at, delivered
	`
	err := db.Pool.QueryRow(context.Background(), query,
		msg.SenderID, msg.ReceiverID, msg.Content, msg.Timestamp,
	).Scan(&chatRecord.ID, &chatRecord.SenderID, &chatRecord.ReceiverID, &chatRecord.Message, &chatRecord.CreatedAt, &chatRecord.Delivered)
	return chatRecord, err
}

func markMessageAsDelivered(messageID int) {
	_, err := db.Pool.Exec(context.Background(), `
		UPDATE chats SET delivered = true WHERE id = $1
	`, messageID)
	if err != nil {
		log.Println("Error marking message as delivered:", err)
	}
}

func fetchUndeliveredMessages(userID int) ([]models.Chat, error) {
	rows, err := db.Pool.Query(context.Background(), `
		SELECT id, sender_id, receiver_id, message, TO_CHAR(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at, delivered
		FROM chats
		WHERE receiver_id = $1 AND delivered = false
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []models.Chat
	for rows.Next() {
		var msg models.Chat
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Message, &msg.CreatedAt, &msg.Delivered); err == nil {
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

func ChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil || userIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	receiverIDStr := r.URL.Query().Get("receiver_id")
	if receiverIDStr == "" {
		http.Error(w, "Receiver ID required", http.StatusBadRequest)
		return
	}
	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		http.Error(w, "Invalid receiver ID", http.StatusBadRequest)
		return
	}

	limit := 10
	offset := 0
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			offset = (page - 1) * limit
		}
	}

	rows, err := db.Pool.Query(context.Background(), `
        SELECT sender_id, receiver_id, message, created_at, delivered
        FROM chats
        WHERE (sender_id = $1 AND receiver_id = $2)
           OR (sender_id = $2 AND receiver_id = $1)
        ORDER BY created_at DESC
        LIMIT $3 OFFSET $4
    `, userID, receiverID, limit, offset)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		var ts time.Time
		if err := rows.Scan(&chat.SenderID, &chat.ReceiverID, &chat.Message, &ts, &chat.Delivered); err != nil {
			log.Println("Error scanning chat message:", err)
			continue
		}
		chat.CreatedAt = ts.Format(time.RFC3339)
		chats = append(chats, chat)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

func UnreadMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil || userIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	rows, err := db.Pool.Query(context.Background(), `
		SELECT sender_id, COUNT(*)
		FROM chats
		WHERE receiver_id = $1 AND delivered = false
		GROUP BY sender_id
	`, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	unread := make(map[int]int)
	for rows.Next() {
		var senderID, count int
		if err := rows.Scan(&senderID, &count); err == nil {
			unread[senderID] = count
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unread)
}

func OnlineStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	json.NewEncoder(w).Encode(onlineUsers)
}
