package error

const (
	// CannotBeBlank indicates a field that was submitted blank, but is required
	CannotBeBlank = "cannot-be-blank"
	// ValidationError indicates that a validation rule such as min / max value was violated
	ValidationError = "validation-error"
)
