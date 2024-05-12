package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/brain-dev-null/gosocks/http"
)

const MAGIC_NUMBER = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func Handshake(handshakeRequest http.HttpRequest) (http.HttpResponse, error) {
	handshakeResponse := http.HttpResponse{StatusCode: 101, Content: []byte{}}

	if handshakeRequest.Method != "GET" {
		return handshakeResponse, fmt.Errorf(
			"handshake error: unexpected method. got=%s",
			handshakeRequest.Method)
	}

	err := expectHeaderValue("Upgrade", "websocket", handshakeRequest)
	if err != nil {
		return handshakeResponse, err
	}

	err = expectHeaderValue("Connection", "Upgrade", handshakeRequest)
	if err != nil {
		return handshakeResponse, err
	}

	secWebsocketKey, err := expectHeaderValuePresent("Sec-WebSocket-Key", handshakeRequest)
	if err != nil {
		return handshakeResponse, err
	}

	err = expectHeaderValue("Sec-WebSocket-Version", "13", handshakeRequest)
	if err != nil {
		return handshakeResponse, err
	}

	handshakeResponse.Headers = generateResponseHeaders(secWebsocketKey)

	return handshakeResponse, nil
}

func expectHeaderValue(headerName string, expectedValue string, request http.HttpRequest) error {
	value, err := expectHeaderValuePresent(headerName, request)

	if err != nil {
		return err
	}

	if value != expectedValue {
		return fmt.Errorf(
			"handshake error: unexpected %s value. got=%s", headerName, value)
	}

	return nil
}

func expectHeaderValuePresent(headerName string, request http.HttpRequest) (string, error) {
	value, exists := request.Headers[headerName]
	if !exists {
		return "", fmt.Errorf("handshake error: missing Sec-WebSocket-Version header")
	}

	return value, nil
}

func generateResponseHeaders(key string) map[string]string {
	return map[string]string{
		"Upgrade":              "websocket",
		"Connection":           "upgrade",
		"Sec-WebSocket-Accept": generateAcceptHeader(key)}
}

func generateAcceptHeader(key string) string {
	magicValue := key + MAGIC_NUMBER
	h := sha1.New()
	io.WriteString(h, magicValue)
	sha1Hash := h.Sum(nil)
	base64Encoded := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(sha1Hash)

	return base64Encoded
}
