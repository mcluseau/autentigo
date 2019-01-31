package api

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

var (
	// ErrInvalidAuthentication indicates an invalid authentication
	ErrInvalidAuthentication = errors.New("invalid authentication")
)

// Authenticator is the interface for authn backends
type Authenticator interface {
	Authenticate(user, password string, expiresAt time.Time) (claims jwt.Claims, err error)
}

// API registering with restful
type API struct {
	CRTData       []byte
	Authenticator Authenticator
	PublicKey     interface{}
	PrivateKey    interface{}
	SigningMethod jwt.SigningMethod
	TokenDuration time.Duration
}

// Register provide a restful.WebService from this API
func (api *API) Register() *restful.WebService {
	ws := &restful.WebService{}
	api.registerBasic(ws)
	api.registerSimple(ws)
	api.registerKeystone(ws)
	api.registerK8sAuthenticator(ws)
	api.registerCertificate(ws)
	return ws
}
