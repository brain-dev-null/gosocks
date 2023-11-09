package main

import (
	"log"

	"github.com/brain-dev-null/gosocks/server"
)

func demoHandler(req *server.Request) server.Response {
	log.Printf("Request: %s", *req)
	return server.Response{}
}

func main() {
	server, err := server.BuildServer(8000)
	if err != nil {
		log.Fatal(err)
	}

	server.SetHandler(demoHandler)

	err = server.Start()

	if err != nil {
		log.Fatal(err)
	}
}
