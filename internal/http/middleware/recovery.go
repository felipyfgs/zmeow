package middleware

import (
	"net/http"
	"runtime/debug"

	"zmeow/internal/http/responses"
	"zmeow/pkg/logger"
)

// NewRecoveryMiddleware cria um middleware de recovery para panic
func NewRecoveryMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.WithFields(map[string]interface{}{
						"panic":       err,
						"stack":       string(debug.Stack()),
						"method":      r.Method,
						"url":         r.URL.String(),
						"user_agent":  r.UserAgent(),
						"remote_addr": r.RemoteAddr,
					}).Error().Msg("Panic recovered")

					// Retornar erro 500 para o cliente
					responses.Error500(w, "Erro interno do servidor", "INTERNAL_ERROR", "Um erro inesperado ocorreu")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}