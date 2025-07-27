package message

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"zmeow/internal/domain/message"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// SendLocationMessageUseCase implementa o caso de uso para envio de localização
type SendLocationMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewSendLocationMessageUseCase cria uma nova instância do caso de uso
func NewSendLocationMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendLocationMessageUseCase {
	return &SendLocationMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para enviar localização
func (uc *SendLocationMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendLocationMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     req.Number,
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"name":      req.Name,
		"address":   req.Address,
	}).Info().Msg("Sending location message")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return nil, err
	}

	// Verificar se a sessão existe
	_, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Verificar se a sessão está conectada usando o WhatsApp Manager
	if !uc.whatsappManager.IsConnected(sessionID) {
		uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session is not connected")
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	// Normalizar número de telefone
	normalizedPhone := uc.normalizePhoneNumber(req.Number)

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client")
		return nil, fmt.Errorf("failed to get WhatsApp client: %w", err)
	}

	// Enviar localização
	messageID, err := client.SendLocationMessage(ctx, sessionID, normalizedPhone, req.Latitude, req.Longitude, req.Name, req.Address)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send location message")
		return nil, fmt.Errorf("failed to send location: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     normalizedPhone,
		"messageId": messageID,
	}).Info().Msg("Location message sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID: messageID,
		Status:    "sent",
		Details: map[string]interface{}{
			"phone":     normalizedPhone,
			"sessionId": sessionID,
			"type":      "location",
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"name":      req.Name,
			"address":   req.Address,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de localização
func (uc *SendLocationMessageUseCase) validateRequest(req message.SendLocationMessageRequest) error {
	if req.Number == "" {
		return fmt.Errorf("phone number is required")
	}

	if req.Latitude < -90 || req.Latitude > 90 {
		return fmt.Errorf("invalid latitude: must be between -90 and 90")
	}

	if req.Longitude < -180 || req.Longitude > 180 {
		return fmt.Errorf("invalid longitude: must be between -180 and 180")
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.Number) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *SendLocationMessageUseCase) isValidPhoneNumber(phone string) bool {
	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Verificar se tem pelo menos 10 dígitos e no máximo 15
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}

	return true
}

// normalizePhoneNumber normaliza o número de telefone para o formato WhatsApp
func (uc *SendLocationMessageUseCase) normalizePhoneNumber(phone string) string {
	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Se não tem código de país, assumir Brasil (55)
	if len(cleaned) == 10 || len(cleaned) == 11 {
		if !strings.HasPrefix(cleaned, "55") {
			cleaned = "55" + cleaned
		}
	}

	// Adicionar sufixo @s.whatsapp.net se não estiver presente
	if !strings.Contains(phone, "@") {
		return cleaned + "@s.whatsapp.net"
	}

	return phone
}
