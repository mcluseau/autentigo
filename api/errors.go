package api

import (
	"log"
	"net/http"
	"runtime"

	"github.com/emicklei/go-restful"
)

// Write error in good http format with error stack in it
func WriteError(err error, response *restful.Response) {
	response.AddHeader("Content-Type", "text/plain")

	if rfErr, ok := err.(restful.ServiceError); ok {
		response.WriteError(rfErr.Code, rfErr)
		return
	}

	ba := make([]byte, 10240)
	n := runtime.Stack(ba, false)
	log.Print("error during request: ", err, "\n", string(ba[:n]))

	status := http.StatusInternalServerError

	response.WriteErrorString(status, http.StatusText(status)+"\n")
}
