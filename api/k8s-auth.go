package api

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (api *API) registerK8sAuthenticator(ws *restful.WebService) {
	ws.
		Route(ws.POST("/review-token").
			To(api.k8sTokenReview).
			Doc("Kubernetes token review").
			Consumes("application/json").
			Produces("application/json").
			Reads(authv1.TokenReview{}).
			Writes(authv1.TokenReview{}))
}

func (api *API) k8sTokenReview(request *restful.Request, response *restful.Response) {
	req := &authv1.TokenReview{}
	if err := request.ReadEntity(req); err != nil {
		response.WriteError(http.StatusBadRequest, err)
		return
	}

	claims, err := api.checkToken(req.Spec.Token)

	groupVersionKind := authv1.SchemeGroupVersion.WithKind("TokenReview")

	tr := &authv1.TokenReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: groupVersionKind.Version,
			Kind:       groupVersionKind.Kind,
		},
	}

	if err != nil {
		tr.Status = authv1.TokenReviewStatus{
			Authenticated: false,
			Error:         err.Error(),
		}

		response.WriteHeaderAndEntity(http.StatusUnauthorized, tr)
		return
	}

	extra := map[string]authv1.ExtraValue{}

	if claims.Email != "" {
		extra["email"] = authv1.ExtraValue{claims.Email}
	}

	if claims.EmailVerified {
		extra["email_verified"] = authv1.ExtraValue{"true"}
	}

	tr.Status = authv1.TokenReviewStatus{
		Authenticated: true,
		User: authv1.UserInfo{
			Username: claims.Subject,
			Groups:   claims.Groups,
			Extra:    extra,
		},
	}

	response.WriteEntity(tr)
}
