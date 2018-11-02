package main

import (
	qcontrollers "github.com/Kasita-Inc/quimby/controllers"
	"github.com/Kasita-Inc/quimby/example/controllers"
	"github.com/Kasita-Inc/quimby/http"
)

func main() {
	rootController := &qcontrollers.HealthCheckController{}
	server := http.CreateRESTServer(":8080", rootController)
	server.Router.AddController(&qcontrollers.HealthCheckController{})
	server.Router.AddController(&controllers.ResourceController{})
	server.Router.AddController(&controllers.EchoController{})

	// API Controllers
	server.Router.AddController(&controllers.APIController{})
	storage := controllers.NewWidgetStorage()
	server.Router.AddController(controllers.NewWidgetController(storage))
	server.Router.AddController(controllers.NewWidgetsController(storage))

	server.ListenAndServe()
}
