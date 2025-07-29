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

// SendButtonsMessageUseCase implementa o caso de uso para envio de mensagem com botões
type SendButtonsMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewSendButtonsMessageUseCase cria uma nova instância do caso de uso
func NewSendButtonsMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendButtonsMessageUseCase {
	return &SendButtonsMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para enviar mensagem com botões
func (uc *SendButtonsMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendButtonsMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"to":          req.To,
		"text":        req.Text,
		"footer":      req.Footer,
		"buttonCount": len(req.Buttons),
	}).Info().Msg("Sending buttons message")

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

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client")
		return nil, fmt.Errorf("failed to get WhatsApp client: %w", err)
	}

	// Enviar mensagem com botões
	messageID, err := client.SendButtonsMessage(ctx, sessionID, normalizedPhone, req.Text, req.Footer, req.Buttons)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send buttons message")
		return nil, fmt.Errorf("failed to send buttons message: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"to":        normalizedPhone,
		"messageId": messageID,
	}).Info().Msg("Buttons message sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     messageID,
		Status: "sent",
		Details: map[string]interface{}{
			"to":          normalizedPhone,
			"sessionId":   sessionID,
			"type":        "buttons",
			"buttonCount": len(req.Buttons),
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de mensagem com botões
func (uc *SendButtonsMessageUseCase) validateRequest(req message.SendButtonsMessageRequest) error {
	if req.To == "" {
		return fmt.Errorf("to is required")
	}

	if req.Text == "" {
		return fmt.Errorf("text is required")
	}

	if len(req.Text) > 4096 {
		return fmt.Errorf("text is too long (max 4096 characters)")
	}

	if len(req.Footer) > 60 {
		return fmt.Errorf("footer is too long (max 60 characters)")
	}

	// Validar botões
	if len(req.Buttons) == 0 {
		return fmt.Errorf("at least one button is required")
	}

	if len(req.Buttons) > 3 {
		return fmt.Errorf("maximum 3 buttons allowed")
	}

	// Validar cada botão
	buttonIDs := make(map[string]bool)
	for i, button := range req.Buttons {
		if button.ID == "" {
			return fmt.Errorf("button %d: ID is required", i+1)
		}

		if button.DisplayText == "" {
			return fmt.Errorf("button %d: display text is required", i+1)
		}

		if len(button.ID) > 256 {
			return fmt.Errorf("button %d: ID is too long (max 256 characters)", i+1)
		}

		if len(button.DisplayText) > 20 {
			return fmt.Errorf("button %d: display text is too long (max 20 characters)", i+1)
		}

		// Verificar IDs únicos
		if buttonIDs[button.ID] {
			return fmt.Errorf("button %d: duplicate ID '%s'", i+1, button.ID)
		}
		buttonIDs[button.ID] = true

		// Validar tipo do botão se fornecido
		if button.Type != "" && button.Type != "RESPONSE" {
			return fmt.Errorf("button %d: invalid type '%s' (only 'RESPONSE' is supported)", i+1, button.Type)
		}
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.To) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *SendButtonsMessageUseCase) isValidPhoneNumber(phone string) bool {
	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Verificar se tem pelo menos 10 dígitos e no máximo 15
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}

	// Verificar se começa com código de país válido
	if len(cleaned) >= 11 && (strings.HasPrefix(cleaned, "55") || strings.HasPrefix(cleaned, "1")) {
		return true
	}

	// Aceitar números com 10-11 dígitos (formato nacional)
	if len(cleaned) >= 10 && len(cleaned) <= 11 {
		return true
	}

	return false
}

// normalizePhoneNumber normaliza o número de telefone para o formato WhatsApp
func (uc *SendButtonsMessageUseCase) normalizePhoneNumber(phone string) string {
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
