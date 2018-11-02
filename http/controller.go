package http

// Controller is the main interface for the request handlers in the Router
type Controller interface {
	GetRoutes() []string
	Get(context *Context)
	Post(context *Context)
	Put(context *Context)
	Patch(context *Context)
	Delete(context *Context)
	Options(context *Context)
	Authenticate(context *Context) bool
}

// HealthCheckRoute is the default URI for quimby health checks
const HealthCheckRoute = "health"
