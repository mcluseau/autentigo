package auth

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// ExtraClaims are our standard extensions to JWT tokens.
type ExtraClaims struct {
	DisplayName   string   `json:"display_name,omitempty"`
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
	Groups        []string `json:"groups,omitempty"`
}

// Claims supporting our ExtraClaims.
type Claims struct {
	jwt.StandardClaims
	ExtraClaims
}
