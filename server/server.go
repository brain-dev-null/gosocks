package server

import (
	"fmt"
	"log"
	"net"

	"github.com/brain-dev-null/gosocks/http"
)

type Server interface {
	Start() error
	Stop()
	SetHttpRoutes(router *http.Router)
}

type gosocksServer struct {
	port       int
	httpRouter *http.Router
	running    bool
}

func NewServer(port int) Server {
	return &gosocksServer{
		port:       port,
		httpRouter: nil,
		running:    false}
}

func (server *gosocksServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", server.port))
	if err != nil {
		return fmt.Errorf("Error creating listener: %w", err)
	}
	server.running = true
	go server.runLoop(listener)
	log.Printf("Server started. Listening on Port %d", server.port)
	return nil
}

func (server *gosocksServer) Stop() {
	server.running = false
}

func (server *gosocksServer) SetHttpRoutes(router *http.Router) {
	server.httpRouter = router
}

func (server *gosocksServer) runLoop(listener net.Listener) {
	for {
		if !server.running {
			break
		}
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
		}

		go server.handleConnection(conn)
	}
}

func (sever *gosocksServer) handleConnection(conn net.Conn) {
	request, err := http.ParseHttpRequest(conn)
	log.Printf("\n%s\n", request.String())

	if err != nil {
		log.Printf("Failed to accept connection: %v\n", err)
	}

	conn.Close()
}
