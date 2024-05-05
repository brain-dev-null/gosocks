package http

import (
	"fmt"
	"strings"
)

type Handler func(HttpRequest) (HttpResponse, error)

type Router interface {
	Route(request HttpRequest) (Handler, error)
	AddRoute(path string, handler Handler) error
}

type recursiveRouter struct {
	root route
}

func (rr *recursiveRouter) Route(request HttpRequest) (Handler, error) {
	segments := strings.Split(request.Path(), "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	handler, _, matched := rr.root.match(segments)

	if !matched {
		return nil, ErrorNotFound(fmt.Sprintf("No route for: %s", request.Path()))
	}

	return handler, nil
}

func (rr *recursiveRouter) AddRoute(path string, handler Handler) error {
	segments := strings.Split(path, "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	return rr.root.merge(segments, handler)
}

func NewRouter() Router {
	root := route{childRoutes: map[string]*route{}, handler: nil}
	return &recursiveRouter{root: root}
}

type route struct {
	childRoutes map[string]*route
	handler     Handler
}

func (r *route) merge(segments []string, handler Handler) error {
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
		childRoute = &route{
			childRoutes: map[string]*route{},
			handler:     nil,
		}
		r.childRoutes[segment] = childRoute
	}
	err := childRoute.merge(remainingSegments, handler)
	return err
}

func (fpe route) match(segments []string) (Handler, int, bool) {
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
