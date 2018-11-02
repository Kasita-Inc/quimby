package controllers

import (
	"net/http"

	qerror "github.com/Kasita-Inc/quimby/error"
	qhttp "github.com/Kasita-Inc/quimby/http"
)

// MethodNotAllowedController serves as a base for controllers that do not
// implement all the complete controller interface.
type MethodNotAllowedController struct{}

// GetRoutes returns an emtpy string array.
func (controller MethodNotAllowedController) GetRoutes() []string {
	return []string{}
}

// Get returns a method not allowed status
func (controller MethodNotAllowedController) Get(context *qhttp.Context) {
	context.SetError(qerror.NewRestError(qerror.MethodNotAllowed, "", nil), http.StatusMethodNotAllowed)
}

// Post returns a method not allowed status
func (controller MethodNotAllowedController) Post(context *qhttp.Context) {
	context.SetError(qerror.NewRestError(qerror.MethodNotAllowed, "", nil), http.StatusMethodNotAllowed)
}

// Put returns a method not allowed status
func (controller MethodNotAllowedController) Put(context *qhttp.Context) {
	context.SetError(qerror.NewRestError(qerror.MethodNotAllowed, "", nil), http.StatusMethodNotAllowed)
}

// Patch returns a method not allowed status
func (controller MethodNotAllowedController) Patch(context *qhttp.Context) {
	context.SetError(qerror.NewRestError(qerror.MethodNotAllowed, "", nil), http.StatusMethodNotAllowed)
}

// Delete returns a method not allowed status
func (controller MethodNotAllowedController) Delete(context *qhttp.Context) {
	context.SetError(qerror.NewRestError(qerror.MethodNotAllowed, "", nil), http.StatusMethodNotAllowed)
}

// Options returns a method not allowed status
func (controller MethodNotAllowedController) Options(context *qhttp.Context) {
	context.SetError(qerror.NewRestError(qerror.MethodNotAllowed, "", nil), http.StatusMethodNotAllowed)
}
