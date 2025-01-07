package users

import (
	"github.com/google/uuid"
)

// UserService contains business logic related to users.
type UserService struct {
	UserRepo UserRepository
}

// NewUserService creates and returns a new instance of UserService.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{UserRepo: repo}
}

func (s *UserService) GetAllUsers() ([]*User, error) {
	return s.UserRepo.FindAll() // Assuming the repository has a method to fetch all users
}

// CreateUser handles the creation of a new user.
func (s *UserService) CreateUser(name, email string) (*User, error) {
	// Create a new user instance
	user := &User{
		ID:    uuid.New(), // Generate a new UUID for the user
		Name:  name,
		Email: email,
	}

	// Save the user using the repository
	createdUser, err := s.UserRepo.Save(user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// GetUserByID handles fetching a user by ID.
func (s *UserService) GetUserByID(id uuid.UUID) (*User, error) {
	return s.UserRepo.FindByID(id)
}
