package auth

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// ExtraClaims are our standard extensions to JWT tokens.
type ExtraClaims struct {
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
	Groups        []string `json:"groups,omitempty"`
}

// Claims supporting our ExtraClaims.
type Claims struct {
	jwt.StandardClaims
	ExtraClaims
}
