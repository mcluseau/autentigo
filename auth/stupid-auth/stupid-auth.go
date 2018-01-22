package stupidauth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mcluseau/autorizo/api"
)

func New() api.Authenticator {
	return stupidAuth{}
}

type stupidAuth struct{}

var _ api.Authenticator = stupidAuth{}

func (sa stupidAuth) Authenticate(user, password string) (jwt.Claims, error) {
	return jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		Subject:   user,
	}, nil
}
