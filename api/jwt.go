package api

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/mcluseau/autentigo/auth"
)

func (api *API) createToken(user string, claims jwt.Claims) (*jwt.Token, string, error) {
	token := jwt.NewWithClaims(api.SigningMethod, claims)
	signed, err := token.SignedString(api.PrivateKey)
	return token, signed, err
}

func (api *API) keyfunc(t *jwt.Token) (interface{}, error) {
	return api.PublicKey, nil
}

func (api *API) checkToken(tokenString string) (*auth.Claims, error) {
	claims := &auth.Claims{}

	if _, err := jwt.ParseWithClaims(tokenString, claims, api.keyfunc); err != nil {
		return nil, err
	}

	if err := claims.Valid(); err != nil {
		return nil, err
	}

	return claims, nil
}

func (api *API) authenticate(user, password string) (jwt.Claims, error) {
	exp := time.Now().Add(api.TokenDuration)
	return api.Authenticator.Authenticate(user, password, exp)
}
