package http

import (
	"fmt"
	"strings"

	"github.com/Kasita-Inc/gadget/stringutil"
)

// WildCard character for paths.
var (
	WildCard = "*"
	Slash    = "/"
)

// Router is the main entry point for the Quimby ReST API server.
type Router struct {
	// Routes All routes currently mapped by the router.
	RouteTree        *RouteNode
	RegisteredRoutes []string
}

// RouteNode serves as a node in the parse tree for parsing incoming routes.
type RouteNode struct {
	// The word value assigned to this node.
	Value string
	// Any subroutes nested below this node.
	SubRoutes map[string]*RouteNode
	// The route template that was mapped to this node (if termnial).
	TemplateRoute string
	// The Controller for this node if terminal.
	Controller Controller
}

func createNode(value string) *RouteNode {
	return &RouteNode{Value: value, SubRoutes: make(map[string]*RouteNode)}
}

// CreateRouter initializes and returns a new instance of Router.
func CreateRouter(rootController Controller) Router {
	router := Router{}
	router.RouteTree = createNode(Slash)
	router.RouteTree.TemplateRoute = Slash
	router.RouteTree.Controller = rootController
	return router
}

// AddController adds a the passed controller on the routes returned by the
// controllers GetRoutes method.
func (router *Router) AddController(controller Controller) error {
	var err error
	for _, route := range controller.GetRoutes() {
		err = router.AddRoute(route, controller)
		if err != nil {
			break
		}
	}
	return err
}

func (node *RouteNode) insertRoute(route []string) *RouteNode {
	pathPart := route[0]
	if strings.HasPrefix(pathPart, stringutil.DOpen) {
		pathPart = WildCard
	}

	// check if the root of the route is in the SubRoutes
	v, ok := node.SubRoutes[pathPart]
	if !ok {
		// no node yet so create a new one
		v = createNode(pathPart)
		node.SubRoutes[v.Value] = v
	}

	if len(route) > 1 {
		// insert the rest of the route into the sub node.
		return v.insertRoute(route[1:])
	}

	return v
}

// AddRoute adds a Controller at the specified route. This will not add
// the routes defined on the controller. Just the route passed.
func (router *Router) AddRoute(route string, controller Controller) error {
	var err error
	var node *RouteNode
	// Trim off any leading or trailing slashes
	route = strings.TrimSpace(route)
	splitRoute := strings.Split(route, Slash)
	// make sure the route does not have any silly things like trailing or
	// double slashes
	if len(stringutil.Clean(splitRoute)) != len(splitRoute) {
		return fmt.Errorf("Invalid route format '%s'. Remove leading, "+
			"trailing, and double slashes", route)
	}
	node = router.RouteTree.insertRoute(splitRoute)
	if node.Controller == nil {
		node.Controller = controller
		node.TemplateRoute = route
	} else {
		err = fmt.Errorf("controller already present at route '%s' (%s)",
			route, node.TemplateRoute)
	}
	router.RegisteredRoutes = append(router.RegisteredRoutes, route)
	return err
}

func (node *RouteNode) find(path []string) *RouteNode {
	var foundNode *RouteNode
	if len(path) > 0 {
		subnode, ok := node.SubRoutes[path[0]]
		// if there is no sub route below this node that matches the head
		// of the slice, check for wildcard
		if !ok {
			subnode, ok = node.SubRoutes[WildCard]
		}
		// if we found either subnode, call find on it
		if ok {
			foundNode = subnode.find(path[1:])
		}
	} else {
		foundNode = node
	}
	return foundNode
}

// FindRouteForPath returns the controller that is currently assigned to
// the passed route. If no controller's route matches the passed path an error
// will be returned.
func (router *Router) FindRouteForPath(path string) (*RouteNode, error) {
	var err error
	var node *RouteNode

	splitPath := strings.Split(path, Slash)

	if path == "" {
		// path is empty, return the root node
		node = router.RouteTree
	} else {
		// otherwise locate the route
		node = router.RouteTree.find(splitPath)
	}
	// if the node or the Controller on the node is nil we didn't
	// locate a route that matched the path (non-terminal).
	if node == nil || node.Controller == nil {
		err = fmt.Errorf("No route defined for path '%s'", path)
	}
	return node, err
}
