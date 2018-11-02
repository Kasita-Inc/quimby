package http

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/log"
	"github.com/Kasita-Inc/gadget/stringutil"
	qerror "github.com/Kasita-Inc/quimby/error"
)

// NoContentError is returned when Read is called and the Request has a 0
// content length.
type NoContentError struct {
	RequestPath   string
	RequestMethod string
	trace         []string
}

// NewNoContentError instantiates a NoContentError with a stack trace
func NewNoContentError(path, method string) errors.TracerError {
	return &NoContentError{
		RequestPath:   path,
		RequestMethod: method,
		trace:         errors.GetStackTrace(),
	}
}

func (err *NoContentError) Error() string {
	return fmt.Sprintf("Request (%s %s) cannot be 'Read' as it has no content.",
		err.RequestMethod, err.RequestPath)
}

// Trace returns the stack trace for the error
func (err *NoContentError) Trace() []string {
	return err.trace
}

// Context serves as a structure that tracks the state of a given http Request
// Response chain.
type Context struct {
	URIParameters map[string]string
	URLParameters url.Values
	URI           string
	URL           *url.URL
	Method        string

	Request  *http.Request
	Response http.ResponseWriter
	Route    *RouteNode

	responseStatus int
	Model          interface{}
	Error          *qerror.RestError

	Extended map[string]interface{}

	Body     string
	bodyRead bool
}

// Status returns the HTTP status of the response
func (context *Context) Status() int {
	return context.responseStatus
}

// SetError sets the HTTP status of the response and the error to be returned
func (context *Context) SetError(err *qerror.RestError, status int) {
	context.responseStatus = status
	context.Error = err
}

// HasError checks if there is an Error set on the Context
func (context *Context) HasError() bool {
	return nil != context.Error
}

// AddError adds an error detail to the context
// Primary use case is to add a series of validation type errors
func (context *Context) AddError(err qerror.FieldError) {
	if !context.HasError() {
		context.SetError(qerror.NewRestError(qerror.ValidationError, "", nil), http.StatusBadRequest)
	}
	context.Error.AddDetail(err)
}

// SetResponse sets the HTTP status and model to be rendered in the response write
// Returns false if there is an Error on the context otherwise true
func (context *Context) SetResponse(model interface{}, status int) bool {
	if context.HasError() {
		return false
	}
	context.responseStatus = status
	context.Model = model
	return true
}

// CreateContext initializes a Context from the passed Response and Request
// pair, and router. The router is used for detemplating and populating the
// URIParameters
func CreateContext(writer http.ResponseWriter, request *http.Request,
	router Router) *Context {
	var err error
	context := &Context{Request: request, Extended: make(map[string]interface{})}
	context.Response = writer
	context.URL = request.URL
	context.URI = request.RequestURI
	context.Method = request.Method
	context.URLParameters, err = url.ParseQuery(request.URL.RawQuery)

	if err != nil {
		// take a hard stance on malformed URL's
		context.SetError(qerror.NewRestError(qerror.MalformedURL, fmt.Sprintf("Malformed URL Parameters '%s'.", request.URL), nil),
			http.StatusBadRequest)
		return context
	}

	cleanPath := strings.Split(context.URI, "?")[0]
	cleanPath = strings.Trim(cleanPath, " /")
	context.Route, err = router.FindRouteForPath(cleanPath)
	if err != nil || context.Route == nil {
		context.SetError(qerror.NewRestError(qerror.InvalidRoute, "", nil), http.StatusBadRequest)
		return context
	}

	context.URIParameters = make(map[string]string)
	if !stringutil.IsWhiteSpace(cleanPath) {
		context.URIParameters, err = stringutil.Detemplate(context.Route.TemplateRoute, cleanPath)
		if err != nil {
			context.SetError(qerror.NewRestError(qerror.InvalidRoute, "", nil), http.StatusInternalServerError)
			return context
		}
	}

	if http.MethodOptions != context.Request.Method && !context.Route.Controller.Authenticate(context) {
		context.SetError(
			qerror.NewRestError(qerror.AuthenticationFailed, InvalidCredentialsErrorMessage, nil), http.StatusUnauthorized)
	}

	return context
}

