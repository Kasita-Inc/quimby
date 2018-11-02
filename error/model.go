package error

import "fmt"

// RestError represents the standard error returned by the API Gateway
type RestError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`
}

// NewRestError instantiates a RestError
func NewRestError(code string, message string, details []interface{}) *RestError {
	return &RestError{
		Code:    code,
		Message: message,
		Details: details,
	}

}

func (restError *RestError) Error() string {
	return fmt.Sprintf("%s (%s): %#v", restError.Message, restError.Code, restError.Details)
}

// AddDetail adds a detail such as a FieldError to an Error response
func (restError *RestError) AddDetail(errorDetail interface{}) {
	restError.Details = append(restError.Details, errorDetail)
}

// FieldError represents a validation error related to a specific input field
type FieldError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// NewFieldError instantiates a FieldError
func NewFieldError(code, message, field string) *FieldError {
	return &FieldError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}
