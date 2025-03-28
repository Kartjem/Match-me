package handlers

import (
	"context"
	"encoding/json"
	"log"
	"matchme-backend/internal/db"
	"matchme-backend/internal/models"
	"matchme-backend/internal/utils"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxRecommendations = 10
	minScoreThreshold  = 8.0
)

func RecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr, err := utils.ExtractUserIDFromTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	// 1. get the viewer's profile data
	viewer, err := getUserByID(userID)
	if err != nil {
		log.Printf("Error fetching viewer user: %v\n", err)
		http.Error(w, "Error retrieving viewer data", http.StatusInternalServerError)
		return
	}

	// 2. fetch potential matches
	//    skipping those the viewer has dismissed or is the same user
	dismissedIDs, err := loadDismissedIDs(userID)
	if err != nil {
		log.Printf("Error loading dismissed IDs: %v\n", err)
		http.Error(w, "Error retrieving matches", http.StatusInternalServerError)
		return
	}

	potential, err := fetchPotentialMatches(userID, dismissedIDs)
	if err != nil {
		log.Printf("Error fetching potential matches: %v\n", err)
		http.Error(w, "Error retrieving matches", http.StatusInternalServerError)
		return
	}

	// 3. score them
	var scored []userWithScore
	for _, m := range potential {
		s, skip := computeMatchScore(viewer, m)
		if skip {
			continue
		}
		if s > 0 {
			scored = append(scored, userWithScore{
				ID:    m.UserID,
				Score: s,
			})
		}
	}

	// 4. sort by descending score (best first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// 5. take top N recommendations
	if len(scored) > maxRecommendations {
		scored = scored[:maxRecommendations]
	}

	// 6. return just the IDs
	resultIDs := make([]int, len(scored))
	for i, sc := range scored {
		resultIDs[i] = sc.ID
	}

	// after computing matches, insert them into the database
	for _, match := range scored {
		_, err := db.Pool.Exec(context.Background(), `
        INSERT INTO recommendations (user_id, recommended_user_id, score)
        VALUES ($1, $2, $3)
        ON CONFLICT DO NOTHING
    `, userID, match.ID, match.Score)
		if err != nil {
			log.Printf("Failed to save recommendation for %d -> %d: %v", userID, match.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resultIDs)
}

func DismissRecommendationHandler(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		DismissedUserID int `json:"dismissedUserId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(), `
        INSERT INTO dismissed_recommendations (user_id, dismissed_user_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, userID, body.DismissedUserID)
	if err != nil {
		log.Printf("Error dismissing recommendation: %v\n", err)
		http.Error(w, "Error dismissing recommendation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Recommendation dismissed successfully",
	})
}

// returns all users with a complete profile that match the viewer's location
func fetchPotentialMatches(viewerID int, dismissed map[int]bool) ([]models.User, error) {
	rows, err := db.Pool.Query(context.Background(), `
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
            profile_picture_url
        FROM users
        WHERE id <> $1
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
          -- Only return users in the same country and city as the viewer
          AND country = (SELECT country FROM users WHERE id = $1)
          AND city = (SELECT city FROM users WHERE id = $1)
          AND (
              id NOT IN (
                  SELECT connected_user_id FROM connections WHERE user_id = $1
                  UNION
                  SELECT user_id FROM connections WHERE connected_user_id = $1
                  UNION
                  SELECT dismissed_user_id FROM dismissed_recommendations WHERE user_id = $1
              )
              OR id IN (
                  SELECT user_id FROM connections WHERE connected_user_id = $1 AND status = 'pending'
              )
          )
    `, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, err
		}
		if dismissed[u.UserID] {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func loadDismissedIDs(viewerID int) (map[int]bool, error) {
	rows, err := db.Pool.Query(context.Background(), `
        SELECT dismissed_user_id
        FROM dismissed_recommendations
        WHERE user_id = $1
    `, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dismissed := make(map[int]bool)
	for rows.Next() {
		var duid int
		if err := rows.Scan(&duid); err == nil {
			dismissed[duid] = true
		}
	}
	return dismissed, nil
}

type userWithScore struct {
	ID    int
	Score float64
}

// computes a matching score between the viewer and the target
// it considers basic location, age, and gender requirements, as well as matching hobbies and interests
// applying a multiplier for items that the viewer has marked as preferred
// finally, if the computed score is below minScoreThreshold, the candidate is skipped
func computeMatchScore(viewer models.User, target models.User) (float64, bool) {
	var score float64

	// 1) require same country
	if viewer.Country == nil || target.Country == nil || !strings.EqualFold(*viewer.Country, *target.Country) {
		return 0, true
	}
	score += 1

	// 1b) require same city
	if viewer.City == nil || target.City == nil || !strings.EqualFold(*viewer.City, *target.City) {
		return 0, true
	}
	score += 2

	// 2) skip if target’s age is outside viewer's min–max range
	targetAge := calcAge(*target.Birthdate)
	if viewer.LookingForMinAge == nil || viewer.LookingForMaxAge == nil {
		return 0, true
	}
	if targetAge < *viewer.LookingForMinAge || targetAge > *viewer.LookingForMaxAge {
		return 0, true
	}
	score += 2

	// 3) gender check. If viewer is looking for a specific gender, require it
	if viewer.LookingForGender != nil && strings.ToLower(*viewer.LookingForGender) != "any" {
		if target.Gender == nil || !strings.EqualFold(*viewer.LookingForGender, *target.Gender) {
			return 0, true
		}
		score += 2
	}

	// 4) hobbies: for each matching hobby, if it is also preferred by the viewer, count double
	var hobbyScore float64
	if viewer.Hobbies != nil && target.Hobbies != nil {
		for _, tHobby := range *target.Hobbies {
			for _, vHobby := range *viewer.Hobbies {
				if strings.EqualFold(tHobby, vHobby) {
					multiplier := 1.0
					if viewer.PreferredHobbies != nil {
						for _, pref := range *viewer.PreferredHobbies {
							if strings.EqualFold(pref, tHobby) {
								multiplier = 2.0
								break
							}
						}
					}
					hobbyScore += multiplier
					break
				}
			}
		}
	}
	score += hobbyScore

	// 5) interests: similar logic with multiplier
	var interestScore float64
	if viewer.Interests != nil && target.Interests != nil {
		for _, tInterest := range *target.Interests {
			for _, vInterest := range *viewer.Interests {
				if strings.EqualFold(tInterest, vInterest) {
					multiplier := 1.0
					if viewer.PreferredInterests != nil {
						for _, pref := range *viewer.PreferredInterests {
							if strings.EqualFold(pref, tInterest) {
								multiplier = 2.0
								break
							}
						}
					}
					interestScore += multiplier
					break
				}
			}
		}
	}
	score += interestScore

	// enforce a minimal score threshold
	if score < minScoreThreshold {
		return score, true
	}

	return score, false
}

// returns how many items match ignoring case
func intersectionLen(a, b []string) int {
	set := make(map[string]bool)
	for _, item := range a {
		set[strings.ToLower(item)] = true
	}
	count := 0
	for _, item := range b {
		if set[strings.ToLower(item)] {
			count++
		}
	}
	return count
}

func calcAge(birthdate string) int {
	t, err := time.Parse("2006-01-02", birthdate)
	if err != nil {
		return 0
	}
	now := time.Now()
	years := now.Year() - t.Year()
	if (now.Month() < t.Month()) || (now.Month() == t.Month() && now.Day() < t.Day()) {
		years--
	}
	if years < 0 {
		years = 0
	}
	return years
}
