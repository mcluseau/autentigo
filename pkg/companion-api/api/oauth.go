package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	restful "github.com/emicklei/go-restful"
	"github.com/mcluseau/autentigo/auth"
	"github.com/mcluseau/autentigo/pkg/companion-api/backend"
	"golang.org/x/oauth2"
)

var (
	setupOauthOnce    = sync.Once{}
	oauthConfig       *oauth2.Config
	oauthState        string
	oauthUserInfosURL string
)

func setup() {
	setupOauthOnce.Do(func() {
		oauthConfig = &oauth2.Config{
			RedirectURL:  requireEnv("OAUTH_MYAPP_ENDPOINT", "This app endpoint") + "/oauth/callback",
			ClientID:     requireEnv("OAUTH_CLIENTID", "oauth id given by oauth client"),
			ClientSecret: requireEnv("OAUTH_CLIENTSECRET", "oauth secret given by oauth client"),
			Scopes:       strings.Split(os.Getenv("OAUTH_SCOPES"), ","),
			Endpoint: oauth2.Endpoint{
				AuthURL:  requireEnv("OAUTH_AUTHURL", "url of the client auth provider"),
				TokenURL: requireEnv("OAUTH_TOKENURL", "url of the client token provider"),
			},
		}
		oauthUserInfosURL = requireEnv("OAUTH_USERINFOSURL", "url of the client user informations provider")

		oauthState = os.Getenv("OAUTH_STATE")
		if len(oauthState) == 0 {
			// If I were a state, I would be a fish
			oauthState = "><((('>"
		}
	})
}

// Register provide a restful.WebService from this API
func (cApi *CompanionAPI) oauthWS() (ws *restful.WebService) {
	ws = &restful.WebService{}
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Path("/oauth")

	ws.
		Route(ws.GET("/register").
			To(cApi.register).
			Doc("Create or update user with informations given by oauth"))

	ws.
		Route(ws.GET("/callback").
			To(cApi.registerCallback).
			Doc("Don't use it directly. Oauth register callback"))

	return ws
}

func (cApi *CompanionAPI) register(request *restful.Request, response *restful.Response) {
	setup()
	url := oauthConfig.AuthCodeURL(oauthState)
	http.Redirect(response.ResponseWriter, request.Request, url, http.StatusTemporaryRedirect)
}

type OAuthUserInfos struct {
	ID    string
	Name  string
	Email string
}

func (cApi *CompanionAPI) registerCallback(request *restful.Request, response *restful.Response) {
	state := request.Request.FormValue("state")
	code := request.Request.FormValue("code")

	if state != oauthState {
		response.WriteErrorString(http.StatusUnauthorized, "invalid oauth state")
	}

	token, err := oauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		response.WriteError(http.StatusUnauthorized, fmt.Errorf("code exchange failed: %s", err.Error()))
		return
	}
	if !token.Valid() {
		response.WriteErrorString(http.StatusUnauthorized, "token not valid")
		return
	}

	r, errr := http.Get(oauthUserInfosURL + "?fields=name,email&access_token=" + token.AccessToken)
	if errr != nil {
		response.WriteError(http.StatusUnauthorized, fmt.Errorf("failed getting user info: %s", err.Error()))
		return
	}

	defer r.Body.Close()
	contents, err := ioutil.ReadAll(r.Body)
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

	err = cApi.Client.CreateUser(userInfos.ID, &backend.UserData{
		ExtraClaims: auth.ExtraClaims{
			DisplayName:   userInfos.Name,
			Email:         userInfos.Email,
			EmailVerified: true,
		},
	})

	if err == ErrUserAlreadyExist {
		err = cApi.Client.UpdateUser(userInfos.ID, func(user *backend.UserData) (_ error) {
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

	response.WriteHeader(http.StatusTemporaryRedirect)
}

func requireEnv(name, description string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatal("Env ", name, " is required: ", description)
	}
	return v
}
