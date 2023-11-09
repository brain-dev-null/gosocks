package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type RequestMethod string
type UrlPath string
type RequestParameters map[string]string
type RequestHeaders map[string]string

type Request struct {
	method     RequestMethod
	path       UrlPath
	parameters RequestParameters
	headers    RequestHeaders
}

func readRequest(conn *net.Conn) (Request, error) {
	reader := bufio.NewReader(*conn)

	request := Request{}

	err := parseFirstLine(reader, &request)
	if err != nil {
		return Request{}, fmt.Errorf("Failed to read first request line: %w", err)
	}

	err = parseHeaders(reader, &request)
	if err != nil {
		return Request{}, fmt.Errorf("Failed to parse headers: %w", err)
	}

	return request, nil
}

func parseFirstLine(reader *bufio.Reader, request *Request) error {
	line, err := readLine(reader)

	if err != nil {
		return fmt.Errorf("Failed to read first line: %w", err)
	}

	elements := strings.Split(line, " ")

	if len(elements) != 3 {
		return fmt.Errorf("Invalid first header line: %s", elements)
	}

	method := elements[0]
	completePath := elements[1]
	if !isValidMethod(method) {
		return fmt.Errorf("Method '%s' is not valid", method)
	}

	urlPath, parameters := parseCompletePath(completePath)

	request.method = RequestMethod(method)
	request.path = UrlPath(urlPath)
	request.parameters = RequestParameters(parameters)

	return nil
}

func parseHeaders(reader *bufio.Reader, request *Request) error {
	headers := map[string]string{}
	for {
		nextLine, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("Failed to read header line: %w", err)
		}

		if len(nextLine) == 0 {
			break
		}

		elements := strings.SplitN(nextLine, ": ", 2)

		if len(elements) == 1 {
			headers[elements[0]] = ""
		} else {
			headers[elements[0]] = elements[1]
		}
	}
	request.headers = headers
	return nil
}

func parseCompletePath(completePath string) (string, map[string]string) {
	elements := strings.SplitN(completePath, "?", 2)

	if len(elements) == 1 {
		return completePath, map[string]string{}
	}

	parameters := parseParameters(elements[1])

	return elements[0], parameters
}

func parseParameters(parameters string) map[string]string {
	parameterElements := strings.Split(parameters, "&")
	result := map[string]string{}

	for _, paramater := range parameterElements {
		elements := strings.SplitN(paramater, "=", 2)
		if len(elements) == 2 {
			result[elements[0]] = elements[1]
		}
	}

	return result
}

func isValidMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PATCH", "PUT", "DELETE"}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}

	return false
}

func readLine(reader *bufio.Reader) (string, error) {
	lineBytes, err := reader.ReadBytes('\n')

	if err != nil {
		return "", fmt.Errorf("Failed to read line: %w", err)
	}

	return string(lineBytes[:len(lineBytes)-2]), nil
}
