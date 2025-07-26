package handlers

import (
	"net/http"

	"zmeow/internal/http/responses"
)

// HealthHandler implementa o handler para health check
type HealthHandler struct{}

// NewHealthHandler cria uma nova instância do health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health verifica a saúde da aplicação
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	responses.Success200(w, "Service is healthy", map[string]interface{}{
		"status": "ok",
		"service": "zmeow-api",
	})
}