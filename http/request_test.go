package http

import (
	"slices"
	"strings"
	"testing"
)

func TestParseHttpRequest(t *testing.T) {
	rawRequest := `POST /foo/bar HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.7.1
Accept: */*
Content-Length: 11
Content-Type: text/plain

hello world
`
	expectedHeaders := map[string]string{
		"Host":           "localhost:8080",
		"User-Agent":     "curl/8.7.1",
		"Accept":         "*/*",
		"Content-Length": "11",
		"Content-Type":   "text/plain",
	}
	expectedContent := "hello world"

	requestReader := strings.NewReader(rawRequest)

	request, err := ParseHttpRequest(requestReader)

	if err != nil {
		t.Fatalf("failed to parse http request: %v", err)
	}

	if request.Method != "POST" {
		t.Errorf("request.Method is not POST. got=%s", request.Method)
	}

	if request.FullPath != "/foo/bar" {
		t.Errorf("request.Path is not /foo/bar. got=%s", request.FullPath)
	}

	if request.Protocol != "HTTP/1.1" {
		t.Errorf("request.Protocol is not HTTP/1.1. got=%s", request.Protocol)
	}

	if len(request.Headers) != len(expectedHeaders) {
		t.Fatalf("request.Headers length is not %d. got=%d",
			len(expectedHeaders), len(request.Headers))
	}

	for headerName, headerValue := range expectedHeaders {
		value, found := request.Headers[headerName]
		if !found {
			t.Errorf("request headers missing header %s", headerName)
			continue
		}

		if value != headerValue {
			t.Errorf("header value for %s is not %s. got=%s",
				headerName, headerValue, value)
		}
	}

	if !slices.Equal(request.Content, []byte(expectedContent)) {
		t.Errorf("request content is not [%s] (%d bytes). got=%s (%d bytes)",
			expectedContent, len(expectedContent), string(request.Content), len(request.Content))
	}
}

func TestParseEmptyHttpRequest(t *testing.T) {
	rawRequest := `POST /foo/bar HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.7.1
Accept: */*
Content-Length: 0
Content-Type: text/plain
`
	expectedHeaders := map[string]string{
		"Host":           "localhost:8080",
		"User-Agent":     "curl/8.7.1",
		"Accept":         "*/*",
		"Content-Length": "0",
		"Content-Type":   "text/plain",
	}
	expectedContent := ""

	requestReader := strings.NewReader(rawRequest)

	request, err := ParseHttpRequest(requestReader)

	if err != nil {
		t.Fatalf("failed to parse http request: %v", err)
	}

	if request.Method != "POST" {
		t.Errorf("request.Method is not POST. got=%s", request.Method)
	}

	if request.FullPath != "/foo/bar" {
		t.Errorf("request.Path is not /foo/bar. got=%s", request.FullPath)
	}

	if request.Protocol != "HTTP/1.1" {
		t.Errorf("request.Protocol is not HTTP/1.1. got=%s", request.Protocol)
	}

	if len(request.Headers) != len(expectedHeaders) {
		t.Fatalf("request.Headers length is not %d. got=%d",
			len(expectedHeaders), len(request.Headers))
	}

	for headerName, headerValue := range expectedHeaders {
		value, found := request.Headers[headerName]
		if !found {
			t.Errorf("request headers missing header %s", headerName)
			continue
		}

		if value != headerValue {
			t.Errorf("header value for %s is not %s. got=%s",
				headerName, headerValue, value)
		}
	}

	if !slices.Equal(request.Content, []byte(expectedContent)) {
		t.Errorf("request content is not [%s] (%d bytes). got=%s (%d bytes)",
			expectedContent, len(expectedContent), string(request.Content), len(request.Content))
	}
}

func TestGetRequestPath(t *testing.T) {
	expectedPath := "/foo/bar/baz"
	request := HttpRequest{
		Method:   "GET",
		FullPath: expectedPath + "?x=1&y=abc",
		Protocol: "HTTP/1.1",
		Headers:  map[string]string{},
		Content:  []byte{},
	}

	path := request.Path()

	if path != expectedPath {
		t.Fatalf("path was not %s. got=%s", expectedPath, path)
	}
}

func TestGetQueryParams(t *testing.T) {
	request := HttpRequest{
		Method:   "GET",
		FullPath: "/foo/bar?a=1&b=2&c=hello&cab",
		Protocol: "HTTP/1.1",
		Headers:  map[string]string{},
		Content:  []byte{}}

	expectedParams := map[string]string{
		"a": "1",
		"b": "2",
		"c": "hello"}

	params := request.GetQueryParams()

	if len(params) > len(expectedParams) {
		t.Errorf("returned more params then there should. expected=%d, got=%d",
			len(expectedParams), len(params))
	}

	for expName, expValue := range expectedParams {
		param, exists := params[expName]
		if !exists {
			t.Errorf("expected parameter %s not found in params", expName)
			continue
		}

		if param != expValue {
			t.Errorf("mismatching param value. expected=%s, got=%s", expValue, param)
		}
	}
}
