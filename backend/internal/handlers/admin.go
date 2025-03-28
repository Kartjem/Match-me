package handlers

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"matchme-backend/internal/db"
	"matchme-backend/internal/models"
)

const adminSecret = "supersecureadminpassword"

var interestsList = []string{"Movies", "Music", "Sports", "Coding", "Nature", "Pets", "Art", "Theatre"}
var hobbiesList = []string{"Reading", "Gaming", "Cooking", "Art", "Sports", "Music", "Travel", "Photography"}
var genders = []string{"male", "female", "other"}

// generate a random user
func generateFakeUser() models.User {
	rand.Seed(time.Now().UnixNano())

	country := validLocations[rand.Intn(len(validLocations))]
	city := cityOptionsMap[country][rand.Intn(len(cityOptionsMap[country]))]

	birthdate := time.Now().AddDate(-rand.Intn(40)-18, 0, 0).Format("2006-01-02")

	// randomly select 3-5 interests & hobbies
	selectedInterests := randomSelection(interestsList, 3, 5)
	selectedHobbies := randomSelection(hobbiesList, 3, 5)

	// hash the password
	password := "password123" // Default password for fake users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	return models.User{
		Email:            strings.ToLower(randomString(8) + "@test.com"),
		Password:         string(hashedPassword), // Store the hashed password
		Fname:            stringPtr(randomString(6)),
		Surname:          stringPtr(randomString(7)),
		Gender:           stringPtr(genders[rand.Intn(len(genders))]),
		Birthdate:        &birthdate,
		About:            stringPtr("I love " + selectedInterests[rand.Intn(len(selectedInterests))]),
		Hobbies:          &selectedHobbies,
		Interests:        &selectedInterests,
		Country:          &country,
		City:             &city,
		LookingForGender: stringPtr("any"),
		LookingForMinAge: intPtr(18),
		LookingForMaxAge: intPtr(50),
		Picture:          stringPtr("https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png"),
	}
}

// Admin API to insert 100+ fake users
func LoadFictitiousUsers(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Secret") != adminSecret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	numUsers := 100
	users := make([]models.User, numUsers)

	for i := range users {
		users[i] = generateFakeUser()
	}

	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	for _, user := range users {
		_, err := tx.Exec(context.Background(), `
			INSERT INTO users (email, password, fname, surname, gender, birthdate, about, hobbies, interests, country, city, looking_for_gender, looking_for_min_age, looking_for_max_age, profile_picture_url)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10, $11, $12, $13, $14, $15)`,
			user.Email, user.Password, user.Fname, user.Surname, user.Gender, user.Birthdate,
			user.About, toJSON(user.Hobbies), toJSON(user.Interests), user.Country, user.City,
			user.LookingForGender, user.LookingForMinAge, user.LookingForMaxAge, user.Picture,
		)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Fake users loaded successfully"})
}

// drops all users from the db
func ResetDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Admin-Secret") != adminSecret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := db.Pool.Exec(context.Background(), `DELETE FROM users`)
	if err != nil {
		http.Error(w, "Failed to reset database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Database reset successfully"})
}

func stringPtr(s string) *string { return &s }
func intPtr(i int) *int          { return &i }
func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}
func randomString(n int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

// randomly selects between min & max items from a slice
func randomSelection(slice []string, min, max int) []string {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(max-min+1) + min
	rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
	return slice[:n]
}
