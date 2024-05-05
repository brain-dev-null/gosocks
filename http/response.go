package http

import (
	"bytes"
	"fmt"
)

const CLRF = "\r\n"

var reasonPhrases = map[int]string{
	200: "OK",
	201: "Created",

	400: "Bad Request",
	401: "Unauthorized",
	403: "Not Allowed",
	404: "Not Found",

	500: "Internal Server Error",
}

type HttpResponse struct {
	StatusCode int
	Headers    map[string]string
	Content    []byte
}

func (response HttpResponse) Serialize() []byte {
	var buffer bytes.Buffer

	status := getStatus(response.StatusCode)
	statusLine := fmt.Sprintf("HTTP/1.1 %s", status)

	buffer.WriteString(statusLine)
	buffer.WriteString(CLRF)

	for headerName, headerValue := range response.Headers {
		headerLine := fmt.Sprintf("%s: %s", headerName, headerValue)
		buffer.WriteString(headerLine)
		buffer.WriteString(CLRF)
	}

	buffer.WriteString(CLRF)
	buffer.Write(response.Content)

	return buffer.Bytes()
}

func getStatus(statusCode int) string {
	phrase, exists := reasonPhrases[statusCode]
	if !exists {
		return fmt.Sprintf("%d", statusCode)
	}
	return fmt.Sprintf("%d %s", statusCode, phrase)
}
