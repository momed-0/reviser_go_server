package inits

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

func DBInit() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}
	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}

	DB = db
	log.Println("PostgreSQL database initialized successfully")
}
