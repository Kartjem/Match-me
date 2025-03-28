package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"matchme-backend/internal/db"
	"matchme-backend/internal/models"
	"matchme-backend/internal/utils"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// returns the full user data to the user themselves
func MeHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Redirect to the full profile endpoint
	redirectURL := fmt.Sprintf("/users/%s/profile", userIDStr)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Password == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(body.Password)
	if err != nil {
		http.Error(w, "Error encrypting password", http.StatusInternalServerError)
		return
	}

	_, err = db.Pool.Exec(context.Background(), `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
	`, body.Email, hashedPassword)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User registered successfully")
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	var user models.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// validate age range
	if user.LookingForMinAge != nil && user.LookingForMaxAge != nil {
		if *user.LookingForMinAge > *user.LookingForMaxAge {
			http.Error(w, "Min age cannot be greater than max age", http.StatusBadRequest)
			return
		}
	}

	// validate country if provided
	if user.Country != nil && *user.Country != "" {
		if !isValidLocation(*user.Country) {
			http.Error(w, "Invalid country specified", http.StatusBadRequest)
			return
		}
	}

	// validate city if provided
	if user.Country != nil && *user.Country != "" && user.City != nil && *user.City != "" {
		if !isValidCity(*user.Country, *user.City) {
			http.Error(w, "Invalid city for the specified country", http.StatusBadRequest)
			return
		}
	}

	query := `
		UPDATE users
		SET
			fname = COALESCE($1, fname),
			surname = COALESCE($2, surname),
			gender = COALESCE($3, gender),
			about = COALESCE($4, about),
			hobbies = COALESCE($5, hobbies),
			interests = COALESCE($6, interests),
			country = COALESCE($7, country),
			city = COALESCE($8, city),
			looking_for_gender = COALESCE($9, looking_for_gender),
			looking_for_min_age = COALESCE($10, looking_for_min_age),
			looking_for_max_age = COALESCE($11, looking_for_max_age),
			profile_picture_url = COALESCE($12, profile_picture_url),
			preferred_hobbies = COALESCE($13, preferred_hobbies),
			preferred_interests = COALESCE($14, preferred_interests),
			birthdate = CASE WHEN $15 != '' THEN $15::date ELSE birthdate END
		WHERE id = $16
	`
	_, err = db.Pool.Exec(context.Background(), query,
		user.Fname,
		user.Surname,
		user.Gender,
		user.About,
		user.Hobbies,
		user.Interests,
		user.Country,
		user.City,
		user.LookingForGender,
		user.LookingForMinAge,
		user.LookingForMaxAge,
		user.Picture,
		user.PreferredHobbies,
		user.PreferredInterests,
		// If empty string, keep existing. Otherwise, parse as date.
		func() string {
			if user.Birthdate == nil {
				return ""
			}
			return *user.Birthdate
		}(),
		userID,
	)
	if err != nil {
		log.Printf("Error updating profile: %v\n", err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../frontend/build/index.html")
}

// /users/{id}, /users/{id}/profile, /users/{id}/bio
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	pathRegex := regexp.MustCompile(`^/users/(\d+)(?:/(profile|bio))?$`)
	parts := pathRegex.FindStringSubmatch(r.URL.Path)
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}

	targetIDStr := parts[1]
	subPath := ""
	if len(parts) >= 3 {
		subPath = parts[2]
	}

	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := getUserByID(targetID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// check if viewer can see this user
	viewerIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	viewerID, _ := strconv.Atoi(viewerIDStr)

	allowed, err := IsUserAllowedToViewProfile(viewerID, targetID)
	if err != nil || !allowed {
		http.NotFound(w, r)
		return
	}

	switch subPath {
	case "profile":
		handleUserProfile(w, user, viewerID)
	case "bio":
		handleUserBio(w, user)
	case "":
		handleUserMinimal(w, user)
	default:
		http.NotFound(w, r)
	}
}

