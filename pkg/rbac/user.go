package rbac

import (
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mcluseau/autentigo/client"
)

// GroupsFromToken returns the groups claimed by the token
func GroupsFromToken(token *jwt.Token) (groups []string) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if claims == nil || !ok {
		return
	}

	tokenGroups, ok := claims["groups"].([]interface{})
	if tokenGroups == nil || !ok {
		return
	}

	groups = make([]string, 0, len(tokenGroups))
	for _, group := range tokenGroups {
		g, ok := group.(string)
		if !ok {
			// anything wrong is bad
			return nil
		}

		groups = append(groups, g)
	}

	return
}

// UserFromToken returns a User object from the given token.
func UserFromToken(token *jwt.Token) (u *User) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if claims == nil || !ok {
		return
	}

	name, ok := claims["sub"].(string)

	if !ok {
		return
	}

	return &User{
		Name:   name,
		Groups: GroupsFromToken(token),
	}
}

const bearerPrefix = "Bearer "

// UserFromRequest returns a User object from the given request or `nil` if
// the token is not found or invalid.
func UserFromRequest(req *http.Request, validationCrt []byte) (u *User) {
	authHeader := req.Header.Get("Authorization")

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return
	}

	tokenStr := authHeader[len(bearerPrefix):]

	token, err := client.Parse(validationCrt, tokenStr)
	if err != nil {
		return
	}

	return UserFromToken(token)
}
