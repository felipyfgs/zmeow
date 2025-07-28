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
// @Summary      Health Check
// @Description  Verifica se a API está funcionando corretamente
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  responses.SuccessResponse  "Serviço funcionando"
// @Router       /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	responses.Success200(w, "Service is healthy", map[string]interface{}{
		"status":  "ok",
		"service": "zmeow-api",
	})
}

// HealthData representa os dados de resposta do health check
type HealthData struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"zmeow-api"`
}
