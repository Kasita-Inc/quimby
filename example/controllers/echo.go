package controllers

import (
	"fmt"
	"net/http"

	"github.com/Kasita-Inc/gadget/stringutil"
	"github.com/Kasita-Inc/quimby/controllers"
	qerror "github.com/Kasita-Inc/quimby/error"
	qhttp "github.com/Kasita-Inc/quimby/http"
)

// EchoController is a debugging tool for echo'ing back the request sent in
// as the body of the response.
type EchoController struct {
	controllers.MethodNotAllowedController
	controllers.NoAuthenticationController
}

// GetRoutes returns the single route 'echo'
func (controller *EchoController) GetRoutes() []string {
	return []string{
		"echo",
		"echo/{{toEcho}}",
	}
}

// Get writes the information from the request to the body of the response.
func (controller *EchoController) Get(context *qhttp.Context) {
	r := context.Request
	context.Write(fmt.Sprintf("Host: %s\n", r.Host))
	context.Write(fmt.Sprintf("RequestURI: %s\n", r.RequestURI))
	context.Write(fmt.Sprintf("Method: %s\n", r.Method))
	context.Write(fmt.Sprintf("RemoteAddr: %s\n", r.RemoteAddr))
	context.Write(fmt.Sprintf("Content Length: %d\n", r.ContentLength))
	context.Write("Headers:\n")
	for k, v := range context.Request.Header {
		context.Write(fmt.Sprintf("\t%s: %s\n", k, v))
	}

	if !stringutil.IsWhiteSpace(context.URIParameters["toEcho"]) {
		context.Write(fmt.Sprintf("URI:\n%s\n", context.URIParameters["toEcho"]))
	}

	if context.Request.ContentLength > 0 {
		context.Write("Body:\n")
		body, err := context.Read()
		if err != nil {
			context.SetError(qerror.NewRestError(qerror.SystemError, "", nil), http.StatusInternalServerError)
			return
		}
		context.Response.Write(body)
	}

}

// Post writes the information from the request to the body of the response.
func (controller *EchoController) Post(context *qhttp.Context) {
	controller.Get(context)
}

// Put writes the information from the request to the body of the response.
func (controller *EchoController) Put(context *qhttp.Context) {
	controller.Get(context)
}

// Delete writes the information from the request to the body of the response.
func (controller *EchoController) Delete(context *qhttp.Context) {
	controller.Get(context)
}
