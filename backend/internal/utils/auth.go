package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"matchme-backend/internal/db"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("your_secret_key")

// creates a JWT with userID and email claims
func GenerateToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"userID": userID,
		"email":  email,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// parses a token string and returns the userID claim
func ExtractUserIDFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		switch v := claims["userID"].(type) {
		case float64:
			return fmt.Sprintf("%d", int(v)), nil
		case string:
			return v, nil
		default:
			return "", errors.New("invalid userID type in token")
		}
	}
	return "", errors.New("invalid token")
}

// retrieves the token from a cookie
func ExtractUserIDFromTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	return ExtractUserIDFromToken(cookie.Value)
}

// returns the bcrypt hash of the plaintext password
func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// compares a bcrypt hashed password with its possible plaintext equivalent
func ComparePassword(hashedPassword, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plain))
}

// checks if a user's profile is complete
func IsProfileComplete(userID int) (bool, error) {
	var count int
	err := db.Pool.QueryRow(context.Background(), `
        SELECT COUNT(*)
        FROM users
        WHERE id = $1
        AND fname IS NOT NULL
        AND surname IS NOT NULL
        AND gender IS NOT NULL
        AND birthdate IS NOT NULL
        AND about IS NOT NULL
        AND hobbies IS NOT NULL
        AND interests IS NOT NULL
        AND country IS NOT NULL
        AND city IS NOT NULL
        AND looking_for_gender IS NOT NULL
        AND looking_for_min_age IS NOT NULL
        AND looking_for_max_age IS NOT NULL
    `, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
