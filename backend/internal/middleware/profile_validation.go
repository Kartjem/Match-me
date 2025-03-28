package middleware

import (
	"context"
	"matchme-backend/internal/db"
	"matchme-backend/internal/utils"
	"net/http"
	"strconv"
)

// ensures the user has a "complete" profile before proceeding
func RequireCompleteProfile(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		var isComplete bool
		err = db.Pool.QueryRow(context.Background(), `
            SELECT COUNT(*) = 1
            FROM users
            WHERE id = $1
              AND fname IS NOT NULL
              AND surname IS NOT NULL
              AND gender IS NOT NULL
              AND birthdate IS NOT NULL
              AND hobbies IS NOT NULL
              AND about IS NOT NULL
              AND interests IS NOT NULL
              AND country IS NOT NULL
              AND city IS NOT NULL
              AND looking_for_gender IS NOT NULL
              AND looking_for_min_age IS NOT NULL
              AND looking_for_max_age IS NOT NULL
        `, userID).Scan(&isComplete)

		if err != nil || !isComplete {
			http.Error(w, "Profile incomplete. Please complete your profile.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := utils.ExtractUserIDFromTokenFromCookie(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
