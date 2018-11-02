package controllers

import (
	nhttp "net/http"
	"time"

	"github.com/Kasita-Inc/quimby/http"
)

// HealthCheckController responds to Get requests with a JSON object representing
// the status of the server.
type HealthCheckController struct {
	MethodNotAllowedController
	NoAuthenticationController
}

// HealthCheckResource is returned by the HealthCheckController for conveying
// the status of the server.
type HealthCheckResource struct {
	Timestamp string
	Status    string
}

// GetRoutes only responds on the 'health' resource endpoint.
func (controller *HealthCheckController) GetRoutes() []string {
	return []string{http.HealthCheckRoute}
}

// Get returns a new instance of the HealthCheckResource
func (controller *HealthCheckController) Get(context *http.Context) {
	model := HealthCheckResource{Timestamp: time.Now().UTC().Format(time.RFC822),
		Status: "OK"}
	context.SetResponse(model, nhttp.StatusOK)
}
