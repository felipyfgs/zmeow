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

// EditMessageUseCase implementa o caso de uso para edição de mensagem
type EditMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewEditMessageUseCase cria uma nova instância do caso de uso
func NewEditMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *EditMessageUseCase {
	return &EditMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para editar mensagem
func (uc *EditMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.EditMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     req.Number,
		"messageId": req.ID,
		"newText":   req.NewText,
	}).Info().Msg("Editing message")

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

	// Editar mensagem
	editedMessageID, err := client.EditMessage(ctx, sessionID, normalizedPhone, req.ID, req.NewText)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to edit message")
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":       sessionID,
		"phone":           normalizedPhone,
		"originalId":      req.ID,
		"editedId":        editedMessageID,
	}).Info().Msg("Message edited successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     editedMessageID,
		Status: "edited",
		Details: map[string]interface{}{
			"phone":      normalizedPhone,
			"sessionId":  sessionID,
			"type":       "edit",
			"originalId": req.ID,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de edição de mensagem
func (uc *EditMessageUseCase) validateRequest(req message.EditMessageRequest) error {
	if req.Number == "" {
		return fmt.Errorf("phone number is required")
	}

	if req.ID == "" {
		return fmt.Errorf("message ID is required")
	}

	if req.NewText == "" {
		return fmt.Errorf("new text is required")
	}

	if len(req.NewText) > 4096 {
		return fmt.Errorf("new text is too long (max 4096 characters)")
	}

	// Validar formato do message ID (deve ser um ID válido)
	if len(req.ID) < 10 {
		return fmt.Errorf("invalid message ID format")
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.Number) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *EditMessageUseCase) isValidPhoneNumber(phone string) bool {
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
func (uc *EditMessageUseCase) normalizePhoneNumber(phone string) string {
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
