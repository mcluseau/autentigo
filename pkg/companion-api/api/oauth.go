package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	restful "github.com/emicklei/go-restful"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/oauth2"

	"github.com/mcluseau/autentigo/auth"
	"github.com/mcluseau/autentigo/pkg/companion-api/backend"
)

type State struct {
	State  string
	UserID string
}

func oauthConfig(provider string) *oauth2.Config {
	upperCaseProvider := strings.ToUpper(provider)

	return &oauth2.Config{
		RedirectURL:  requireEnv("OAUTH_APP_ENDPOINT", "This app endpoint") + "/oauth/" + provider + "/callback",
		ClientID:     requireEnv(upperCaseProvider+"_CLIENTID", "oauth id given by oauth client"),
		ClientSecret: requireEnv(upperCaseProvider+"_CLIENTSECRET", "oauth secret given by oauth client"),
		Scopes:       strings.Split(os.Getenv(upperCaseProvider+"_SCOPES"), ","),
		Endpoint: oauth2.Endpoint{
			AuthURL:  requireEnv(upperCaseProvider+"_AUTHURL", "url of the client auth provider"),
			TokenURL: requireEnv(upperCaseProvider+"_TOKENURL", "url of the client token provider"),
		},
	}
}

func oauthClientIdentityURL(provider string) string {
	upperCaseProvider := strings.ToUpper(provider)
	return requireEnv(upperCaseProvider+"_USERIDENTITYURL", "client identity url given by provider "+provider)
}

func oauthState() (state string) {
	state = os.Getenv("OAUTH_STATE")
	if len(state) == 0 {
		// If I were a state, I would be a fish
		state = "><((('>"
	}
	return
}

// Register provide a restful.WebService from this API
func (cApi *CompanionAPI) oauthWS() (ws *restful.WebService) {
	ws = &restful.WebService{}
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Path("/oauth")

	ws.
		Route(ws.GET("/{provider}/{user-id}").
			To(cApi.register).
			Param(ws.PathParameter("provider", "oauth client").DataType("string")).
			Doc("Create or update user with informations given by oauth"))

	ws.
		Route(ws.GET("/{provider}/callback").
			To(cApi.callback).
			Param(ws.PathParameter("provider", "oauth client").DataType("string")).
			Doc("Don't use it directly. Oauth register callback"))

	return ws
}

func (cApi *CompanionAPI) register(request *restful.Request, response *restful.Response) {
	provider := request.PathParameter("provider")
	userID := request.PathParameter("user-id")

	if len(userID) == 0 {
		id, err := uuid.NewV4()
		if err != nil {
			response.WriteError(http.StatusUnprocessableEntity, err)
			return
		}
		userID = id.String()
	}

	state, err := json.Marshal(State{oauthState(), userID})
	if err != nil {
		panic(err)
	}

	url := oauthConfig(provider).AuthCodeURL(string(state))
	http.Redirect(response.ResponseWriter, request.Request, url, http.StatusTemporaryRedirect)
}

func (cApi *CompanionAPI) callback(request *restful.Request, response *restful.Response) {
	provider := request.PathParameter("provider")
	code := request.Request.FormValue("code")

	state := &State{}
	if err := json.Unmarshal([]byte(request.Request.FormValue("state")), state); err != nil {
		response.WriteErrorString(http.StatusUnprocessableEntity, "invalid oauth state form")
		return
	}

	if state.State != oauthState() {
		response.WriteErrorString(http.StatusUnauthorized, "invalid oauth state")
		return
	}

	token, err := oauthConfig(provider).Exchange(oauth2.NoContext, code)
	if err != nil {
		response.WriteError(http.StatusUnauthorized, fmt.Errorf("code exchange failed: %s", err.Error()))
		return
	}
	if !token.Valid() {
		response.WriteErrorString(http.StatusUnauthorized, "token not valid")
		return
	}

	identityResponse, err := http.Get(oauthClientIdentityURL(provider) + "?access_token=" + token.AccessToken)
	if err != nil {
		response.WriteError(http.StatusUnauthorized, fmt.Errorf("failed getting client identity: %s", err.Error()))
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
	name := stringValue(clientIdentity, "name")
	email := stringValue(clientIdentity, "email")

	if len(id) == 0 {
		id = stringValue(clientIdentity, "sub") // different between some oauth providers
		if len(id) == 0 {
			response.WriteErrorString(http.StatusUnprocessableEntity, "client identity given by oauth is unprocessable")
			return
		}
	}

	b := &backend.UserData{
		ExtraClaims: auth.ExtraClaims{
			DisplayName:   name,
			Email:         email,
			EmailVerified: true,
		},
	}
	err = cApi.Client.CreateUser(state.UserID, b)

	if err == ErrUserAlreadyExist {
		err = cApi.Client.UpdateUser(state.UserID, func(user *backend.UserData) (_ error) {
			user.ExtraClaims.DisplayName = name
			if len(email) != 0 && email != user.ExtraClaims.Email {
				user.ExtraClaims.Email = email
				user.ExtraClaims.EmailVerified = true
			}
			return
		})
	}
	if err != nil {
		response.WriteError(http.StatusInternalServerError, fmt.Errorf("client identity cannot be upgraded in backend: %s", err.Error()))
		return
	}

	if err = cApi.Client.PutUserID(provider, id, state.UserID); err != nil {
		response.WriteError(http.StatusInternalServerError, fmt.Errorf("oauth information cannot be upgraded in backend: %s", err.Error()))
		return
	}

	response.AddHeader("Authorization", "Bearer "+token.AccessToken)
	response.WriteHeader(http.StatusOK)
}

func stringValue(m map[string]interface{}, field string) string {
	v, ok := m[field]
	if !ok {
		return ""
	}
	return v.(string)
}

func requireEnv(name, description string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatal("Env ", name, " is required: ", description)
	}
	return v
}
