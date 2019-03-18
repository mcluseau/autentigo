package api

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/mcluseau/autentigo/pkg/companion-api/backend"
	"github.com/mcluseau/autentigo/pkg/rbac"
)

// Register provide a restful.WebService from this API
func (cApi *CompanionAPI) meWS() (ws *restful.WebService) {
	ws = &restful.WebService{}
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(requireRole("self-service"))
	ws.Doc("Requires the self-service role")

	ws.
		Route(ws.PUT("/me").
			To(cApi.getMe).
			Doc("Get informations on the authenticated user.").
			Writes(&MeResponse{}))

	ws.
		Route(ws.PUT("/me/password").
			To(cApi.updatePassword).
			Doc("Update the authenticated user's password.").
			Reads(UpdatePasswordReq{}))

	return ws
}

type MeResponse struct {
	Sub string
}

func (cApi *CompanionAPI) getMe(request *restful.Request, response *restful.Response) {
	u := request.Attribute("user").(*rbac.User)
	response.WriteEntity(MeResponse{u.Name})
}

type UpdatePasswordReq struct {
	NewPassword string
}

func (cApi *CompanionAPI) updatePassword(request *restful.Request, response *restful.Response) {
	u := request.Attribute("user").(*rbac.User)

	r := &UpdatePasswordReq{}
	if err := request.ReadEntity(r); err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	h := sha256.Sum256([]byte(r.NewPassword))
	passwordHash := hex.EncodeToString(h[:])

	err := cApi.Client.UpdateUser(u.Name, func(user *backend.UserData) error {
		user.PasswordHash = passwordHash
		return nil
	})

	if err != nil {
		log.Print("update error on user ", u.Name, ": ", err)
		sc := http.StatusInternalServerError
		response.WriteErrorString(sc, http.StatusText(sc))
		return
	}
}
