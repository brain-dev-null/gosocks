package websocket

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
)

type WsHandler struct {
	OnOpen    func(WsConnection)
	OnMessage func(WsMessageEvent, WsConnection)
	OnClose   func(WsCloseEvent, WsConnection)
	OnError   func(WsConnection)
}

type WsMessageEvent struct {
	Data []byte
}

type WsCloseEvent struct {
	Code     uint16
	Reason   string
	WasClean bool
}

type WsConnection interface {
	Close() error
	SendText(string) error
	SendBinary([]byte) error
}

type wsConnection struct {
	reader      *bufio.Reader
	connection  net.Conn
	closed      bool
	partialData []byte
	handler     WsHandler
}

func NewWsConnection(handler WsHandler) func(net.Conn) {
	return func(conn net.Conn) {
		connection := &wsConnection{
			reader:      bufio.NewReader(conn),
			connection:  conn,
			closed:      false,
			partialData: nil,
			handler:     handler}
		go connection.run()
	}
}

func (wsConn *wsConnection) Close() error {
	// Todo implement!
	return nil
}

func (wsConn *wsConnection) SendText(text string) error {
	// Todo implement!
	return nil
}

func (wsConn *wsConnection) SendBinary(data []byte) error {
	// Todo implement!
	return nil
}

func (wsConn *wsConnection) run() {
	defer wsConn.Close()
	for !wsConn.closed {
		wsConn.rcvNextMsg()
	}
}

func (wsconn *wsConnection) rcvNextMsg() error {
	frame, err := DeserialzeWebSocketFrame(wsconn.reader)
	if err != nil {
		return err
	}

	if isCloseFrame(frame) {
		code, err := getStatusCode(frame)
		if err != nil {
			return err
		}
		reason, err := getReason(frame)
		if err != nil {
			return err
		}

		event := WsCloseEvent{Code: code, Reason: reason, WasClean: true}
		err = wsconn.connection.Close()
		if err != nil {
			event.WasClean = false
		}
		wsconn.closed = true

		go wsconn.handler.OnClose(event, wsconn)

		return err
	}

	if isUnfragmentedFrame(frame) {
		if wsconn.partialData != nil {
			return fmt.Errorf(
				"expected fragmented message frame, got unfragmented one")
		}
		event := WsMessageEvent{Data: frame.Payload}
		go wsconn.handler.OnMessage(event, wsconn)
		return nil
	}

	if isFragmentedStartFrame(frame) {
		if wsconn.partialData != nil {
			return fmt.Errorf(
				"expected continouation frame, got start frame")
		}
		wsconn.partialData = frame.Payload
		return nil
	}

	if isFragmentedContinouationFrame(frame) {
		if wsconn.partialData == nil {
			return fmt.Errorf(
				"expected start frame, got continouation frame")
		}
		wsconn.partialData = append(wsconn.partialData, frame.Payload...)
	}

	if isFragmentedTerminationFrame(frame) {
		if wsconn.partialData == nil {
			return fmt.Errorf(
				"expected start frame, got termination frame")
		}
		fullData := append(wsconn.partialData, frame.Payload...)
		wsconn.partialData = nil
		event := WsMessageEvent{Data: fullData}
		go wsconn.handler.OnMessage(event, wsconn)
		return nil
	}

	return fmt.Errorf("unexpected frame type: fin=%t opcode=%d",
		frame.Fin, frame.OpCode)
}

func isUnfragmentedFrame(msg WebSocketFrame) bool {
	return msg.Fin && msg.OpCode != 0
}

func isFragmentedStartFrame(msg WebSocketFrame) bool {
	return !msg.Fin && msg.OpCode != 0
}

func isFragmentedContinouationFrame(msg WebSocketFrame) bool {
	return !msg.Fin && msg.OpCode == 0
}

func isFragmentedTerminationFrame(msg WebSocketFrame) bool {
	return msg.Fin && msg.OpCode == 0
}

func isCloseFrame(frame WebSocketFrame) bool {
	return frame.OpCode == 8
}

func getStatusCode(frame WebSocketFrame) (uint16, error) {
	if len(frame.Payload) < 2 {
		return 0, fmt.Errorf("no status code in closing frame payload")
	}
	return binary.BigEndian.Uint16(frame.Payload[:1]), nil
}

func getReason(frame WebSocketFrame) (string, error) {
	if len(frame.Payload) == 2 {
		return "", nil
	}
	reasonData := frame.Payload[2:]

	return string(reasonData), nil
}
