package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type HttpRequest struct {
	Method   string
	FullPath string
	Protocol string
	Headers  map[string]string
	Content  []byte
}

func (request HttpRequest) Path() string {
	cleanPath, _, _ := strings.Cut(request.FullPath, "?")
	return cleanPath
}

func (request HttpRequest) GetQueryParams() map[string]string {
	_, paramString, found := strings.Cut(request.FullPath, "?")
	if !found {
		return map[string]string{}
	}

	params := map[string]string{}
	paramElems := strings.Split(paramString, "&")
	for _, paramElem := range paramElems {
		paramName, paramValue, found := strings.Cut(paramElem, "=")
		if !found {
			continue
		}

		params[paramName] = paramValue
	}

	return params
}

func (request HttpRequest) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf(
		"%s %s %s\n",
		request.Method,
		request.FullPath,
		request.Protocol))

	for headerName, headerValue := range request.Headers {
		buffer.WriteString(fmt.Sprintf(
			"%s: %s\n",
			headerName,
			headerValue))
	}

	if len(request.Content) > 0 {
		buffer.WriteString(fmt.Sprintf(
			"\n%s\n",
			string(request.Content)))
	}

	return buffer.String()
}

func parseRequestMethod(reader *bufio.Reader) (string, error) {
	method, err := reader.ReadString(' ')
	if err != nil {
		return "", err
	}

	method = strings.TrimSpace(method)

	if !isValidRequestMethod(method) {
		return "", fmt.Errorf("request method [%s] is invalid", method)
	}

	return method, nil
}

func isValidRequestMethod(requestMethod string) bool {
	switch requestMethod {
	case "GET":
		return true
	case "POST":
		return true
	case "PUT":
		return true
	case "PATCH":
		return true
	case "OPTIONS":
		return true
	case "HEAD":
		return true
	case "DELETE":
		return true
	case "CONNECT":
		return true
	case "TRACE":
		return true
	}

	return false
}

func parseRequestPath(reader *bufio.Reader) (string, error) {
	path, err := reader.ReadString(' ')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(path)), nil
}

func parseRequestProtocol(reader *bufio.Reader) (string, error) {
	protocol, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	protocol = strings.TrimSpace(protocol)

	if protocol != "HTTP/1.1" {
		return "", fmt.Errorf("protocol [%s] is not supported", protocol)
	}

	return protocol, nil
}

func parseRequestHeaders(reader *bufio.Reader) (map[string]string, error) {
	headers := make(map[string]string)

	line, err := readHeaderLine(reader)
	if err != nil {
		return nil, err
	}

	for len(line) != 0 {
		headerName, headerValue, err := parseHeaderLine(line)
		if err != nil {
			return nil, err
		}

		headers[headerName] = headerValue

		line, err = readHeaderLine(reader)
		if err != nil {
			return nil, err
		}
	}

	return headers, nil
}

func readHeaderLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')

	if err != nil {
		if err == io.EOF {
			return "", nil
		}
		return "", fmt.Errorf("failed to read header line: %w", err)
	}
	line = strings.TrimSpace(line)
	return line, nil
}

func parseHeaderLine(line string) (string, string, error) {
	headerName, headerValue, ok := strings.Cut(line, ":")

	if !ok {
		return "", "", fmt.Errorf("header line malformed: %s", line)
	}

	return strings.TrimSpace(headerName), strings.TrimSpace(headerValue), nil
}

func parseContentLength(headers map[string]string) (int, error) {
	rawContentLength, exists := headers["Content-Length"]
	if !exists {
		return 0, nil
	}

	contentLength, err := strconv.Atoi(rawContentLength)
	if err != nil {
		return -1, fmt.Errorf("failed to parse Content-Length header: %w", err)
	}

	if contentLength < 0 {
		return -1, fmt.Errorf("invalid Content-Length: %d", contentLength)
	}

	return contentLength, nil
}

func parseRequestContent(reader *bufio.Reader, size int) ([]byte, error) {
	content := make([]byte, size)

	_, err := reader.Read(content)

	if err != nil {
		return nil, fmt.Errorf("failed to read request content: %w", err)
	}

	if len(content) == 0 {
		return []byte{}, nil
	}

	return content, nil
}

func ParseHttpRequest(rawReader io.Reader) (HttpRequest, error) {
	request := HttpRequest{}
	bufReader := bufio.NewReader(rawReader)

	method, err := parseRequestMethod(bufReader)
	if err != nil {
		return request, fmt.Errorf("failed to parse request method: %w", err)
	}

	path, err := parseRequestPath(bufReader)
	if err != nil {
		return request, fmt.Errorf("failed to parse request path: %w", err)
	}

	protocol, err := parseRequestProtocol(bufReader)
	if err != nil {
		return request, fmt.Errorf("failed to parse request protocol: %w", err)
	}

	headers, err := parseRequestHeaders(bufReader)
	if err != nil {
		return request, fmt.Errorf("failed to parse request headers: %w", err)
	}

	contentLength, err := parseContentLength(headers)
	if err != nil {
		return request, fmt.Errorf("failed to parse content length: %w", err)
	}

	content, err := parseRequestContent(bufReader, contentLength)
	if err != nil {
		return request, fmt.Errorf("failed to read request content: %w", err)
	}

	request.Method = method
	request.FullPath = path
	request.Protocol = protocol
	request.Headers = headers
	request.Content = content

	return request, nil
}
