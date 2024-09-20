package userService

import (
	"time"

	"github.com/google/uuid"
	"github.com/sunikka/clich-backend/internal/database"
)

type UserService struct {
	listenAddr string
	storage    *database.Queries
}

// Separate struct for user, with the sensitive information stripped out
type userRes struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Admin     bool      `json:"admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserUpdateReq struct {
	Username string `json:"username,omitempty"`
	// Password string `json:"password,omitempty"`
	// Email string `json:"email,omitempty"`
}
