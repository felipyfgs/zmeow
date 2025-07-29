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

// SendContactMessageUseCase implementa o caso de uso para envio de contato
type SendContactMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewSendContactMessageUseCase cria uma nova instância do caso de uso
func NewSendContactMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendContactMessageUseCase {
	return &SendContactMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para enviar contato
func (uc *SendContactMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendContactMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"to":          req.To,
		"contactName": req.ContactName,
		"contactJID":  req.ContactJID,
	}).Info().Msg("Sending contact message")

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
	normalizedPhone := uc.normalizePhoneNumber(req.To)
	normalizedContactJID := uc.normalizePhoneNumber(req.ContactJID)

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client")
		return nil, fmt.Errorf("failed to get WhatsApp client: %w", err)
	}

	// Enviar contato
	messageID, err := client.SendContactMessage(ctx, sessionID, normalizedPhone, req.ContactName, normalizedContactJID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send contact message")
		return nil, fmt.Errorf("failed to send contact: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"to":        normalizedPhone,
		"messageId": messageID,
	}).Info().Msg("Contact message sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     messageID,
		Status: "sent",
		Details: map[string]interface{}{
			"to":          normalizedPhone,
			"sessionId":   sessionID,
			"type":        "contact",
			"contactName": req.ContactName,
			"contactJID":  normalizedContactJID,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de contato
func (uc *SendContactMessageUseCase) validateRequest(req message.SendContactMessageRequest) error {
	if req.To == "" {
		return fmt.Errorf("to is required")
	}

	if req.ContactName == "" {
		return fmt.Errorf("contact name is required")
	}

	if req.ContactJID == "" {
		return fmt.Errorf("contact JID is required")
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.To) {
		return fmt.Errorf("invalid phone number format")
	}

	// Validar formato do JID do contato
	if !uc.isValidPhoneNumber(req.ContactJID) {
		return fmt.Errorf("invalid contact JID format")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *SendContactMessageUseCase) isValidPhoneNumber(phone string) bool {
	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Verificar se tem pelo menos 10 dígitos e no máximo 15
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}

	return true
}

// normalizePhoneNumber normaliza o número de telefone para o formato WhatsApp
func (uc *SendContactMessageUseCase) normalizePhoneNumber(phone string) string {
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
