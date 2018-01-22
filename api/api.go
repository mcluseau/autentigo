package api

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

var (
	ErrInvalidAuthentication = errors.New("invalid authentication")
)

type Authenticator interface {
	Authenticate(user, password string) (claims jwt.Claims, err error)
}

type API struct {
	Authenticator Authenticator
	PublicKey     interface{}
	PrivateKey    interface{}
	SigningMethod jwt.SigningMethod
}

func (api *API) Register() *restful.WebService {
	ws := &restful.WebService{}
	api.registerKeystone(ws)
	return ws
}
