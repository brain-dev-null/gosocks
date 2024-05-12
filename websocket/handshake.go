package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"

	"github.com/brain-dev-null/gosocks/http"
)

const MAGIC_NUMBER = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func Handshake(handshakeRequest http.HttpRequest) (http.HttpResponse, error) {
	handshakeResponse := http.HttpResponse{}
	if handshakeRequest.Method != "GET" {
		return handshakeResponse, fmt.Errorf(
			"handshake error: unexpected method. got=%s",
			handshakeRequest.Method)
	}

	updgradeHeader, exists := handshakeRequest.Headers["Upgrade"]
	if !exists {
		return handshakeResponse, fmt.Errorf(
			"handshake error: missing Upgrade header")
	}

	if updgradeHeader != "websocket" {
		return handshakeResponse, fmt.Errorf(
			"handshake error: unexpexted Upgrade header value. got=%s",
			updgradeHeader)
	}

	connectionHeader, exists := handshakeRequest.Headers["Connection"]
	if !exists {
		return handshakeResponse, fmt.Errorf(
			"handshake error: missing Connection header")
	}

	if connectionHeader != "Upgrade" {
		return handshakeResponse, fmt.Errorf(
			"handshake error: unexpected Connection header value. got=%s",
			connectionHeader)
	}

	secWebsocketKey, exists := handshakeRequest.Headers["Sec-WebSocket-Key"]
	if !exists {
		return handshakeResponse, fmt.Errorf(
			"handshake error: missing Sec-WebSocket-Key header")
	}

	secWebsocketVersion, exists := handshakeRequest.Headers["Sec-WebSocket-Version"]
	if !exists {
		return handshakeResponse, fmt.Errorf(
			"handshake error: missing Sec-WebSocket-Version header")
	}
	if secWebsocketVersion != "13" {
		return handshakeResponse, fmt.Errorf(
			"handshake error: unexptected Sec-WebSocket-Version value. got=%s",
			secWebsocketVersion)
	}

	responseHeaders := map[string]string{
		"Upgrade":              "websocket",
		"Connection":           "upgrade",
		"Sec-WebSocket-Accept": generateAcceptHeader(secWebsocketKey)}

	handshakeResponse.StatusCode = 101
	handshakeResponse.Headers = responseHeaders
	handshakeRequest.Content = []byte{}

	return handshakeResponse, nil
}

func generateAcceptHeader(key string) string {
	log.Printf("Key: %s", key)
	magicValue := key + MAGIC_NUMBER
	log.Printf("Magic Value: %s", magicValue)
	h := sha1.New()
	io.WriteString(h, magicValue)
	sha1Hash := h.Sum(nil)
	log.Printf("Sha1 Hash (%d): %x", len(sha1Hash), sha1Hash)
	base64Encoded := base64.StdEncoding.WithPadding(base64.StdPadding).EncodeToString(sha1Hash)
	log.Printf("Base64 Encoded: %s", base64Encoded)

	return base64Encoded
}
