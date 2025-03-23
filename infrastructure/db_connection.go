package infrastructure

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NewDBConnection() (*sql.DB, error) {
	// Get database connection details from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Check if environment variables are empty and log an error
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatalf("Missing required database environment variables: DB_USER=%s, DB_PASSWORD=%s, DB_HOST=%s, DB_PORT=%s, DB_NAME=%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	// Create the DSN (Data Source Name) for the database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	var db *sql.DB
	var err error

	// Retry connection until successful or tries exceed 10
	for tries := 1; ; tries++ {
		// Initialize the database connection
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Error opening database: %v", err)
		} else {
			// Attempt to ping the database
			err = db.Ping()
			if err == nil {
				log.Println("Connected to the database!")
				break
			}
		}

		if tries >= 10 {
			log.Fatalf("Failed to connect to the database after %d tries", tries)
			return nil, err
		}

		// Wait before retrying
		log.Println("Waiting for database to be ready...")
		time.Sleep(2 * time.Second) // Retry every 2 seconds
	}

	return db, nil
}
