package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB() {
	dsn := "postgres://postgres:58247774n@localhost:5432/match_me"
	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	fmt.Println("Connected to the database successfully.")
}

func CloseDB() {
	Pool.Close()
}

func ApplyMigrations() {
	files, err := filepath.Glob("internal/db/migrations/*.sql")
	if err != nil {
		log.Fatalf("Failed to read migrations: %v\n", err)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v\n", file, err)
		}

		_, err = Pool.Exec(context.Background(), string(content))
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v\n", file, err)
		}

		log.Printf("Applied migration: %s\n", file)
	}
}
