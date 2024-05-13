package main

import (
	"fmt"
	"log"

	"github.com/brain-dev-null/gosocks/http"
	"github.com/brain-dev-null/gosocks/server"
	"github.com/brain-dev-null/gosocks/websocket"
)

func main() {
	srv := server.NewServer(8080)
	routes := server.NewRouter()
	websocketEchoHandler := websocket.WsHandler{
		OnOpen: func(conn websocket.WsConnection) { log.Println("Connection opened") },
		OnMessage: func(wme websocket.WsMessageEvent, wc websocket.WsConnection) {
			msg := string(wme.Data)
			log.Printf("Received: %s\n", msg)
		},
		OnClose: func(wce websocket.WsCloseEvent, wc websocket.WsConnection) {
			log.Printf("Closed: %d(%s) clean=%t\n", wce.Code, wce.Reason, wce.WasClean)
		},
		OnError: func(err error, conn websocket.WsConnection) {
			log.Printf("Error: %s\n", err.Error())
		},
	}
	websocketEcho := websocket.NewWsConnection(websocketEchoHandler)
	routes.AddRoute("/greet", echo)
	routes.AddWebSocket("/wstest", websocketEcho)
	srv.SetRoutes(routes)
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
