package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"zmeow/internal/http/responses"
)

// NewRateLimit cria um middleware de rate limiting
func NewRateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	return httprate.Limit(
		requestsPerMinute,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			responses.TooManyRequests(w, "Rate limit excedido")
		}),
	)
}