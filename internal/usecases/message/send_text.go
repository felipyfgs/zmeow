package message

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"zmeow/internal/domain/message"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// SendTextMessageUseCase implementa o caso de uso para envio de mensagem de texto
type SendTextMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
	numberValidator *NumberValidator
}

// NewSendTextMessageUseCase cria uma nova instância do caso de uso
func NewSendTextMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendTextMessageUseCase {
	return &SendTextMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
		numberValidator: NewNumberValidator(),
	}
}

// Execute executa o caso de uso para enviar mensagem de texto
func (uc *SendTextMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendTextMessageRequest) (*message.SendMessageResponse, error) {
	// Obter destinatário
	destination := uc.numberValidator.GetDestination(req.Number, req.GroupJid)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"number":      req.Number,
		"groupJid":    req.GroupJid,
		"destination": destination,
		"text":        req.Text,
	}).Info().Msg("Sending text message")

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

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client")
		return nil, fmt.Errorf("failed to get WhatsApp client: %w", err)
	}

	// Enviar mensagem
	messageID, err := client.SendTextMessage(ctx, sessionID, destination, req.Text)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send text message")
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"destination": destination,
		"messageId":   messageID,
	}).Info().Msg("Text message sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     messageID,
		Status: "sent",
		Details: map[string]interface{}{
			"number":      req.Number,
			"groupJid":    req.GroupJid,
			"destination": destination,
			"sessionId":   sessionID,
			"type":        "text",
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de mensagem
func (uc *SendTextMessageUseCase) validateRequest(req message.SendTextMessageRequest) error {
	// Validar destinatário (number ou groupJid)
	if err := uc.numberValidator.ValidateDestination(req.Number, req.GroupJid); err != nil {
		return err
	}

	if req.Text == "" {
		return fmt.Errorf("text is required")
	}

	if len(req.Text) > 4096 {
		return fmt.Errorf("text is too long (max 4096 characters)")
	}

	return nil
}
