package users

import (
	"errors"
	"fmt" // import the fmt package

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService contains business logic related to users.
type UserService struct {
	UserRepo UserRepository
}

// NewUserService creates and returns a new instance of UserService.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{UserRepo: repo}
}

// GetAllUsers retrieves all users from the database.
func (s *UserService) GetAllUsers() ([]*User, error) {
	return s.UserRepo.FindAll()
}

// CreateUser handles the creation of a new user.
func (s *UserService) CreateUser(name, email, password string) (*User, error) {
	// Hash the password before saving it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create a new user instance
	user := &User{
		ID:       uuid.New(), // Generate a new UUID for the user
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	// Save the user using the repository
	createdUser, err := s.UserRepo.Save(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return createdUser, nil
}

// CreateSubUser handles the creation of a new sub-user.
func (s *UserService) CreateSubUser(rootUserID uuid.UUID, name, email, password string) (*User, error) {
	// Hash the password before saving it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create a new sub-user instance
	subUser := &User{
		ID:       uuid.New(), // Generate a new UUID for the sub-user
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		// RootUserID:  rootUserID, // Link to the root user
	}

	// Save the sub-user using the repository
	createdUser, err := s.UserRepo.SaveSubUser(subUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-user: %v", err)
	}

	return createdUser, nil
}

// GetUserByID handles fetching a user by ID.
func (s *UserService) GetUserByID(id uuid.UUID) (*User, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user by ID: %v", err)
	}
	return user, nil
}

// GetUserByEmail fetches a user by email.
func (s *UserService) GetUserByEmail(email string) (*User, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user by email: %v", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
