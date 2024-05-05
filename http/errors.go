package http

import (
	"fmt"
)

type HttpError struct {
	StatusCode int
	Message    string
}

func (he HttpError) Error() string {
	return fmt.Sprintf("%d: %s", he.StatusCode, he.Message)
}

func (he HttpError) ToResponse() HttpResponse {
	return HttpResponse{
		StatusCode: he.StatusCode,
		Headers:    map[string]string{},
		Content:    []byte{}}
}

func BadRequest(message string) HttpError {
	return HttpError{
		StatusCode: 400,
		Message:    message,
	}
}

func ErrorNotFound(message string) HttpError {
	return HttpError{
		StatusCode: 404,
		Message:    message,
	}
}

func InternalServerError(message string) HttpError {
	return HttpError{
		StatusCode: 500,
		Message:    message,
	}
}
