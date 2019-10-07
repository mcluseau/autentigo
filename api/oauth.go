package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	restful "github.com/emicklei/go-restful"
)

const bearerPrefix = "Bearer "

func (api *API) registerOauth(ws *restful.WebService) {
	ws.
		Route(ws.GET("/oauth/{provider}").
			To(api.oauthAuthenticate).
			Doc("Authenticate using oauth token").
			Param(restful.HeaderParameter("Authorization", "Oauth authorization header")).
			Produces("application/json").
			Writes(AuthResponse{}))
}

func (api *API) oauthAuthenticate(request *restful.Request, response *restful.Response) {
	defer func() {
		if err := recover(); err != nil {
			// unhandled error
			WriteError(err.(error), response)
		}
	}()

	authHeader := request.HeaderParameter("Authorization")
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		response.WriteErrorString(http.StatusUnauthorized, "missing bearer prefix")
		return
	}

	accessToken := authHeader[len(bearerPrefix):]
	provider := request.PathParameter("provider")

	baseURL, err := oauthClientIdentityURL(provider)
	if err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	identityResponse, err := http.Get(baseURL + "?access_token=" + accessToken)
	if err != nil {
		response.WriteError(http.StatusUnauthorized, fmt.Errorf("failed getting client identity by oauth: %s", err.Error()))
		return
	}

	defer identityResponse.Body.Close()
	contents, err := ioutil.ReadAll(identityResponse.Body)
	if err != nil {
		response.WriteError(http.StatusUnprocessableEntity, fmt.Errorf("failed reading response body: %s", err.Error()))
		return
	}

	var clientIdentity map[string]interface{}
	if err := json.Unmarshal(contents, &clientIdentity); err != nil {
		response.WriteError(http.StatusUnprocessableEntity, fmt.Errorf("failed unmarshalling contents: %s", err.Error()))
		return
	}

	id := stringValue(clientIdentity, "id")
	if len(id) == 0 {
		id = stringValue(clientIdentity, "sub") // different between some oauth providers
		if len(id) == 0 {
			response.WriteErrorString(http.StatusUnprocessableEntity, "client identity given by oauth is unprocessable")
			return
		}
	}

	exp := time.Now().Add(api.TokenDuration)
	user, claims, err := api.Authenticator.FindUser(id, provider, exp)
	if err != nil {
		response.WriteError(http.StatusUnprocessableEntity, fmt.Errorf("associated user not found: %s", err.Error()))
		return
	}
	_, tokenString, err := api.createToken(user, claims)
	if err != nil {
		response.WriteError(http.StatusUnprocessableEntity, fmt.Errorf("Not possible to create valid token: %s", err.Error()))
		return
	}

	_, err = api.checkToken(tokenString)
	if err != nil {
		response.WriteError(http.StatusUnprocessableEntity, fmt.Errorf("Not valid token after creation: %s", err.Error()))
		return
	}

	response.WriteEntity(&AuthResponse{tokenString, claims})
}

func stringValue(m map[string]interface{}, field string) string {
	v, ok := m[field]
	if !ok {
		return ""
	}
	return v.(string)
}

func oauthClientIdentityURL(provider string) (value string, err error) {
	urlEnv := strings.ToUpper(provider) + "_USERIDENTITYURL"
	value = os.Getenv(urlEnv)
	if len(value) == 0 {
		err = fmt.Errorf("client identity url given by provider %s is missing, please verify autentigo configuration [%s]", provider, urlEnv)
	}
	return
}
