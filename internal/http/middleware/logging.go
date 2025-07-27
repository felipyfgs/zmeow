package middleware

import (
	"net/http"
	"time"

	"zmeow/pkg/logger"

	"github.com/go-chi/chi/v5/middleware"
)

// NewLoggingMiddleware cria um middleware de logging usando o logger centralizado
func NewLoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Usar o wrapper de response writer do chi para capturar status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				duration := time.Since(start)
				status := ww.Status()

				// Só logar erros, requests muito lentos ou em debug mode
				if status >= 400 {
					// Erros sempre são logados
					log.WithFields(map[string]interface{}{
						"method": r.Method,
						"path":   r.URL.Path,
						"status": status,
						"ms":     duration.Milliseconds(),
					}).Error().Msg("HTTP error")
				} else if duration > 3*time.Second {
					// Requests muito lentos como warning (3s+ para operações como stickers)
					log.WithFields(map[string]interface{}{
						"method": r.Method,
						"path":   r.URL.Path,
						"status": status,
						"ms":     duration.Milliseconds(),
					}).Warn().Msg("Slow request")
				} else {
					// Requests normais só em debug
					log.WithFields(map[string]interface{}{
						"method": r.Method,
						"path":   r.URL.Path,
						"status": status,
						"ms":     duration.Milliseconds(),
					}).Debug().Msg("HTTP")
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
