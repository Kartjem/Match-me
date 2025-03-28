package main

import (
	"log"
	"net/http"

	"matchme-backend/internal/db"
	"matchme-backend/internal/handlers"
	"matchme-backend/internal/middleware"
)

// adds CORS headers to the response
func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-Secret")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database connection
	db.InitDB()
	defer db.CloseDB()

	// Apply DB migrations
	db.ApplyMigrations()

	// Serve static frontend
	fs := http.FileServer(http.Dir("../frontend/build"))
	http.Handle("/", enableCORS(fs))

	// Public routes
	http.Handle("/register", enableCORS(http.HandlerFunc(handlers.RegisterHandler)))
	http.Handle("/login", enableCORS(http.HandlerFunc(handlers.LoginHandler)))
	http.Handle("/logout", enableCORS(http.HandlerFunc(handlers.LogoutHandler)))

	// Protected routes
	http.Handle("/me", enableCORS(middleware.RequireAuth(http.HandlerFunc(handlers.MeHandler))))

	// Serve "/profile" as a fallback to index
	http.Handle("/profile", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.ProfileHandler))))

	// Update user’s data
	http.Handle("/update-profile", enableCORS(middleware.RequireAuth(http.HandlerFunc(handlers.UpdateProfileHandler))))

	// Connect/disconnect routes
	http.Handle("/connect", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.ConnectHandler))))
	http.Handle("/connections/requests", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.FetchIncomingRequestsHandler))))
	http.Handle("/connections/respond", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.RespondToConnectionRequestHandler))))
	http.Handle("/connections", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.ConnectionsHandler))))

	// Register specific endpoints BEFORE the generic /users/ handler.
	http.Handle("/users/online-status", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.OnlineStatusHandler))))

	// Generic /users/ endpoints for profile access.
	http.Handle("/users/", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.UsersHandler))))

	// Chat routes
	http.Handle("/ws/chat", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.ChatWebSocketHandler))))
	http.Handle("/chats", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.ChatHistoryHandler))))
	http.Handle("/chats/unread", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.UnreadMessagesHandler))))

	// Notifications
	http.Handle("/notifications", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.FetchNotificationsHandler))))
	http.Handle("/notifications/mark-as-read", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.MarkNotificationAsReadHandler))))

	// Recommendations
	http.Handle("/recommendations", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.RecommendationsHandler))))
	http.Handle("/recommendations/dismiss", enableCORS(middleware.RequireCompleteProfile(http.HandlerFunc(handlers.DismissRecommendationHandler))))

	// Admin panel route
	http.Handle("/admin", enableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/build/index.html")
	})))

	// Admin routes
	http.Handle("/admin/load-fake-users", enableCORS(http.HandlerFunc(handlers.LoadFictitiousUsers)))
	http.Handle("/admin/reset-database", enableCORS(http.HandlerFunc(handlers.ResetDatabase)))

	log.Println("Server is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
