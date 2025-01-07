package infrastructure

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(db *sql.DB, action string, version string) {
	// Create a MySQL driver instance from the provided *sql.DB
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Failed to create MySQL driver instance: %v", err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://infrastructure/migrations", // Corrected path to migration files
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}

	// Perform the migration action
	switch action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		log.Println("Migrations applied successfully.")
	case "down":
		if err := m.Down(); err != nil {
			log.Fatalf("Failed to roll back migrations: %v", err)
		}
		log.Println("Migrations rolled back successfully.")
	case "force":
		if version == "" {
			log.Fatalf("You must specify a version for the force action.")
		}
		if err := m.Force(parseVersion(version)); err != nil {
			log.Fatalf("Failed to force migrations to version %s: %v", version, err)
		}
		log.Println("Forced migrations to specified version.")
	default:
		log.Fatalf("Unknown migration action: %s", action)
	}
}

func parseVersion(version string) int {
	var v int
	if _, err := fmt.Sscanf(version, "%d", &v); err != nil {
		log.Fatalf("Invalid version format: %v", err)
	}
	return v
}

func MigrationsCliArguments(dbConn *sql.DB) {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			action := "up" // Default action
			if len(os.Args) > 2 {
				action = os.Args[2] // e.g., "down", "force"
			}
			version := ""
			if action == "force" && len(os.Args) > 3 {
				version = os.Args[3]
			}

			// Run migrations
			runMigrations(dbConn, action, version)

			// Exit the application after migrations are done
			os.Exit(0)
		default:
			log.Fatalf("Unknown command: %s", os.Args[1])
		}
	}
}
