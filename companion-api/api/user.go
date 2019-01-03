package api

import (
	"errors"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/mcluseau/autorizo/companion-api/backend"
)

// CreateUserReq is a request to create a new UserData
type CreateUserReq struct {
	ID   string           `json:"id"`
	User backend.UserData `json:"user"`
}

// Register provide a restful.WebService from this API
func (cApi *CompanionAPI) registerUsers(ws *restful.WebService) {
	ws.
		Route(ws.POST("/users").
			To(cApi.createUser).
			Doc("Create a new user.").
			Consumes("application/json").
			Reads(CreateUserReq{}))

	ws.
		Route(ws.PUT("/users/{user-id}").
			To(cApi.updateUser).
			Doc("Update an existing user.").
			Consumes("application/json").
			Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
			Reads(backend.UserData{}))

	ws.
		Route(ws.PATCH("/users/{user-id}").
			To(cApi.patchUser).
			Doc("Patch an existing user (json-patch format).").
			Consumes("application/json").
			Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

	ws.
		Route(ws.DELETE("/users/{user-id}").
			To(cApi.deleteUser).
			Doc("Delete an existing user.").
			Consumes("application/json").
			Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

}

func (cApi *CompanionAPI) createUser(request *restful.Request, response *restful.Response) {
	defer func() {
		if err := recover(); err != nil {
			// unhandled error
			writeError(err.(error), response)
		}
	}()

	userReq := &CreateUserReq{}
	if err := request.ReadEntity(userReq); err != nil {
		panic(err)
	}

	if len(userReq.ID) == 0 {
		panic(ErrMissingUserId)
	}

	if len(userReq.User.PasswordHash) == 0 {
		panic(ErrMissingUserPassword)
	}

	if err := cApi.Client.CreateUser(userReq.ID, &userReq.User); err != nil {
		panic(err)
	}

	response.WriteHeader(http.StatusCreated)
}

func (cApi *CompanionAPI) updateUser(request *restful.Request, response *restful.Response) {
	defer func() {
		if err := recover(); err != nil {
			// unhandled error
			writeError(err.(error), response)
		}
	}()

	id := request.PathParameter("user-id")

	userData := &backend.UserData{}
	if err := request.ReadEntity(userData); err != nil {
		panic(err)
	}

	err := cApi.Client.UpdateUser(id, func(user *backend.UserData) error {
		*user = *userData
		return nil
	})

	if err != nil {
		panic(err)
	}

	response.WriteHeader(http.StatusOK)
}

func (cApi *CompanionAPI) patchUser(request *restful.Request, response *restful.Response) {
	// TODO
	response.WriteError(http.StatusNotImplemented, errors.New("not implemented"))
}

func (cApi *CompanionAPI) deleteUser(request *restful.Request, response *restful.Response) {
	defer func() {
		if err := recover(); err != nil {
			// unhandled error
			writeError(err.(error), response)
		}
	}()

	id := request.PathParameter("user-id")

	if err := cApi.Client.DeleteUser(id); err != nil {
		panic(err)
	}

	response.WriteHeader(http.StatusOK)
}