// InvalidCredentialsErrorMessage is returned when Credentials are invalid
const InvalidCredentialsErrorMessage = "Invalid Credentials"

// Read reads the entire body of the request and returns it as a slice of
// bytes
func (context *Context) Read() ([]byte, error) {
	if context.bodyRead {
		return []byte(context.Body), nil
	}
	if context.Request.ContentLength <= 0 {
		return nil, NewNoContentError("", "")
	}
	body := make([]byte, context.Request.ContentLength)
	n, err := io.ReadFull(context.Request.Body, body)

	if err == io.ErrUnexpectedEOF {
		log.Errorf("warning:%s:%s: Request.ContentLength (%d) mismatch with actual body length (%d)", context.URI,
			context.Request.RemoteAddr, n, context.Request.ContentLength)
	}
	// Ignore EOF error
	if io.EOF == err {
		err = nil
	}
	context.Body = string(body)
	context.bodyRead = true
	return body, err
}

// readJSON takes the JSON content type body of the Request and unmarshals a JSON object the
// same type as the passed implementation of interface{}
func (context *Context) readJSON(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

// readForm takes the form-urlencoded content type body of the Request and populates an object
// of the same type as the passed interface{}
func (context *Context) readForm(body []byte, target interface{}) error {
	values, err := url.ParseQuery(string(body))
	if nil != err {
		return err
	}
	return context.valuesToObject(values, target)
}

func (context *Context) inspectModel(target interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(target))
	fieldMap := make(map[string]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fieldMap[v.Type().Field(i).Name] = v.Field(i).Type()
	}
	return fieldMap
}

func (context *Context) valuesToObject(values url.Values, target interface{}) error {
	fieldMap := context.inspectModel(target)
	valueMap := make(map[string]interface{})
	for fieldName, fieldType := range fieldMap {
		switch fieldType.(reflect.Type).Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Struct, reflect.UnsafePointer:
			break
		case reflect.Slice, reflect.Array:
			var arrayValues []string
			arrayFieldName := stringutil.Underscore(fieldName)
			queryVal := values.Get(arrayFieldName)
			// ignore empty query string arguments
			if !stringutil.IsEmpty(queryVal) {
				arrayValues = append(arrayValues, values[arrayFieldName]...)
			}
			arrayFieldName = stringutil.Underscore(fieldName) + "[]"
			queryVal = values.Get(arrayFieldName)
			if !stringutil.IsEmpty(queryVal) {
				arrayValues = append(arrayValues, values[arrayFieldName]...)
			}
			if len(arrayValues) > 0 {
				valueMap[fieldName] = arrayValues
			}
		default:
			queryVal := values.Get(stringutil.Underscore(fieldName))
			if !stringutil.IsEmpty(queryVal) {
				valueMap[fieldName] = queryVal
			}
		}
	}
	data, err := json.Marshal(valueMap)
	if nil != err {
		return err
	}
	return json.Unmarshal(data, target)
}

// ReadObject reads the body of the Request and unmarshals an object the
// same type as the passed implementation of interface{}
func (context *Context) ReadObject(target interface{}) error {
	body, err := context.Read()

	if err != nil {
		return err
	}

	contentType, _, err := mime.ParseMediaType(context.Request.Header.Get(contentTypeHeader))
	if nil != err {
		context.SetError(qerror.NewRestError(qerror.ValidationError, err.Error(), nil), http.StatusNotAcceptable)
		return err
	}
	switch contentType {
	case contentTypeForm:
		err = context.readForm(body, target)
	case contentTypeJSON:
		err = context.readJSON(body, target)
	default:
		err = errors.New("Unsupported contentType (%s) provided", contentType)
	}

	if nil != err {
		context.SetError(qerror.NewRestError(qerror.ValidationError, err.Error(), nil), http.StatusNotAcceptable)
	}

	return err
}

// Write writes a string to the response body.
func (context *Context) Write(s string) {
	context.Response.Write([]byte(s))
}

// ReadQueryParams converts URL Parameters into an Object
func (context *Context) ReadQueryParams(target interface{}) error {
	return context.valuesToObject(context.URLParameters, target)
}
