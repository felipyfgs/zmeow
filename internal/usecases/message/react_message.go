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

// ReactMessageUseCase implementa o caso de uso para reagir a mensagem
type ReactMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewReactMessageUseCase cria uma nova instância do caso de uso
func NewReactMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *ReactMessageUseCase {
	return &ReactMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para reagir a mensagem
func (uc *ReactMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.ReactMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"number":    req.Number,
		"messageId": req.ID,
		"reaction":  req.Reaction,
	}).Info().Msg("Reacting to message")

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
	normalizedPhone := uc.normalizePhoneNumber(req.Number)

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client")
		return nil, fmt.Errorf("failed to get WhatsApp client: %w", err)
	}

	// Reagir à mensagem
	err = client.ReactMessage(ctx, sessionID, normalizedPhone, req.ID, req.Reaction)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to react to message")
		return nil, fmt.Errorf("failed to react to message: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"number":    normalizedPhone,
		"messageId": req.ID,
		"reaction":  req.Reaction,
	}).Info().Msg("Reaction sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     req.ID,
		Status: "reacted",
		Details: map[string]interface{}{
			"number":    normalizedPhone,
			"sessionId": sessionID,
			"type":      "reaction",
			"reaction":  req.Reaction,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de reagir a mensagem
func (uc *ReactMessageUseCase) validateRequest(req message.ReactMessageRequest) error {
	if req.Number == "" {
		return fmt.Errorf("phone number is required")
	}

	if req.ID == "" {
		return fmt.Errorf("messageID is required")
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.Number) {
		return fmt.Errorf("invalid phone number format")
	}

	// Validar reação (básico) - string vazia é permitida para remoção
	if len(req.Reaction) > 10 {
		return fmt.Errorf("reaction too long")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *ReactMessageUseCase) isValidPhoneNumber(phone string) bool {
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
func (uc *ReactMessageUseCase) normalizePhoneNumber(phone string) string {
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
