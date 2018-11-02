package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Kasita-Inc/gadget/generator"
	"github.com/Kasita-Inc/gadget/stringutil"
	"github.com/Kasita-Inc/quimby/controllers"
	qerror "github.com/Kasita-Inc/quimby/error"
	qhttp "github.com/Kasita-Inc/quimby/http"
)

// WidgetStorage is a simple in-memory map for widgets
type WidgetStorage struct {
	widgets map[string]*Widget
	mutex   sync.RWMutex
}

// NewWidgetStorage returns an instace of WidgetStorage
func NewWidgetStorage() *WidgetStorage {
	return &WidgetStorage{
		widgets: make(map[string]*Widget),
	}
}

// Get returns a Widget from storage by ID if found
func (s *WidgetStorage) Get(key string) (*Widget, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	w, ok := s.widgets[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return w, nil
}

// List of all the Widgets in storage
func (s *WidgetStorage) List() []*Widget {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	widgets := []*Widget{}
	for _, widget := range s.widgets {
		widgets = append(widgets, widget)
	}
	return widgets
}

// Delete a Widget from storage
func (s *WidgetStorage) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.widgets, key)
}

// Create a new Widget in storage
func (s *WidgetStorage) Create(req *WidgetRequest) *Widget {
	widget := &Widget{
		ID:           generator.ID("wgt"),
		SerialNumber: req.SerialNumber,
		Description:  req.Description,
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.widgets[widget.ID] = widget
	return widget
}

// Update a Widget in storage
func (s *WidgetStorage) Update(widget *Widget) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.widgets[widget.ID]
	if !ok {
		return errors.New("not found")
	}

	s.widgets[widget.ID] = widget
	return nil
}

// WidgetRequest is the allowed input for a Widget (POST, PUT)
type WidgetRequest struct {
	SerialNumber string `json:"serial_number"`
	Description  string `json:"description"`
	issues       []string
}

// Valid indicates if the request is complete
func (req *WidgetRequest) Valid() bool {
	req.issues = []string{}
	if stringutil.IsWhiteSpace(req.SerialNumber) {
		req.issues = append(req.issues, "SerialNumber cannot be blank")
	}
	if stringutil.IsWhiteSpace(req.Description) {
		req.issues = append(req.issues, "Description cannot be blank")
	}
	return 0 == len(req.issues)
}

// Error returns the error string if the request has issues
func (req *WidgetRequest) Error() string {
	if 0 == len(req.issues) {
		return ""
	}
	return strings.Join(req.issues, "\n")
}

// WidgetPatch is the allowed input for a Widget (PATCH)
type WidgetPatch struct {
	Description string `json:"description"`
	issues      []string
}

// Valid indicates if the request is complete
func (req *WidgetPatch) Valid() bool {
	req.issues = []string{}
	if stringutil.IsWhiteSpace(req.Description) {
		req.issues = append(req.issues, "Description cannot be blank")
	}
	return 0 == len(req.issues)
}

// Error returns the error string if the request has issues
func (req *WidgetPatch) Error() string {
	if 0 == len(req.issues) {
		return ""
	}
	return strings.Join(req.issues, "\n")
}

// Widget is an example object representing a non-descript widget
type Widget struct {
	ID           string `json:"id"`
	SerialNumber string `json:"serial_number"`
	Description  string `json:"description"`
}

// WidgetCollection is a collection of Widgets
type WidgetCollection struct {
	Items []*Widget `json:"items"`
}

// WidgetsController handles operations on the Widget Collection (create and list)
type WidgetsController struct {
	controllers.MethodNotAllowedController
	storage *WidgetStorage
}

// NewWidgetsController returns an initialized WidgetsController
func NewWidgetsController(storage *WidgetStorage) *WidgetsController {
	return &WidgetsController{
		storage: storage,
	}
}

// GetRoutes establishes routes for the AccountPermissionsController
func (controller *WidgetsController) GetRoutes() []string {
	return []string{
		"api/widgets",
	}
}

// Authenticate always returns true
func (controller WidgetsController) Authenticate(context *qhttp.Context) bool {
	return widgetAuthentication(context)
}

func widgetAuthentication(context *qhttp.Context) bool {
	return "valid" == context.Request.Header.Get("Authorization")
}

// Get a list of Widgets
func (controller *WidgetsController) Get(context *qhttp.Context) {
	resp := &WidgetCollection{
		Items: controller.storage.List(),
	}
	context.SetResponse(resp, http.StatusOK)
}

