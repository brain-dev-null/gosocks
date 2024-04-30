package http

import (
	"testing"
)

func TestSimpleRouter(t *testing.T) {
	router := NewRouter()
	expectedResponse := HttpResponse{StatusCode: 123}

	handler := func(r HttpRequest) HttpResponse {
		return expectedResponse
	}

	request := HttpRequest{Path: "/foo/bar/baz?ab=z"}

	router.AddRoute("/foo/bar/baz", handler)

	routedHandler, err := router.Route(request)

	if err != nil {
		t.Fatalf("Routing failed: %v", err)
	}

	response := routedHandler(request)

	if response.StatusCode != 123 {
		t.Fatalf("wrong response returned. expected=%+v got=%+v", expectedResponse, response)
	}
}
