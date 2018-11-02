package http

import (
	"net/http"
	"testing"
)

/******************************************************
 *          Supporting code for tests                 *
 ******************************************************/
type TestController struct {
	ID           string
	Routes       []string
	MethodCalled string
}

func NewTestController(ID string) TestController {
	return TestController{ID: ID, Routes: []string{}}
}

func (controller *TestController) Authenticate(context *Context) bool {
	return true
}

func (controller *TestController) GetRoutes() []string {
	return controller.Routes
}

func (controller *TestController) Get(context *Context) {
	controller.MethodCalled = http.MethodGet
	context.SetResponse(http.MethodGet, http.StatusOK)
}

func (controller *TestController) Post(context *Context) {
	controller.MethodCalled = http.MethodPost
	context.SetResponse(http.MethodPost, http.StatusCreated)
}

func (controller *TestController) Put(context *Context) {
	controller.MethodCalled = http.MethodPut
	context.SetResponse(http.MethodPut, http.StatusAccepted)
}

func (controller *TestController) Patch(context *Context) {
	controller.MethodCalled = http.MethodPatch
	context.SetResponse(http.MethodPatch, http.StatusAccepted)
}

func (controller *TestController) Delete(context *Context) {
	controller.MethodCalled = http.MethodDelete
	context.SetResponse(http.MethodDelete, http.StatusNoContent)
}

func (controller *TestController) Options(context *Context) {
	controller.MethodCalled = http.MethodOptions
	context.SetResponse(http.MethodOptions, http.StatusOK)
}

type NoAuthTestController struct {
	ID           string
	Routes       []string
	MethodCalled string
}

func NewNoAuthTestController(ID string) NoAuthTestController {
	return NoAuthTestController{ID: ID, Routes: []string{}}
}

func (controller *NoAuthTestController) Authenticate(context *Context) bool {
	return false
}

func (controller *NoAuthTestController) GetRoutes() []string {
	return controller.Routes
}

func (controller *NoAuthTestController) Get(context *Context) {
	controller.MethodCalled = http.MethodGet
	context.SetResponse(http.MethodGet, http.StatusOK)
}

func (controller *NoAuthTestController) Post(context *Context) {
	controller.MethodCalled = http.MethodPost
	context.SetResponse(http.MethodPost, http.StatusCreated)
}

func (controller *NoAuthTestController) Put(context *Context) {
	controller.MethodCalled = http.MethodPut
	context.SetResponse(http.MethodPut, http.StatusAccepted)
}

func (controller *NoAuthTestController) Patch(context *Context) {
	controller.MethodCalled = http.MethodPatch
	context.SetResponse(http.MethodPatch, http.StatusAccepted)
}

func (controller *NoAuthTestController) Delete(context *Context) {
	controller.MethodCalled = http.MethodDelete
	context.SetResponse(http.MethodDelete, http.StatusNoContent)
}

func (controller *NoAuthTestController) Options(context *Context) {
	controller.MethodCalled = http.MethodOptions
	context.SetResponse(http.MethodOptions, http.StatusOK)
}

/******************************************************
 *                      Tests                         *
 ******************************************************/

func TestBadRoute(t *testing.T) {
	r := CreateRouter(nil)
	controller := &TestController{ID: "foo", Routes: []string{}}
	err := r.AddRoute("/route", controller)
	if err == nil {
		t.Error("Leading slash should fail.")
	}
	err = r.AddRoute("route/", controller)
	if err == nil {
		t.Error("Trailing slash should fail.")
	}
	err = r.AddRoute("route//route", controller)
	if err == nil {
		t.Error("double slash should fail.")
	}
}

func TestAddRoute(t *testing.T) {
	r := CreateRouter(nil)
	expectedID := "controller1"
	controllerRoute := "controllerRoute1"
	expectedController := &TestController{ID: expectedID,
		Routes: []string{controllerRoute}}
	controller2 := &TestController{ID: "foo", Routes: []string{"testing"}}
	expected := "route"
	r.AddRoute(expected, expectedController)
	r.AddRoute("route2", controller2)

	// Controllers route should not have been added
	_, routeAdded := r.RouteTree.SubRoutes[controllerRoute]
	if routeAdded {
		t.Error("Controller's Route should not have been added.")
	}
	var actualNode *RouteNode
	// We expect the passed route to have been added
	actualNode, routeAdded = r.RouteTree.SubRoutes[expected]
	if !routeAdded {
		t.Error("Route should have been added to controller")
	}

	// We expect the route to be the controller with the correct ID and Type.
	coercedController, correctType := actualNode.Controller.(*TestController)
	if !correctType {
		t.Error("Incorrect Controller type returned for route.")
	}
	if coercedController.ID != expectedID {
		t.Errorf("Incorrect Controller returned for route. Got %s",
			coercedController.ID)
	}
}

func TestAddControllerAtExistingRoute(t *testing.T) {
	controller := &TestController{ID: "ID", Routes: []string{}}
	router := CreateRouter(nil)
	route := "route"
	err := router.AddRoute(route, controller)
	if err != nil {
		t.Error("AddRoute should have been successful on empty router.")
	}
	err = router.AddRoute(route, controller)
	if err == nil {
		t.Error("Subsequent call to add route with same route should fail.")
	}
}

