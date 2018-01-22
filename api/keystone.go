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
			Operation("authenticate").
			Reads(KeystoneAuthReq{}).
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

	stdClaims := jwt.StandardClaims{}

	if _, err := jwt.ParseWithClaims(tokenString, &stdClaims, func(t *jwt.Token) (interface{}, error) {
		return api.PublicKey, nil
	}); err != nil {
		panic(err)
	}

	authResp := &KeystoneAuthResponse{}
	authResp.Token.IssuedAt = time.Unix(stdClaims.IssuedAt, 0)
	authResp.Token.ExpiresAt = time.Unix(stdClaims.ExpiresAt, 0)
	authResp.Token.User.Id = stdClaims.Subject
	authResp.Token.User.Name = stdClaims.Subject

	response.Header().Set("X-Subject-Token", tokenString)
	response.WriteHeaderAndEntity(http.StatusCreated, authResp)
}
