package api

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

// Keystone API (like) auth request
type KeystoneAuthReq struct {
	Auth *KeystoneAuth `json:"auth"`
}

type KeystoneAuth struct {
	Identity struct {
		Methods  []string `json:"methods"`
		Password struct {
			User struct {
				Id       string `json:"id,omitempty"`
				Name     string `json:"name,omitempty"`
				Password string `json:"password"`
				Domain   struct {
					Id   string `json:"id,omitempty"`
					Name string `json:"name,omitempty"`
				} `json:"domain"`
			} `json:"user"`
		} `json:"password"`
	} `json:"identity"`
}

type KeystoneAuthResponse struct {
	Token struct {
		IssuedAt  time.Time `json:"issued_at"`
		ExpiresAt time.Time `json:"expires_at"`
		User      struct {
			Id   string `json:"id,omitempty"`
			Name string `json:"name,omitempty"`
		} `json:"user"`
	} `json:"token"`
}

func (api *API) registerKeystone(ws *restful.WebService) {
	path := "/v3/auth/tokens"

	ws.
		Route(ws.POST(path).
			To(api.keystoneAuthenticate).
			Doc("Authenticate using a Keystone-style request").
			Reads(KeystoneAuthReq{}).
			Writes(KeystoneAuthResponse{}))

	ws.
		Route(ws.GET(path).
			To(api.keystoneShow).
			Doc("Validates and shows information for a token").
			Param(restful.HeaderParameter(
				"X-Auth-Token", "A valid authentication token for an administrative user.")).
			Param(restful.HeaderParameter(
				"X-Subject-Token", "The authentication token.")).
			Writes(KeystoneAuthResponse{}))
}

func (api *API) keystoneAuthenticate(request *restful.Request, response *restful.Response) {
	defer func() {
		if err := recover(); err != nil {
			// unhandled error
			writeError(err.(error), response)
		}
	}()

	authReq := KeystoneAuthReq{}
	if err := request.ReadEntity(&authReq); err != nil {
		writeError(err, response)
		return
	}
	if authReq.Auth == nil {
		response.WriteErrorString(http.StatusUnauthorized, "No authentication provided")
		return
	}

	user := authReq.Auth.Identity.Password.User
	login := user.Id
	if login == "" {
		login = user.Name
	}

	claims, err := api.Authenticator.Authenticate(login, user.Password)
	if err == ErrInvalidAuthentication {
		response.WriteErrorString(http.StatusUnauthorized, "Authentication failed")
		return
	} else if err != nil {
		panic(err)
	}

	_, tokenString, err := api.createToken(login, claims)

	if err != nil {
		panic(err)
	}

	stdClaims, err := api.checkToken(tokenString)
	if err != nil {
		panic(err)
	}

	authResp := newKeystoneAuthRespFromClaims(stdClaims)

	response.Header().Set("X-Subject-Token", tokenString)
	response.WriteHeaderAndEntity(http.StatusCreated, authResp)
}

func newKeystoneAuthRespFromClaims(claims *jwt.StandardClaims) *KeystoneAuthResponse {
	authResp := &KeystoneAuthResponse{}
	authResp.Token.IssuedAt = time.Unix(claims.IssuedAt, 0)
	authResp.Token.ExpiresAt = time.Unix(claims.ExpiresAt, 0)
	authResp.Token.User.Id = claims.Subject
	authResp.Token.User.Name = claims.Subject
	return authResp
}

func (api *API) keystoneCheck(request *restful.Request, response *restful.Response) {
	api.keystoneCheckClaims(request, response)
}

func (api *API) keystoneShow(request *restful.Request, response *restful.Response) {
	claims := api.keystoneCheckClaims(request, response)

	if claims == nil {
		return
	}

	response.WriteEntity(newKeystoneAuthRespFromClaims(claims))
}

// return nil iff check fails (response already filled)
func (api *API) keystoneCheckClaims(request *restful.Request, response *restful.Response) *jwt.StandardClaims {
	authToken := request.HeaderParameter("X-Auth-Token")
	if _, err := api.checkToken(authToken); err != nil {
		response.WriteError(http.StatusUnauthorized, err)
		return nil
	}

	subjectToken := request.HeaderParameter("X-Subject-Token")
	claims, err := api.checkToken(subjectToken)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return nil
	}

	response.WriteHeader(http.StatusOK)
	return claims
}