// Post to create a widget
func (controller *WidgetsController) Post(context *qhttp.Context) {
	req := &WidgetRequest{}
	if err := context.ReadObject(req); nil != err {
		context.SetError(&qerror.RestError{Code: qerror.ValidationError, Message: err.Error()}, http.StatusNotAcceptable)
		return
	}
	widget := controller.storage.Create(req)
	context.SetResponse(widget, http.StatusCreated)
}

// WidgetController handles operations on a Widget (get, replace, update, delete)
type WidgetController struct {
	controllers.MethodNotAllowedController
	storage *WidgetStorage
}

// NewWidgetController returns an initialized WidgetController
func NewWidgetController(storage *WidgetStorage) *WidgetController {
	return &WidgetController{
		storage: storage,
	}
}

// Authenticate always returns true
func (controller WidgetController) Authenticate(context *qhttp.Context) bool {
	return widgetAuthentication(context)
}

// GetRoutes establishes routes for the AccountPermissionsController
func (controller *WidgetController) GetRoutes() []string {
	return []string{
		"api/widgets/{{id}}",
	}
}

// Get a Widget
func (controller *WidgetController) Get(context *qhttp.Context) {
	widget, err := controller.storage.Get(context.URIParameters["id"])
	if err != nil {
		context.SetError(
			qerror.NewRestError(qerror.NotFound, fmt.Sprintf("No widget for ID %s found", context.URIParameters["id"]), nil),
			http.StatusNotFound)
		return
	}
	context.SetResponse(widget, http.StatusOK)
}

// Delete to delete a widget
func (controller *WidgetController) Delete(context *qhttp.Context) {
	controller.storage.Delete(context.URIParameters["id"])
	context.SetResponse("", http.StatusNoContent)
}

// Put to replace a Widget
func (controller *WidgetController) Put(context *qhttp.Context) {
	req := &WidgetRequest{}
	if err := context.ReadObject(req); nil != err {
		context.SetError(&qerror.RestError{Code: qerror.ValidationError, Message: err.Error()}, http.StatusNotAcceptable)
		return
	}
	if !req.Valid() {
		context.SetError(&qerror.RestError{Code: qerror.ValidationError, Message: req.Error()}, http.StatusNotAcceptable)
		return
	}
	widget := &Widget{
		ID:           context.URIParameters["id"],
		Description:  req.Description,
		SerialNumber: req.SerialNumber,
	}
	if err := controller.storage.Update(widget); nil != err {
		context.SetError(
			qerror.NewRestError(qerror.NotFound, fmt.Sprintf("No widget for ID %s found", context.URIParameters["id"]), nil),
			http.StatusNotFound)
	}

	context.SetResponse(widget, http.StatusOK)
}

// Patch update part of a Widget
func (controller *WidgetController) Patch(context *qhttp.Context) {
	widget, err := controller.storage.Get(context.URIParameters["id"])
	if err != nil {
		context.SetError(
			qerror.NewRestError(qerror.NotFound, fmt.Sprintf("No widget for ID %s found", context.URIParameters["id"]), nil),
			http.StatusNotFound)
		return
	}

	req := &WidgetPatch{}
	if err := context.ReadObject(req); nil != err {
		context.SetError(&qerror.RestError{Code: qerror.ValidationError, Message: err.Error()}, http.StatusNotAcceptable)
		return
	}
	if !req.Valid() {
		context.SetError(&qerror.RestError{Code: qerror.ValidationError, Message: req.Error()}, http.StatusNotAcceptable)
		return
	}
	widget.Description = req.Description
	if err := controller.storage.Update(widget); nil != err {
		context.SetError(
			qerror.NewRestError(qerror.NotFound, fmt.Sprintf("No widget for ID %s found", context.URIParameters["id"]), nil),
			http.StatusNotFound)
	}

	context.SetResponse(widget, http.StatusOK)
}

// APIController displays instructions for authenticating with the Widget API
type APIController struct {
	controllers.MethodNotAllowedController
	controllers.NoAuthenticationController
}

// GetRoutes establishes routes for the AccountPermissionsController
func (controller *APIController) GetRoutes() []string {
	return []string{
		"api",
	}
}

// Get the instructions for the example API
func (controller *APIController) Get(context *qhttp.Context) {
	context.Response.Header().Add("Content-Type", "text/html")
	context.SetResponse(`<html>
	<head>
		<title>Example API Documenation</title>
	</head>
	<body>
		<h3>Endpoints</h3>
		<ul>
			<li>[GET / POST] /api/widgets</li>
			<li>[GET / PUT / PATCH / DELETE] /api/widgets/{{id}}</li>
		</ul>
		<h3>Authentication</h3>
		<p>Send a header of <b>Authorization</b> with a value of <b>valid</b></p>
	</body>
	</html>`, http.StatusOK)
}
