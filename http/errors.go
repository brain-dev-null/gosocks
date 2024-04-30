package http

import "fmt"

type HttpError struct {
	StatusCode int
	Message    string
}

func (he HttpError) Error() string {
	return fmt.Sprintf("%d: %s", he.StatusCode, he.Message)
}

func ErrorNotFound(message string) error {
	return HttpError{
		StatusCode: 404,
		Message:    message,
	}
}
