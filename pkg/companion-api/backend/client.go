package backend

import (
	"github.com/mcluseau/autentigo/auth"
)

// UserData is a simple user struct with paswordhash and claims
type UserData struct {
	PasswordHash string           `json:"password"`
	OauthToken   string           `json:"oauth_token"`
	ExtraClaims  auth.ExtraClaims `json:"claims"`
}

// Client is the interface for all backends clients
type Client interface {
	CreateUser(id string, user *UserData) error
	UpdateUser(id string, update func(user *UserData) error) error
	DeleteUser(id string) error
}
