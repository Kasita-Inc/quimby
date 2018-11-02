package error

const (
	// MethodNotAllowed indicates that the attempted VERB is not implemented for that endpoint
	MethodNotAllowed = "method-not-allowed"
	// MalformedURL indicates that the URL was not parsable as input
	MalformedURL = "malformed-url"
	// InvalidRoute indicates that the route does not exist
	InvalidRoute = "invalid-route"
	// AuthenticationFailed indicates that authentication did not complete successfully
	AuthenticationFailed = "authentication-failed"
	// NotAuthorized indicates that the currently authenticated user is not permitted to perform an action
	NotAuthorized = "not-authorized"
	// SystemError indicates that a systemic issue has occurred with the request
	SystemError = "system-error"
	// NotFound indicates that the requested resource was not found
	NotFound = "not-found"
)
