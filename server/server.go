package server

import "net"

type port int16

type Response struct{}

type RequestHandler func(*Request) Response

type Server interface {
	SetHandler(RequestHandler)
	Start() error
}

func BuildServer(portNumber int16) (Server, error) {
	return createSimpleServer(portNumber)
}

func createSimpleServer(portNumber int16) (*simpleServer, error) {
	return &simpleServer{
		serverPort: port(portNumber),
		serverIp:   net.IPv4(127, 0, 0, 1),
	}, nil
}
