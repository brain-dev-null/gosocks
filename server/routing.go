package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/brain-dev-null/gosocks/http"
)

type HttpHandler func(http.HttpRequest) (http.HttpResponse, error)
type WebSocketHandler func(net.Conn)

type Router interface {
	RouteHttpRequest(request http.HttpRequest) (HttpHandler, error)
	RouteWebSocket(request http.HttpRequest) (WebSocketHandler, error)
	AddRoute(path string, handler HttpHandler) error
	AddWebSocket(path string, handler WebSocketHandler) error
}

type recursiveRouter struct {
	httpRoot      httpRoute
	websocketRoot websocketRoute
}

func (rr *recursiveRouter) RouteHttpRequest(request http.HttpRequest) (HttpHandler, error) {
	segments := strings.Split(request.Path(), "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	handler, _, matched := rr.httpRoot.match(segments)

	if !matched {
		return nil, http.ErrorNotFound(fmt.Sprintf("No route for: %s", request.Path()))
	}

	return handler, nil
}

func (rr *recursiveRouter) RouteWebSocket(request http.HttpRequest) (WebSocketHandler, error) {
	segments := strings.Split(request.Path(), "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	handler, _, matched := rr.websocketRoot.match(segments)

	if !matched {
		return nil, http.ErrorNotFound(fmt.Sprintf("No route for: %s", request.Path()))
	}

	return handler, nil
}
func (rr *recursiveRouter) AddRoute(path string, handler HttpHandler) error {
	segments := strings.Split(path, "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	return rr.httpRoot.merge(segments, handler)
}

func (rr *recursiveRouter) AddWebSocket(path string, handler WebSocketHandler) error {
	segments := strings.Split(path, "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	return rr.websocketRoot.merge(segments, handler)
}

func NewRouter() Router {
	httpRoot := httpRoute{childRoutes: map[string]*httpRoute{}, handler: nil}
	websocketRoot := websocketRoute{childRoutes: map[string]*websocketRoute{}, handler: nil}
	return &recursiveRouter{httpRoot: httpRoot, websocketRoot: websocketRoot}
}

type httpRoute struct {
	childRoutes map[string]*httpRoute
	handler     HttpHandler
}

type websocketRoute struct {
	childRoutes map[string]*websocketRoute
	handler     WebSocketHandler
}

func (r *httpRoute) merge(segments []string, handler HttpHandler) error {
	if len(segments) == 0 {
		if r.handler != nil {
			return fmt.Errorf("conflicting path!")
		}
		r.handler = handler
		return nil
	}

	segment, remainingSegments := segments[0], segments[1:]

	childRoute, exists := r.childRoutes[segment]
	if !exists {
		childRoute = &httpRoute{
			childRoutes: map[string]*httpRoute{},
			handler:     nil,
		}
		r.childRoutes[segment] = childRoute
	}
	err := childRoute.merge(remainingSegments, handler)
	return err
}

func (fpe httpRoute) match(segments []string) (HttpHandler, int, bool) {
	if len(segments) == 0 {
		if fpe.handler == nil {
			return nil, -1, false
		}
		return fpe.handler, 1, true
	}

	segment := segments[0]
	remainingSegments := segments[1:]

	childRoute, exists := fpe.childRoutes[segment]

	if !exists {
		return nil, -1, false
	}

	return childRoute.match(remainingSegments)
}

func (wsr *websocketRoute) merge(segments []string, handler WebSocketHandler) error {
	if len(segments) == 0 {
		if wsr.handler != nil {
			return fmt.Errorf("conflicting path!")
		}
		wsr.handler = handler
		return nil
	}

	segment, remainingSegments := segments[0], segments[1:]

	childRoute, exists := wsr.childRoutes[segment]
	if !exists {
		childRoute = &websocketRoute{
			childRoutes: map[string]*websocketRoute{},
			handler:     nil,
		}
		wsr.childRoutes[segment] = childRoute
	}
	err := childRoute.merge(remainingSegments, handler)
	return err
}

func (wsr websocketRoute) match(segments []string) (WebSocketHandler, int, bool) {
	if len(segments) == 0 {
		if wsr.handler == nil {
			return nil, -1, false
		}
		return wsr.handler, 1, true
	}

	segment := segments[0]
	remainingSegments := segments[1:]

	childRoute, exists := wsr.childRoutes[segment]

	if !exists {
		return nil, -1, false
	}

	return childRoute.match(remainingSegments)
}
