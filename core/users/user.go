package users

import "github.com/google/uuid"

// User represents a user entity.
type User struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   string    `json:"password,omitempty"`
	RootUserID uuid.UUID `json:"root_user_id,omitempty"`
}
