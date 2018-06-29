package auth

import (
	jwt "github.com/dgrijalva/jwt-go"
)

type ExtraClaims struct {
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
	Groups        []string `json:"groups,omitempty"`
}

type Claims struct {
	jwt.StandardClaims
	ExtraClaims
}
