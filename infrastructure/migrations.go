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

// runMigrations runs database migrations based on the specified action (up, down, force).
func runMigrations(db *sql.DB, action string, version string) {
	// Debug: Print current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	log.Printf("Current working directory: %s", wd)

	// Debug: Print migration path
	migrationPath := "file://app/infrastructure/migrations"
	log.Printf("Migration path: %s", migrationPath)

	// Create a MySQL driver instance from the provided *sql.DB
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Failed to create MySQL driver instance: %v", err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrate: %v", err)
	}

	// Perform the migration action
	switch action {
	case "up":
		log.Println("Applying migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
		log.Println("Migrations applied successfully.")
	case "down":
		log.Println("Rolling back migrations...")
		if err := m.Down(); err != nil {
			log.Fatalf("Failed to roll back migrations: %v", err)
		}
		log.Println("Migrations rolled back successfully.")
	case "force":
		if version == "" {
			log.Fatalf("You must specify a version for the force action.")
		}
		log.Printf("Forcing migrations to version %s...", version)
		if err := m.Force(parseVersion(version)); err != nil {
			log.Fatalf("Failed to force migrations to version %s: %v", version, err)
		}
		log.Println("Forced migrations to specified version.")
	default:
		log.Fatalf("Unknown migration action: %s", action)
	}
}

// parseVersion converts a version string to an integer.
func parseVersion(version string) int {
	var v int
	if _, err := fmt.Sscanf(version, "%d", &v); err != nil {
		log.Fatalf("Invalid version format: %v", err)
	}
	return v
}

// MigrationsCliArguments handles command-line arguments for running migrations.
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
