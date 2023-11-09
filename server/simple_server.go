package server

import (
	"fmt"
	"log"
	"net"
)

type simpleServer struct {
	serverPort    port
	serverIp      net.IP
	serverHandler RequestHandler
}

func (s *simpleServer) SetHandler(handler RequestHandler) {
	s.serverHandler = handler
}

func writeResponse(conn *net.Conn, response Response) error {
	log.Printf("TODO: Implement writing response")
	return nil
}

func (s *simpleServer) Start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:"+fmt.Sprint(s.serverPort))

	if err != nil {
		return err
	}

	defer listener.Close()

	for {
		err = s.handleNextRequest(listener)
		if err != nil {
			fmt.Printf("Error during request Handling: %s", err.Error())
		}
	}
}

func (s simpleServer) handleNextRequest(listener net.Listener) error {
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("Error establishing new connection: %w", err)
	}

	defer conn.Close()

	request, err := readRequest(&conn)
	if err != nil {
		return fmt.Errorf("Error reading request: %w", err)
	}

	response := s.serverHandler(&request)

	err = writeResponse(&conn, response)
	if err != nil {
		return fmt.Errorf("Error writing response: %w", err)
	}

	err = conn.Close()
	if err != nil {
		return fmt.Errorf("Error closing the connection: %w\n", err)
	}

	return nil
}
