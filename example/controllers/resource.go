package controllers

import (
	"fmt"

	"github.com/Kasita-Inc/quimby/controllers"
	"github.com/Kasita-Inc/quimby/http"
)

// ResourceController is a sample controller implementation
type ResourceController struct {
	controllers.MethodNotAllowedController
	controllers.NoAuthenticationController
}

// GetRoutes demonstrates multiple routes and multiple URI parameters
func (controller *ResourceController) GetRoutes() []string {
	return []string{
		"resource/{{id}}",
		"resource/{{id}}/{{subresource}}"}
}

// Get returns the ID and subresource (if provided)
func (controller *ResourceController) Get(context *http.Context) {
	id, _ := context.URIParameters["id"]
	subresource, ok := context.URIParameters["subresource"]
	context.Write(fmt.Sprintf("ID: %s\n", id))
	if ok {
		context.Write(fmt.Sprintf("Subresource: %s", subresource))
	}
}
