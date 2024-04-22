package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

var validRequestMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"PATCH":   true,
	"OPTIONS": true,
	"HEAD":    true,
	"DELETE":  true,
	"CONNECT": true,
	"TRACE":   true,
}

type HttpRequest struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
	Content  []byte
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
	_, found := validRequestMethods[requestMethod]

	return found
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
	elems := strings.SplitN(line, ":", 2)

	if len(elems) != 2 {
		return "", "", fmt.Errorf("header line malformed: %s", line)
	}

	return strings.TrimSpace(elems[0]), strings.TrimSpace(elems[1]), nil
}

func parseRequestContent(reader *bufio.Reader) ([]byte, error) {
	var buffer bytes.Buffer

	_, err := reader.WriteTo(&buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read request content: %w", err)
	}

	content := buffer.Bytes()

	return content[:len(content)-1], nil
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

	content, err := parseRequestContent(bufReader)
	if err != nil {
		return request, fmt.Errorf("failed to read request content: %w", err)
	}

	request.Method = method
	request.Path = path
	request.Protocol = protocol
	request.Headers = headers
	request.Content = content

	return request, nil
}
