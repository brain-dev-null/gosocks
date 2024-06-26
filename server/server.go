package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/brain-dev-null/gosocks/http"
	"github.com/brain-dev-null/gosocks/websocket"
)

type Server interface {
	Start() error
	Stop()
	SetRoutes(router Router)
}

type gosocksServer struct {
	port       int
	httpRouter Router
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

func (server *gosocksServer) SetRoutes(router Router) {
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
	start := time.Now()
	request, reader, err := http.ParseHttpRequest(conn)
	if err != nil {
		log.Printf("failed to parse http request: %v", err)
		response := http.BadRequest("").ToResponse().Serialize()
		conn.Write(response)
		conn.Close()
	}

	if isWebSocketUpgradeRequest(request) {
		sever.handleWebsocket(request, conn, reader, start)
		return
	}

	response, err := sever.handleRequest(request)

	duration := time.Now().Sub(start)

	if err != nil {
		if httpError, ok := err.(http.HttpError); ok {
			log.Printf("Error: %s", httpError.Message)
			response = httpError.ToResponse()
		} else {
			log.Printf("Error: %v", err)
			response = http.InternalServerError("").ToResponse()
		}
	}

	accessLog(request, response, duration)

	serializedResponse := response.Serialize()

	conn.Write(serializedResponse)

	conn.Close()
}

func isWebSocketUpgradeRequest(request http.HttpRequest) bool {
	if request.Method != "GET" {
		return false
	}

	updgradeHeader, exists := request.Headers["Upgrade"]
	if !exists {
		return false
	}

	if updgradeHeader != "websocket" {
		return false
	}

	connectionHeader, exists := request.Headers["Connection"]
	if !exists {
		return false
	}

	if connectionHeader != "Upgrade" {
		return false
	}

	return true
}

func (server *gosocksServer) handleRequest(request http.HttpRequest) (http.HttpResponse, error) {
	handle, err := server.httpRouter.RouteHttpRequest(request)
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

func (server *gosocksServer) handleWebsocket(initialRequest http.HttpRequest, conn net.Conn, reader *bufio.Reader, start time.Time) {
	handle, err := server.httpRouter.RouteWebSocket(initialRequest)
	if err != nil {
		log.Printf("not found: %s", initialRequest.Path())
		return
	}

	handhakeResponse, err := websocket.Handshake(initialRequest)
	if err != nil {
		log.Printf("handshake error: %v", err)
		response := http.BadRequest(err.Error())
		conn.Write(response.ToResponse().Serialize())
		return
	}
	_, err = conn.Write(handhakeResponse.Serialize())

	duration := time.Now().Sub(start)

	if err != nil {
		log.Printf("error sending handshake response: %v", err)
		conn.Close()
		return
	}

	accessLog(initialRequest, handhakeResponse, duration)

	go handle(conn, reader)
}

func accessLog(request http.HttpRequest, response http.HttpResponse, duration time.Duration) {
	log.Printf(
		"%s %s %s %d\n",
		request.Method,
		request.FullPath,
		http.GetStatus(response.StatusCode),
		duration.Microseconds())
}
