package api

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

var (
	ErrInvalidAuthentication = errors.New("invalid authentication")
)

type Authenticator interface {
	Authenticate(user, password string, expiresAt time.Time) (claims jwt.Claims, err error)
}

type API struct {
	Authenticator Authenticator
	PublicKey     interface{}
	PrivateKey    interface{}
	SigningMethod jwt.SigningMethod
	TokenDuration time.Duration
}

func (api *API) Register() *restful.WebService {
	ws := &restful.WebService{}
	api.registerKeystone(ws)
	return ws
}
