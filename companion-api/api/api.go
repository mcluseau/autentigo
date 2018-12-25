package api

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/mcluseau/autorizo/companion-api/backend"
)

var (
	// ErrMissingContent indicates an inexistent user.
	ErrMissingUser = restful.NewError(http.StatusConflict, "Missing user")
	// ErrMissingUserId indicates an user without an id.
	ErrMissingUserId = restful.NewError(http.StatusUnprocessableEntity, "No user id given")
	// ErrMissingUserPassword indicates an user without a password.
	ErrMissingUserPassword = restful.NewError(http.StatusUnprocessableEntity, "No user password given.")
	// ErrUserAlreadyExist indicates an existing user that should not be.
	ErrUserAlreadyExist = restful.NewError(http.StatusConflict, "User already exist")
	//ErrPatchFail indicates the json-patch update fails.
	ErrPatchFail = restful.NewError(http.StatusConflict, "Patch update fails")
)

// CompanionAPI registering with restful
type CompanionAPI struct {
	Client backend.Client
}

// Register provide a restful.WebService from this API
func (cApi *CompanionAPI) Register() *restful.WebService {
	ws := &restful.WebService{}
	cApi.registerUsers(ws)
	return ws
}
