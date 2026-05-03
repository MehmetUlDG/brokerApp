package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "broker_user"
		}
		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "broker_secret"
		}
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "broker_db"
		}
		sslmode := os.Getenv("DB_SSL_MODE")
		if sslmode == "" {
			sslmode = "disable"
		}
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	sqlBytes, err := os.ReadFile("migrations/000004_create_transactions_table.up.sql")
	if err != nil {
		log.Fatalf("Failed to read migration: %v", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	fmt.Println("Migration applied successfully!")
}
