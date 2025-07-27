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

// SendListMessageUseCase implementa o caso de uso para envio de mensagem com lista
type SendListMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewSendListMessageUseCase cria uma nova instância do caso de uso
func NewSendListMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendListMessageUseCase {
	return &SendListMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para enviar mensagem com lista
func (uc *SendListMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendListMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId":    sessionID,
		"phone":        req.Number,
		"text":         req.Text,
		"footer":       req.Footer,
		"title":        req.Title,
		"buttonText":   req.ButtonText,
		"sectionCount": len(req.Sections),
	}).Info().Msg("Sending list message")

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

	// Enviar mensagem com lista
	messageID, err := client.SendListMessage(ctx, sessionID, normalizedPhone, req.Text, req.Footer, req.Title, req.ButtonText, req.Sections)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send list message")
		return nil, fmt.Errorf("failed to send list message: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     normalizedPhone,
		"messageId": messageID,
	}).Info().Msg("List message sent successfully")

	// Calcular total de itens
	totalRows := 0
	for _, section := range req.Sections {
		totalRows += len(section.Rows)
	}

	// Criar resposta
	response := &message.SendMessageResponse{
		ID: messageID,
		Status:    "sent",
		Details: map[string]interface{}{
			"phone":        normalizedPhone,
			"sessionId":    sessionID,
			"type":         "list",
			"sectionCount": len(req.Sections),
			"totalRows":    totalRows,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de mensagem com lista
func (uc *SendListMessageUseCase) validateRequest(req message.SendListMessageRequest) error {
	if req.Number == "" {
		return fmt.Errorf("phone number is required")
	}

	if req.Text == "" {
		return fmt.Errorf("text is required")
	}

	if req.Title == "" {
		return fmt.Errorf("title is required")
	}

	if req.ButtonText == "" {
		return fmt.Errorf("button text is required")
	}

	if len(req.Text) > 4096 {
		return fmt.Errorf("text is too long (max 4096 characters)")
	}

	if len(req.Footer) > 60 {
		return fmt.Errorf("footer is too long (max 60 characters)")
	}

	if len(req.Title) > 60 {
		return fmt.Errorf("title is too long (max 60 characters)")
	}

	if len(req.ButtonText) > 20 {
		return fmt.Errorf("button text is too long (max 20 characters)")
	}

	// Validar seções
	if len(req.Sections) == 0 {
		return fmt.Errorf("at least one section is required")
	}

	if len(req.Sections) > 10 {
		return fmt.Errorf("maximum 10 sections allowed")
	}

	// Validar cada seção
	rowIDs := make(map[string]bool)
	for i, section := range req.Sections {
		if section.Title == "" {
			return fmt.Errorf("section %d: title is required", i+1)
		}

		if len(section.Title) > 24 {
			return fmt.Errorf("section %d: title is too long (max 24 characters)", i+1)
		}

		if len(section.Rows) == 0 {
			return fmt.Errorf("section %d: at least one row is required", i+1)
		}

		if len(section.Rows) > 10 {
			return fmt.Errorf("section %d: maximum 10 rows per section allowed", i+1)
		}

		// Validar cada linha da seção
		for j, row := range section.Rows {
			if row.ID == "" {
				return fmt.Errorf("section %d, row %d: ID is required", i+1, j+1)
			}

			if row.Title == "" {
				return fmt.Errorf("section %d, row %d: title is required", i+1, j+1)
			}

			if len(row.ID) > 200 {
				return fmt.Errorf("section %d, row %d: ID is too long (max 200 characters)", i+1, j+1)
			}

			if len(row.Title) > 24 {
				return fmt.Errorf("section %d, row %d: title is too long (max 24 characters)", i+1, j+1)
			}

			if len(row.Description) > 72 {
				return fmt.Errorf("section %d, row %d: description is too long (max 72 characters)", i+1, j+1)
			}

			// Verificar IDs únicos globalmente
			if rowIDs[row.ID] {
				return fmt.Errorf("section %d, row %d: duplicate ID '%s'", i+1, j+1, row.ID)
			}
			rowIDs[row.ID] = true
		}
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.Number) {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *SendListMessageUseCase) isValidPhoneNumber(phone string) bool {
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
func (uc *SendListMessageUseCase) normalizePhoneNumber(phone string) string {
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
