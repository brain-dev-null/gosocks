package websocket

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const STATUS_INTERNAL_SERVER_ERROR uint16 = 1011

const STATE_OPEN = "open"
const STATE_CLOSING = "closing"
const STATE_CLOSED = "closed"

type WsHandler struct {
	OnOpen    func(conn WsConnection)
	OnMessage func(event WsMessageEvent, conn WsConnection)
	OnClose   func(event WsCloseEvent, conn WsConnection)
	OnError   func(err error, conn WsConnection)
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
	Ping() error
	Pong(pingData []byte) error
	Close(statusCode uint16, reason string) error
	SendText(text string) error
	SendBinary(data []byte) error
}

type wsConnection struct {
	reader      *bufio.Reader
	connection  net.Conn
	partialData []byte
	handler     WsHandler
	isClient    bool
	state       string
}

func NewWsConnection(handler WsHandler) func(net.Conn, *bufio.Reader) {
	return func(conn net.Conn, reader *bufio.Reader) {
		connection := wsConnection{
			reader:      reader,
			connection:  conn,
			partialData: nil,
			handler:     handler,
			isClient:    false,
			state:       STATE_OPEN}
		handler.OnOpen(&connection)
		connection.run()
	}
}

func (wsConn *wsConnection) Close(statusCode uint16, reason string) error {
	defer wsConn.connection.Close()

	if wsConn.state == STATE_CLOSING {
		wsConn.state = STATE_CLOSED
		event := WsCloseEvent{Code: statusCode, Reason: reason, WasClean: true}
		go wsConn.handler.OnClose(event, wsConn)
		return nil
	}

	if wsConn.state == STATE_CLOSED {
		return nil
	}

	wsConn.state = STATE_CLOSING
	closeFrame := NewCloseFrame(statusCode, reason, wsConn.isClient)
	_, err := wsConn.connection.Write(closeFrame.Serialize())

	if err != nil {
		err := fmt.Errorf("failed to send close frame: %w", err)
		event := WsCloseEvent{Code: statusCode, Reason: reason, WasClean: false}
		log.Println(err.Error())
		go wsConn.handler.OnClose(event, wsConn)
		return err
	}

	event := WsCloseEvent{Code: statusCode, Reason: reason, WasClean: true}
	go wsConn.handler.OnClose(event, wsConn)

	return nil
}

func (wsConn *wsConnection) Ping() error {
	pongFrame := NewPingFrame([]byte{}, false)
	data := pongFrame.Serialize()

	_, err := wsConn.connection.Write(data)

	if err != nil {
		err := fmt.Errorf("failed to send ping frame: %w", err)
		log.Println(err.Error())
		go wsConn.handleInternalError(err)
		return err
	}

	return nil
}

func (wsConn *wsConnection) Pong(pingData []byte) error {
	pongFrame := NewPongFrame(pingData, false)
	data := pongFrame.Serialize()

	_, err := wsConn.connection.Write(data)

	if err != nil {
		log.Printf("failed to send pong frame: %v", err)
		wsConn.handleInternalError(err)
		return err
	}

	return nil
}

func (wsConn *wsConnection) SendText(text string) error {
	if wsConn.state != STATE_OPEN {
		return fmt.Errorf("connection closed")
	}
	frame := NewTextFrame(false, text).Serialize()
	_, err := wsConn.connection.Write(frame)
	if err != nil {
		err := fmt.Errorf("error during send: %w", err)
		wsConn.handleInternalError(err)
		return err
	}
	return nil
}

func (wsConn *wsConnection) SendBinary(data []byte) error {
	if wsConn.state != STATE_OPEN {
		return fmt.Errorf("connection closed")
	}
	frame := NewBinaryFrame(false, data).Serialize()
	_, err := wsConn.connection.Write(frame)
	if err != nil {
		err := fmt.Errorf("error during send: %w", err)
		wsConn.handleInternalError(err)
		return err
	}
	return nil
}

func (wsConn *wsConnection) run() {
	defer wsConn.connection.Close()
	for wsConn.state == STATE_OPEN {
		wsConn.rcvNextMsg()
	}
}

func (wsconn *wsConnection) rcvNextMsg() {
	frame, err := DeserialzeWebSocketFrame(wsconn.reader)
	if err != nil {
		err := fmt.Errorf("failed to deserialize next frame: %w", err)
		wsconn.handleInternalError(err)
		return
	}

	if isCloseFrame(frame) {
		log.Println("Received close frame")
		code := getStatusCode(frame)
		reason := getReason(frame)

		wsconn.Close(code, reason)
		return
	}

	if isPingFrame(frame) {
		wsconn.Pong(frame.Payload)
		return
	}

	if isPongFrame(frame) {
		return
	}

	if isUnfragmentedFrame(frame) {
		if wsconn.partialData != nil {
			err := fmt.Errorf(
				"expected fragmented message frame, got unfragmented one")
			wsconn.handleInternalError(err)
			return
		}
		event := WsMessageEvent{Data: frame.Payload}
		wsconn.handler.OnMessage(event, wsconn)
		return
	}

	if isFragmentedStartFrame(frame) {
		if wsconn.partialData != nil {
			err := fmt.Errorf(
				"expected continouation frame, got start frame")
			wsconn.handleInternalError(err)
			return
		}
		wsconn.partialData = frame.Payload
		return
	}

	if isFragmentedContinouationFrame(frame) {
		if wsconn.partialData == nil {
			err := fmt.Errorf(
				"expected start frame, got continouation frame")
			wsconn.handleInternalError(err)
			return
		}
		wsconn.partialData = append(wsconn.partialData, frame.Payload...)
	}

	if isFragmentedTerminationFrame(frame) {
		if wsconn.partialData == nil {
			err := fmt.Errorf(
				"expected start frame, got termination frame")
			wsconn.handleInternalError(err)
			return
		}
		fullData := append(wsconn.partialData, frame.Payload...)
		wsconn.partialData = nil
		event := WsMessageEvent{Data: fullData}
		wsconn.handler.OnMessage(event, wsconn)
		return
	}

	log.Printf("unexpected frame type: fin=%t opcode=%d",
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
	return frame.OpCode == OPCODE_CLOSE
}

func isPingFrame(frame WebSocketFrame) bool {
	return frame.OpCode == OPCODE_PING
}

func isPongFrame(frame WebSocketFrame) bool {
	return frame.OpCode == OPCODE_PONG
}

func getStatusCode(frame WebSocketFrame) uint16 {
	if len(frame.Payload) < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(frame.Payload[:2])
}

func getReason(frame WebSocketFrame) string {
	if len(frame.Payload) < 3 {
		return ""
	}
	reasonData := frame.Payload[2:]

	return string(reasonData)
}

func (wsConn *wsConnection) handleInternalError(err error) {
	go wsConn.handler.OnError(err, wsConn)
	wsConn.Close(STATUS_INTERNAL_SERVER_ERROR, "")
}
