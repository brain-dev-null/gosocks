package http

import (
	"testing"
)

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
