package auth

import (
	"errors"
	"time"

	"shade_web_server/core/users"

	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key (keep this safe, ideally load from env)
var jwtSecret = []byte("your_secret_key")

// AuthService handles authentication logic.
type AuthService struct {
	UserService *users.UserService
}

// NewAuthService initializes AuthService.
func NewAuthService(userService *users.UserService) *AuthService {
	return &AuthService{UserService: userService}
}

// GenerateJWT generates a JWT token for a user.
func (s *AuthService) GenerateJWT(user *users.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// AuthenticateUser checks user credentials and returns a JWT token.
func (s *AuthService) AuthenticateUser(email string) (string, error) {
	// Check if user exists
	user, err := s.UserService.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Generate JWT for the authenticated user
	token, err := s.GenerateJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}