func handleUserMinimal(w http.ResponseWriter, user models.User) {
	// return name and profile link
	name := ""
	if user.Fname != nil && user.Surname != nil {
		name = fmt.Sprintf("%s %s", *user.Fname, *user.Surname)
	}
	resp := map[string]interface{}{
		"id":    user.UserID,
		"name":  name,
		"photo": user.Picture,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleUserProfile(w http.ResponseWriter, user models.User, viewerID int) {
	resp := map[string]interface{}{
		"id":                  user.UserID,
		"fname":               user.Fname,
		"surname":             user.Surname,
		"about":               user.About,
		"profile_picture_url": user.Picture,
		"gender":              user.Gender,
		"birthdate":           user.Birthdate,
		"hobbies":             user.Hobbies,
		"interests":           user.Interests,
		"country":             user.Country,
		"city":                user.City,
		"looking_for_gender":  user.LookingForGender,
		"looking_for_min_age": user.LookingForMinAge,
		"looking_for_max_age": user.LookingForMaxAge,
		"preferred_hobbies":   user.PreferredHobbies,
		"preferred_interests": user.PreferredInterests,
	}
	if viewerID == user.UserID {
		resp["email"] = user.Email
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleUserBio(w http.ResponseWriter, user models.User) {
	resp := map[string]interface{}{
		"id":        user.UserID,
		"gender":    user.Gender,
		"birthdate": user.Birthdate,
		"hobbies":   user.Hobbies,
		"about":     user.About,
		"interests": user.Interests,
		"country":   user.Country,
		"city":      user.City,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// checks recommended, pending, or connected
func IsUserAllowedToViewProfile(viewerID, targetID int) (bool, error) {
	if viewerID == targetID {
		return true, nil
	}

	var count int
	err := db.Pool.QueryRow(context.Background(), `
        SELECT COUNT(*)
        FROM connections
        WHERE
        ((user_id = $1 AND connected_user_id = $2)
         OR (user_id = $2 AND connected_user_id = $1))
        AND status IN ('pending','accepted')
    `, viewerID, targetID).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// check if the user is in recommendations
	err = db.Pool.QueryRow(context.Background(), `
        SELECT COUNT(*)
        FROM recommendations
        WHERE user_id = $1 AND recommended_user_id = $2
    `, viewerID, targetID).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	log.Printf("Profile access denied: Viewer ID: %d, Target ID: %d", viewerID, targetID)

	return false, nil
}

func getUserByID(userID int) (models.User, error) {
	var u models.User
	err := db.Pool.QueryRow(context.Background(), `
        SELECT
            id,
            email,
            password,
            fname,
            surname,
            gender,
            TO_CHAR(birthdate, 'YYYY-MM-DD') AS birthdate,
            about,
            hobbies,
            interests,
            country,
            city,
            looking_for_gender,
            looking_for_min_age,
            looking_for_max_age,
            profile_picture_url,
            preferred_hobbies,
            preferred_interests
        FROM users
        WHERE id = $1
    `, userID).Scan(
		&u.UserID,
		&u.Email,
		&u.Password,
		&u.Fname,
		&u.Surname,
		&u.Gender,
		&u.Birthdate,
		&u.About,
		&u.Hobbies,
		&u.Interests,
		&u.Country,
		&u.City,
		&u.LookingForGender,
		&u.LookingForMinAge,
		&u.LookingForMaxAge,
		&u.Picture,
		&u.PreferredHobbies,
		&u.PreferredInterests,
	)
	return u, err
}

func isValidLocation(country string) bool {
	for _, c := range validLocations {
		if strings.EqualFold(c, country) {
			return true
		}
	}
	return false
}

func isValidCity(country, city string) bool {
	cities, ok := cityOptionsMap[strings.Title(strings.ToLower(country))]
	if !ok {
		cities = cityOptionsMap[country]
	}
	for _, cc := range cities {
		if strings.EqualFold(cc, city) {
			return true
		}
	}
	return false
}
