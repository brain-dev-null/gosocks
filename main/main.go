package main

import (
	"fmt"
	"log"

	"github.com/brain-dev-null/gosocks/http"
	"github.com/brain-dev-null/gosocks/server"
)

func main() {
	srv := server.NewServer(8080)
	routes := http.NewRouter()
	routes.AddRoute("/greet", echo)
	srv.SetHttpRoutes(routes)
	err := srv.Start()
	if err != nil {
		log.Panicf("error: %v", err)
	}

	for {
	}
}

func echo(request http.HttpRequest) (http.HttpResponse, error) {
	params := request.GetQueryParams()

	firstName, exists := params["first_name"]
	if !exists {
		return http.HttpResponse{}, http.BadRequest("missing param: first_name")
	}

	lastName, exists := params["last_name"]
	if !exists {
		return http.HttpResponse{}, http.BadRequest("missing param: last_name")
	}

	greeting := fmt.Sprintf("Hello, %s %s", firstName, lastName)
	responseObject := EchoResponse{
		FirstName: firstName,
		LastName:  lastName,
		Greeting:  greeting}

	return http.NewJsonResponse(&responseObject, 200)
}

type EchoResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Greeting  string `json:"greeting"`
}
