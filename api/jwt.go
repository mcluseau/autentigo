package api

import (
	"github.com/dgrijalva/jwt-go"
)

func (api *API) createToken(user string, claims jwt.Claims) (*jwt.Token, string, error) {
	token := jwt.NewWithClaims(api.SigningMethod, claims)
	signed, err := token.SignedString(api.PrivateKey)
	return token, signed, err
}
