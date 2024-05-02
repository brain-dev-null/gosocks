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

func TestSimpleRouterMismatch(t *testing.T) {
	router := NewRouter()
	expectedResponse := HttpResponse{StatusCode: 123}

	handler := func(r HttpRequest) HttpResponse {
		return expectedResponse
	}

	request := HttpRequest{Path: "/foo/bar?ab=z"}

	router.AddRoute("/foo/bar/baz", handler)

	_, err := router.Route(request)

	if err == nil {
		t.Fatalf("expected routing to fail")
	}

	httpError, ok := err.(HttpError)

	if !ok {
		t.Fatalf("expected routing error to be HttpError. got=%T", err)
	}

	if httpError.StatusCode != 404 {
		t.Fatalf("expected http error 404. got=%d", httpError.StatusCode)
	}
}

func TestEmptyRoute(t *testing.T) {
	router := NewRouter()
	expectedResponse := HttpResponse{StatusCode: 123}
	handler := func(r HttpRequest) HttpResponse {
		return expectedResponse
	}
	router.AddRoute("/", handler)

	request := HttpRequest{Path: "/?ab=c"}

	routedHandler, err := router.Route(request)

	if err != nil {
		t.Fatalf("routing failed: %v", err)
	}

	response := routedHandler(request)

	if response.StatusCode != 123 {
		t.Fatalf("wrong response returned. expected=%+v got=%+v", expectedResponse, response)
	}
}

func TestComplexRouter(t *testing.T) {
	router := NewRouter()

	routes := []struct {
		r      string
		status int
	}{
		{"/", 1},
		{"/foo", 2},
		{"/foo/bar", 3},
		{"/baz/bar", 4},
	}

	tests := []struct {
		r              string
		expectedMatch  bool
		expectedStatus int
	}{
		{"/?a=1", true, 1},
		{"/baz?bb=2", false, -1},
		{"/foo?a=1&b=2", true, 2},
		{"/foo/baz?a=0", false, -1},
		{"/foo/bar?a=1&b=2&c=3", true, 3},
	}
	for _, rt := range routes {
		handler := buildStatusCodeHandler(rt.status)
		router.AddRoute(rt.r, handler)
	}

	for _, tt := range tests {
		request := HttpRequest{Path: tt.r}
		handler, err := router.Route(request)

		if tt.expectedMatch {
			if err != nil {
				t.Errorf("routing failed: %v\n", err)
				continue
			}

			response := handler(request)
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code does not match expected status code %d. got=%d",
					tt.expectedStatus, response.StatusCode)
				continue
			}
		} else {
			if err == nil {
				t.Errorf("expected routing to fail: %s", tt.r)
				continue
			}

			httpError, ok := err.(HttpError)
			if !ok {
				t.Errorf("expected routing error of type HttpError. got=%T\n", httpError)
				continue
			}

			if httpError.StatusCode != 404 {
				t.Errorf("expected HttpError with code 404. got=%d", httpError.StatusCode)
				continue
			}
		}
	}
}

func buildStatusCodeHandler(statusCode int) Handler {
	return func(hr HttpRequest) HttpResponse {
		return HttpResponse{StatusCode: statusCode}
	}
}
