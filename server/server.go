package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/brain-dev-null/gosocks/http"
)

type Server interface {
	Start() error
	Stop()
	SetHttpRoutes(router http.Router)
}

type gosocksServer struct {
	port       int
	httpRouter http.Router
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

func (server *gosocksServer) SetHttpRoutes(router http.Router) {
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

	start := time.Now()

	response, err := sever.handleRequest(request)

	duration := time.Now().Sub(start).Microseconds()

	if err != nil {
		if httpError, ok := err.(http.HttpError); ok {
			log.Printf("Error: %s", httpError.Message)
			response = httpError.ToResponse()
		} else {
			log.Printf("Error: %v", err)
			response = http.InternalServerError("").ToResponse()
		}
	}

	logString := fmt.Sprintf(
		"%s %s %d %d",
		request.Method,
		request.FullPath,
		response.StatusCode,
		duration)

	log.Println(logString)

	postProcessResponse(&response)

	serializedResponse := response.Serialize()

	conn.Write(serializedResponse)

	conn.Close()
}

func (server *gosocksServer) handleRequest(request http.HttpRequest) (http.HttpResponse, error) {
	handle, err := server.httpRouter.Route(request)
	if err != nil {
		return http.HttpResponse{}, err
	}

	response, err := handle(request)
	if err != nil {
		return http.HttpResponse{}, err
	}

	return response, nil
}

func postProcessResponse(response *http.HttpResponse) {
	contentLength := len(response.Content)
	response.Headers["Content-Length"] = fmt.Sprintf("%d", contentLength)
}
