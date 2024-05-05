package main

import (
	"log"

	"github.com/brain-dev-null/gosocks/http"
	"github.com/brain-dev-null/gosocks/server"
)

func main() {
	srv := server.NewServer(8080)
	routes := http.NewRouter()
	srv.SetHttpRoutes(&routes)
	err := srv.Start()
	if err != nil {
		log.Panicf("error: %v", err)
	}

	for {
	}
}
