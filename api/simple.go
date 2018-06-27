package api

import (
	"net/http"

	"github.com/emicklei/go-restful"
)

func (api *API) registerSimple(ws *restful.WebService) {
	ws.
		Route(ws.POST("/simple").
			To(api.simpleAuthenticate).
			Doc("Authenticate").
			Consumes("application/json").
			Produces("application/json").
			Param(setCookieHeader()).
			Param(setCookieDomainHeader()).
			Param(setCookieInsecureHeader()).
			Reads(AuthReq{}).
			Writes(AuthResponse{}))
}

type AuthReq struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (api *API) simpleAuthenticate(request *restful.Request, response *restful.Response) {
	defer func() {
		if err := recover(); err != nil {
			// unhandled error
			writeError(err.(error), response)
		}
	}()

	authReq := AuthReq{}
	if err := request.ReadEntity(&authReq); err != nil {
		writeError(err, response)
		return
	}

	if authReq.User == "" {
		response.WriteErrorString(http.StatusUnauthorized, "No user given.")
		return
	}

	if authReq.Password == "" {
		response.WriteErrorString(http.StatusUnauthorized, "No password given.")
		return
	}

	api.writeAuthResponse(request, response, authReq.User, authReq.Password)
}

func setCookieHeader() *restful.Parameter {
	return restful.HeaderParameter(
		"X-Set-Cookie", "Set the (HTTP only) cookie specified in this header, return an empty 201 Created response.")
}

func setCookieInsecureHeader() *restful.Parameter {
	return restful.HeaderParameter(
		"X-Set-Cookie-Insecure", "If set to \"yes\", the authorization cookie will not be secure.")
}

func setCookieDomainHeader() *restful.Parameter {
	return restful.HeaderParameter(
		"X-Set-Cookie-Domain", "The domain of the authorization cookie.")
}

func (api *API) writeAuthResponse(request *restful.Request, response *restful.Response, user, password string) {
	claims, err := api.authenticate(user, password)
	if err == ErrInvalidAuthentication {
		response.WriteErrorString(http.StatusUnauthorized, "Authentication failed.\n")
		return
	} else if err != nil {
		panic(err)
	}

	_, tokenString, err := api.createToken(user, claims)

	if err != nil {
		panic(err)
	}

	_, err = api.checkToken(tokenString)
	if err != nil {
		panic(err)
	}

	if cookieName := request.HeaderParameter("X-Set-Cookie"); cookieName != "" {
		// with only set the cookie
		isSecure := true
		if secureHeader := request.HeaderParameter("X-Set-Cookie-Insecure"); secureHeader == "yes" {
			isSecure = false
		}

		http.SetCookie(response.ResponseWriter, &http.Cookie{
			Domain:   request.HeaderParameter("X-Set-Cookie-Domain"),
			HttpOnly: true, // it's the whole point of that
			Secure:   isSecure,
			Name:     cookieName,
			Value:    tokenString,
		})

		response.WriteHeader(http.StatusCreated)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, &AuthResponse{tokenString})
}
