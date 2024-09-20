package utils

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string
	HashedPw  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserInfo struct {
	ID     string
	Name   string
	Active bool
	// conn *websocket.Conn
	// token jwt.Token
}

type Message struct {
	SenderID uuid.UUID
	Username string
	Content  string
}
