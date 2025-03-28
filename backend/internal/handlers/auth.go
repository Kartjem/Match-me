package handlers

import (
	"context"
	"encoding/json"
	"matchme-backend/internal/db"
	"matchme-backend/internal/models"
	"matchme-backend/internal/utils"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var user models.User
	err = db.Pool.QueryRow(context.Background(), `
		SELECT id, email, password FROM users WHERE email = $1
	`, creds.Email).Scan(&user.UserID, &user.Email, &user.Password)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(user.UserID, user.Email)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User logged in successfully",
		"token":   token,
		"user":    user,
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
