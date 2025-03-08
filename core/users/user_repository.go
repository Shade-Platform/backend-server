package users

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	"github.com/google/uuid"
)

// UserRepository defines methods for interacting with the data store.
type UserRepository interface {
	Save(user *User) (*User, error)       // Save a user to the database
	FindByID(id uuid.UUID) (*User, error) // Retrieve a user by UUID
	FindAll() ([]*User, error)
	FindByEmail(email string) (*User, error)
	// Additional methods for data access (e.g., Update, Delete, etc.)
}

// FindByEmail retrieves a user by their email address
func (repo *MySQLUserRepository) FindByEmail(email string) (*User, error) {
	// Define the query
	query := `SELECT id, name, email, password FROM users WHERE email = ?`

	// Create a User struct to hold the result
	var user User

	// Execute the query and scan the result into the user struct
	err := repo.DB.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil if no user is found
		}
		return nil, fmt.Errorf("failed to find user by email: %v", err)
	}

	return &user, nil
}

// MySQLUserRepository is the implementation of UserRepository using MySQL.
type MySQLUserRepository struct {
	DB *sql.DB
}

// NewMySQLUserRepository creates a new MySQLUserRepository
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{DB: db}
}

// Save stores a user in the database
func (repo *MySQLUserRepository) Save(user *User) (*User, error) {
	// If the user doesn't have an ID, generate a new UUID
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Prepare the query to insert the user into the database
	query := `INSERT INTO users (id, name, email, password) 
			  VALUES (?, ?, ?, ?) 
			  ON DUPLICATE KEY UPDATE name=?, email=?, password=?`

	// Execute the query
	_, err := repo.DB.Exec(query, user.ID, user.Name, user.Email, user.Password, user.Name, user.Email, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	return user, nil
}

func (repo *MySQLUserRepository) FindAll() ([]*User, error) {
	// Define the query to fetch all users
	query := `SELECT id, name, email FROM users`

	// Execute the query
	rows, err := repo.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %v", err)
	}
	defer rows.Close()

	var users []*User

	// Loop through the rows and scan the data into User structs
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, &user)
	}

	// Check for errors that occurred during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while reading rows: %v", err)
	}

	return users, nil
}

// FindByID retrieves a user by ID from the database
func (repo *MySQLUserRepository) FindByID(id uuid.UUID) (*User, error) {
	// Query the database for the user by ID
	query := `SELECT id, name, email, password, created_at FROM users WHERE id = ?`

	row := repo.DB.QueryRow(query, id)

	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %v", err)
	}

	return &user, nil
}
