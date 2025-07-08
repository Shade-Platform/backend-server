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
	Save(user *User) (*User, error)          // Save a user to the database
	SaveSubUser(user *User) (*User, error)   // Save a sub-user to the database
	FindByID(id uuid.UUID) (*User, error)    // Retrieve a user by UUID
	FindByEmail(email string) (*User, error) // Retrieve a user by email
	FindAll() (int, error)                   // Retrieve all users
}

// MySQLUserRepository is the implementation of UserRepository using MySQL.
type MySQLUserRepository struct {
	DB *sql.DB
}

// NewMySQLUserRepository creates a new MySQLUserRepository.
func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{DB: db}
}

// Save stores a user in the database.
func (repo *MySQLUserRepository) Save(user *User) (*User, error) {
	// Generate a new UUID if the user doesn't have an ID
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Prepare the query to insert or update the user
	query := `
	INSERT INTO users (id, name, email, password)
	VALUES (?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE 
		name = VALUES(name), 
		email = VALUES(email), 
		password = VALUES(password)
`

	// Execute the query
	_, err := repo.DB.Exec(
		query,
		// user.ID, user.Name, user.Email, user.Password, user.RootUserID,
		user.ID, user.Name, user.Email, user.Password,
		// user.Name, user.Email, user.Password, user.RootUserID,
		// user.Name, user.Email, user.Password,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %v", err)
	}

	return user, nil
}

// SaveSubUser stores a sub-user in the database.
func (repo *MySQLUserRepository) SaveSubUser(user *User) (*User, error) {
	// Generate a new UUID if the sub-user doesn't have an ID
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Prepare the query to insert the sub-user
	query := `
		INSERT INTO users (id, name, email, password, root_user_id) 
		VALUES (?, ?, ?, ?, ?)
	`

	// Execute the query
	// _, err := repo.DB.Exec(query, user.ID, user.Name, user.Email, user.Password, user.RootUserID)
	_, err := repo.DB.Exec(query, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to save sub-user: %v", err)
	}

	return user, nil
}

// FindByID retrieves a user by ID from the database.
func (repo *MySQLUserRepository) FindByID(id uuid.UUID) (*User, error) {
	// Prepare the query to fetch the user by ID
	query := `
		SELECT id, name, email, root_user_id 
		FROM users 
		WHERE id = ?
	`

	// Execute the query and scan the result into a User struct
	var user User
	// err := repo.DB.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.RootUserID)
	err := repo.DB.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.RootUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %v", err)
	}

	return &user, nil
}

// FindByEmail retrieves a user by their email address.
func (repo *MySQLUserRepository) FindByEmail(email string) (*User, error) {
	// Prepare the query to fetch the user by email
	query := `
		SELECT id, name, email, password
		FROM users 
		WHERE email = ?
	`

	// Execute the query and scan the result into a User struct
	var user User
	// err := repo.DB.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.RootUserID)
	err := repo.DB.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil if no user is found
		}
		return nil, fmt.Errorf("failed to find user by email: %v", err)
	}

	return &user, nil
}

// FindAll retrieves count of users from the database.
func (repo *MySQLUserRepository) FindAll() (int, error) {
	// Prepare the query to fetch all users
	query := `
		SELECT count(*)
		FROM users
	`

	// Execute the query
	rows, err := repo.DB.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch users: %v", err)
	}
	defer rows.Close()

	var users int
	rows.Next()
	err = rows.Scan(&users)

	if err != nil {
		return 0, err
	}

	// Loop through the rows and scan the data into User structs
	// var users []*User
	// for rows.Next() {
	// 	var user User
	// 	err := rows.Scan(&user.ID, &user.Name, &user.Email)
	// 	// err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.RootUserID)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to scan user: %v", err)
	// 	}
	// 	users = append(users, &user)
	// }

	// Check for errors that occurred during row iteration
	// if err := rows.Err(); err != nil {
	// 	return nil, fmt.Errorf("error occurred while reading rows: %v", err)
	// }

	return users, nil
}
