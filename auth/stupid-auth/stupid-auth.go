package stupidauth

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mcluseau/autentigo/api"
)

// New Authenticator with no backend
func New() api.Authenticator {
	return stupidAuth{}
}

type stupidAuth struct{}

var _ api.Authenticator = stupidAuth{}

func (sa stupidAuth) Authenticate(user, password string, expiresAt time.Time) (jwt.Claims, error) {
	return jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: expiresAt.Unix(),
		Subject:   user,
	}, nil
}

func (sa stupidAuth) FindUser(clientID, provider string, expiresAt time.Time) (userID string, claims jwt.Claims, err error) {
	err = errors.New("inconsistent with stupid auth")
	return
}
