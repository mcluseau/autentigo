package api

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/mcluseau/autentigo/pkg/companion-api/backend"
	"github.com/mcluseau/autentigo/pkg/rbac"
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
	Client     backend.Client
	AdminToken string
}

// Register provide a restful.WebService from this API
func (cApi *CompanionAPI) WebServices() []*restful.WebService {
	return []*restful.WebService{
		cApi.meWS(),
		cApi.usersWS(),
	}
}

func requireRole(bypass, role string) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		if len(bypass) != 0 && req.HeaderParameter("Authorization") == "Bearer "+bypass {
			chain.ProcessFilter(req, resp)
			return
		}

		u := rbac.UserFromRequest(req.Request, rbac.DefaultValidationCertificate)
		if u == nil {
			sc := http.StatusUnauthorized
			resp.WriteErrorString(sc, http.StatusText(sc))
			return
		}

		if !rbac.Match(role, u) {
			sc := http.StatusForbidden
			resp.WriteErrorString(sc, http.StatusText(sc))
			return
		}

		req.SetAttribute("user", u)

		chain.ProcessFilter(req, resp)
	}
}
