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

// DeleteMessageUseCase implementa o caso de uso para deletar mensagem
type DeleteMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewDeleteMessageUseCase cria uma nova instância do caso de uso
func NewDeleteMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *DeleteMessageUseCase {
	return &DeleteMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para deletar mensagem
func (uc *DeleteMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.DeleteMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"to":        req.To,
		"messageId": req.ID,
		"forMe":     req.ForMe,
	}).Info().Msg("Deleting message")

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

	// Verificar se a sessão está conectada
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

	// Deletar mensagem
	err = client.DeleteMessage(ctx, sessionID, normalizedPhone, req.ID, req.ForMe)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to delete message")
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"number":    normalizedPhone,
		"messageId": req.ID,
		"forMe":     req.ForMe,
	}).Info().Msg("Message deleted successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     req.ID,
		Status: "deleted",
		Details: map[string]interface{}{
			"to":        normalizedPhone,
			"sessionId": sessionID,
			"type":      "delete",
			"forMe":     req.ForMe,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de deletar mensagem
func (uc *DeleteMessageUseCase) validateRequest(req message.DeleteMessageRequest) error {
	if req.To == "" {
		return fmt.Errorf("to is required")
	}

	if req.ID == "" {
		return fmt.Errorf("messageID is required")
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.To) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *DeleteMessageUseCase) isValidPhoneNumber(phone string) bool {
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
func (uc *DeleteMessageUseCase) normalizePhoneNumber(phone string) string {
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
