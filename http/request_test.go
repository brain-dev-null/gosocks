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
    expectedHeaders := map[string]string {
        "Host": "localhost:8080",
        "User-Agent": "curl/8.7.1",
        "Accept": "*/*",
        "Content-Length": "11",
        "Content-Type": "text/plain",
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

    if request.Path != "/foo/bar" {
        t.Errorf("request.Path is not /foo/bar. got=%s", request.Path)
    }

    if request.Protocol!= "HTTP/1.1" {
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
        t.Errorf("request content is not [%s]. got=%s", expectedContent, string(expectedContent))
    }
}
