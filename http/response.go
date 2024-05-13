package http

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const CLRF = "\r\n"

const CONTENT_TYPE_PLAIN = "text/plain"
const CONTENT_TYPE_JSON = "application/json"

var reasonPhrases = map[int]string{
	100: "Continue",
	101: "Switching Protocols",

	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",

	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	307: "Temporary Redirect",

	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	426: "Upgrade Required",

	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
}

type HttpResponse struct {
	StatusCode int
	Headers    map[string]string
	Content    []byte
}

func NewPlainTextResponse(content string, statusCode int) HttpResponse {
	headers := map[string]string{"Content-Type": "text/plain"}
	return HttpResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Content:    []byte(content)}
}

func NewJsonResponse(content interface{}, statusCode int) (HttpResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}
	serializedContent, err := json.Marshal(content)
	if err != nil {
		return HttpResponse{}, err
	}

	response := HttpResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Content:    serializedContent}

	return response, nil
}

func (response HttpResponse) Serialize() []byte {
	var buffer bytes.Buffer

	status := GetStatus(response.StatusCode)
	statusLine := fmt.Sprintf("HTTP/1.1 %s", status)

	buffer.WriteString(statusLine)
	buffer.WriteString(CLRF)

	contentLength := len(response.Content)
	response.Headers["Content-Length"] = fmt.Sprintf("%d", contentLength)

	for headerName, headerValue := range response.Headers {
		headerLine := fmt.Sprintf("%s: %s", headerName, headerValue)
		buffer.WriteString(headerLine)
		buffer.WriteString(CLRF)
	}

	buffer.WriteString(CLRF)
	buffer.Write(response.Content)

	return buffer.Bytes()
}

func GetStatus(statusCode int) string {
	phrase, exists := reasonPhrases[statusCode]
	if !exists {
		return fmt.Sprintf("%d", statusCode)
	}
	return fmt.Sprintf("%d %s", statusCode, phrase)
}
