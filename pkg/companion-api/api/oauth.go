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
	"github.com/mcluseau/autentigo/auth"
	"github.com/mcluseau/autentigo/pkg/companion-api/backend"
	"golang.org/x/oauth2"
)

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

func oauthUserInfosURL(provider string) string {
	upperCaseProvider := strings.ToUpper(provider)
	return requireEnv(upperCaseProvider+"_USERINFOSURL", "url of the client user informations provider")
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
		Route(ws.GET("/{provider}").
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
	url := oauthConfig(provider).AuthCodeURL(oauthState())
	http.Redirect(response.ResponseWriter, request.Request, url, http.StatusTemporaryRedirect)
}

type OAuthUserInfos struct {
	ID    string
	Name  string
	Email string
}

func (cApi *CompanionAPI) callback(request *restful.Request, response *restful.Response) {
	provider := request.PathParameter("provider")
	state := request.Request.FormValue("state")
	code := request.Request.FormValue("code")

	if state != oauthState() {
		response.WriteErrorString(http.StatusUnauthorized, "invalid oauth state")
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

	infoResp, err := http.Get(oauthUserInfosURL(provider) + "?fields=name,email&access_token=" + token.AccessToken)
	if err != nil {
		response.WriteError(http.StatusUnauthorized, fmt.Errorf("failed getting user info: %s", err.Error()))
		return
	}

	defer infoResp.Body.Close()
	contents, err := ioutil.ReadAll(infoResp.Body)
	if err != nil {
		response.WriteError(http.StatusUnprocessableEntity, fmt.Errorf("failed reading response body: %s", err.Error()))
		return
	}

	userInfos := OAuthUserInfos{}
	if err := json.Unmarshal(contents, &userInfos); err != nil {
		response.WriteError(http.StatusUnprocessableEntity, fmt.Errorf("failed unmarshalling contents: %s", err.Error()))
		return
	}

	if len(userInfos.ID) == 0 {
		response.WriteErrorString(http.StatusUnprocessableEntity, "user infos given by oauth are unprocessable")
		return
	}

	b := &backend.UserData{
		OauthTokens: []backend.OauthToken{backend.OauthToken{
			Provider: provider,
			Token:    token.AccessToken,
		}},
		ExtraClaims: auth.ExtraClaims{
			DisplayName:   userInfos.Name,
			Email:         userInfos.Email,
			EmailVerified: true,
		},
	}
	err = cApi.Client.CreateUser(userInfos.ID, b)

	if err == ErrUserAlreadyExist {
		err = cApi.Client.UpdateUser(userInfos.ID, func(user *backend.UserData) (_ error) {
			found := false
			for _, uToken := range user.OauthTokens {
				if uToken.Provider == provider {
					uToken.Token = token.AccessToken
					found = true
					break
				}
			}
			if !found {
				user.OauthTokens = append(user.OauthTokens, backend.OauthToken{
					Provider: provider,
					Token:    token.AccessToken,
				})
			}
			user.ExtraClaims.DisplayName = userInfos.Name
			if len(userInfos.Email) != 0 && userInfos.Email != user.ExtraClaims.Email {
				user.ExtraClaims.Email = userInfos.Email
				user.ExtraClaims.EmailVerified = true
			}
			return
		})
	}
	if err != nil {
		response.WriteError(http.StatusInternalServerError, fmt.Errorf("user infos cannot be upgraded in backend: %s", err.Error()))
	}
	response.AddHeader("Authorization", "Bearer "+token.AccessToken)
	response.WriteHeader(http.StatusOK)
}

func requireEnv(name, description string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatal("Env ", name, " is required: ", description)
	}
	return v
}
