package session

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// SetProxyUseCase implementa o caso de uso para configurar proxy em uma sessão
type SetProxyUseCase struct {
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
	validator       *validator.Validate
}

// NewSetProxyUseCase cria uma nova instância do caso de uso
func NewSetProxyUseCase(
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SetProxyUseCase {
	return &SetProxyUseCase{
		whatsappManager: whatsappManager,
		logger:          logger.WithComponent("set-proxy-usecase"),
		validator:       validator.New(),
	}
}

// SetProxyRequest representa os dados para configurar proxy
type SetProxyRequest struct {
	ProxyURL string `json:"proxyUrl" validate:"required,url"`
}

// SetProxyResponse representa a resposta da configuração de proxy
type SetProxyResponse struct {
	ProxyURL string `json:"proxyUrl"`
}

// Execute executa o caso de uso para configurar proxy
func (uc *SetProxyUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req SetProxyRequest) (*SetProxyResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"proxyUrl":  req.ProxyURL,
	}).Info().Msg("Setting proxy configuration")

	// Validar entrada
	if err := uc.validator.Struct(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid proxy configuration request")
		return nil, err
	}

	// Configurar proxy no manager do WhatsApp
	if err := uc.whatsappManager.SetProxy(sessionID, req.ProxyURL); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to set proxy configuration")
		return nil, err
	}

	response := &SetProxyResponse{
		ProxyURL: req.ProxyURL,
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"proxyUrl":  req.ProxyURL,
	}).Info().Msg("Proxy configuration set successfully")

	return response, nil
}