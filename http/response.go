package http

type HttpResponse struct {
	StatusCode int
	Headers    map[string]string
	Content    []byte
}
