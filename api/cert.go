package api

import restful "github.com/emicklei/go-restful"

func (api *API) registerCertificate(ws *restful.WebService) {
	ws.
		Route(ws.GET("/validation-certificate").
			To(api.validationCertificate).
			Doc("Returns the certificate to use to validate token from this server").
			Produces("application/x-x509-user-cert"))
}

func (api *API) validationCertificate(request *restful.Request, response *restful.Response) {
	response.Write(api.CRTData)
}