func TestAddControllerAtExistingRouteComplex(t *testing.T) {
	controller := &TestController{ID: "ID", Routes: []string{}}
	router := CreateRouter(nil)
	route := "route/{{id}}/something"
	err := router.AddRoute(route, controller)
	if err != nil {
		t.Error("AddRoute should have been successful on empty router.")
	}
	err = router.AddRoute(route, controller)
	if err == nil {
		t.Error("Subsequent call to add route with same route should fail.")
	}
}

func TestAddEmptyRoute(t *testing.T) {
	r := CreateRouter(nil)
	controller := &TestController{ID: "foo", Routes: []string{}}
	err := r.AddRoute("", controller)

	if err == nil {
		t.Error("Empty route should fail.")
	}

	err = r.AddRoute(" / ", controller)
	if err == nil {
		t.Error("Empty route should fail.")
	}
}

func testRoute(router Router, route string, expectedID string, t *testing.T) {
	actual, err := router.FindRouteForPath(route)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Route: %s, Node Value: %s, Subroutes: %d", route, actual.Value,
		len(actual.SubRoutes))
	if actual.Controller == nil {
		t.Error("Non-terminal node returned.")
		return
	}
	// We expect the route to be the controller with the correct ID and Type.
	coercedController, correctType := actual.Controller.(*TestController)
	if !correctType {
		t.Error("Incorrect Controller type returned for route.")
		return
	}
	if coercedController.ID != expectedID {
		t.Errorf("Incorrect Controller returned for route. Got %s",
			coercedController.ID)
		return
	}
}

func TestAddController(t *testing.T) {
	r := CreateRouter(nil)
	expected := "TestAddController"
	route1 := "bar/baz"
	route2 := "foo/bar"
	controller := &TestController{ID: expected, Routes: []string{route1, route2}}
	r.AddController(controller)
	testRoute(r, route1, expected, t)
	testRoute(r, route2, expected, t)
}

func TestFindControllerSingleRouteExact(t *testing.T) {
	r := CreateRouter(nil)
	expected := "controllerID"
	controller := &TestController{ID: expected, Routes: []string{}}
	route := "route/something"
	err := r.AddRoute(route, controller)
	if err != nil {
		t.Error(err)
	}

	testRoute(r, route, expected, t)
}

func TestFindControllerSingleRouteWild(t *testing.T) {
	r := CreateRouter(nil)
	expected := "controllerID"
	controller := &TestController{ID: expected, Routes: []string{}}
	route := "foo/{{baz}}/bar"
	err := r.AddRoute(route, controller)
	if err != nil {
		t.Error(err)
	}
	actual, err2 := r.FindRouteForPath("foo/whatever/bar")
	if err2 != nil {
		t.Error(err)
		t.FailNow()
	}

	if actual.Controller == nil {
		t.Error("Non-terminal node returned.")
	}
	// We expect the route to be the controller with the correct ID and Type.
	actualController := actual.Controller
	coercedController, correctType := actualController.(*TestController)
	if !correctType {
		t.Error("Incorrect Controller type returned for route.")
	}
	if coercedController.ID != expected {
		t.Errorf("Incorrect Controller returned for route. Got %s",
			coercedController.ID)
	}
}

func TestFindControllerSinglePathExactOverWild(t *testing.T) {
	r := CreateRouter(nil)
	expected := "TestFindControllerSinglePathExactOverWild"
	controller := &TestController{ID: expected, Routes: []string{}}
	controller2 := &TestController{ID: "bad", Routes: []string{}}
	exactRoute := "foo/bar/baz"
	wildRoute := "foo/{{id}}/baz"

	err := r.AddRoute(exactRoute, controller)
	if err != nil {
		t.Error(err)
	}

	err = r.AddRoute(wildRoute, controller2)
	if err != nil {
		t.Error(err)
	}

	actual, err2 := r.FindRouteForPath(exactRoute)
	if err2 != nil {
		t.Error(err)
	}

	if actual.Controller == nil {
		t.Error("Non-terminal node returned.")
	}
	// We expect the route to be the controller with the correct ID and Type.
	coercedController, correctType := actual.Controller.(*TestController)
	if !correctType {
		t.Error("Incorrect Controller type returned for route.")
	}
	if coercedController.ID != expected {
		t.Errorf("Incorrect Controller returned for route. Got %s",
			coercedController.ID)
	}
}

func TestFindControllerNoRoute(t *testing.T) {
	r := CreateRouter(nil)
	controller := &TestController{ID: "TestFindControllerNoRoute",
		Routes: []string{}}
	r.AddRoute("route/*/something", controller)
	r.AddRoute("route/awef/something", controller)
	_, err := r.FindRouteForPath("does/not/exist")
	if err == nil {
		t.Errorf("Expected no route for path error.")
	}
}

func TestFindControllerNoRoutePartialMatch(t *testing.T) {
	r := CreateRouter(nil)
	controller := &TestController{ID: "TestFindControllerNoRoute",
		Routes: []string{}}
	r.AddRoute("route/*/something", controller)
	r.AddRoute("route/awef/something", controller)
	node, err := r.FindRouteForPath("route/foo/")
	if err == nil {
		t.Errorf("Expected no route for path error. Got node with Value: %s",
			node.Value)
	}
}
