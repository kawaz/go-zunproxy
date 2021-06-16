package middleware

import (
	"net/http"
)

var NoopHandler = noopHandler{}

type noopHandler struct{}

func (noop noopHandler) Handle(h http.Handler) http.Handler {
	return h
}
