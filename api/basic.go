package api

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
)

func (api *API) registerBasic(ws *restful.WebService) {
	ws.
		Route(ws.GET("/basic").
			To(api.basicAuthenticate).
			Doc("Authenticate using HTTP basic auth").
			Param(restful.HeaderParameter(
				"Authorization", "Basic authorization header")).
			Param(setCookieHeader()).
			Param(setCookieDomainHeader()).
			Param(setCookieInsecureHeader()).
			Produces("application/json").
			Writes(AuthResponse{}))
}

func (api *API) basicAuthenticate(request *restful.Request, response *restful.Response) {
	defer func() {
		if err := recover(); err != nil {
			// unhandled error
			WriteError(err.(error), response)
		}
	}()

	user, password, ok := request.Request.BasicAuth()
	if !ok {
		response.Header().Set("WWW-Authenticate", `Basic realm="Autorizo"`)
		response.WriteErrorString(http.StatusUnauthorized, "Unauthorized.\n")
		return
	}

	api.writeAuthResponse(request, response, user, password)
}
