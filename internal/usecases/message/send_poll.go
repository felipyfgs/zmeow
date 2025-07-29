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

// SendPollMessageUseCase implementa o caso de uso para envio de enquete
type SendPollMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewSendPollMessageUseCase cria uma nova instância do caso de uso
func NewSendPollMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendPollMessageUseCase {
	return &SendPollMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para enviar enquete
func (uc *SendPollMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendPollMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId":       sessionID,
		"to":              req.To,
		"name":            req.Name,
		"optionCount":     len(req.Options),
		"selectableCount": req.SelectableOptionsCount,
	}).Info().Msg("Sending poll message")

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

	// Enviar enquete
	messageID, err := client.SendPollMessage(ctx, sessionID, normalizedPhone, req.Name, req.Options, req.SelectableOptionsCount)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send poll message")
		return nil, fmt.Errorf("failed to send poll: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"to":        normalizedPhone,
		"messageId": messageID,
	}).Info().Msg("Poll message sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     messageID,
		Status: "sent",
		Details: map[string]interface{}{
			"to":              normalizedPhone,
			"sessionId":       sessionID,
			"type":            "poll",
			"optionCount":     len(req.Options),
			"selectableCount": req.SelectableOptionsCount,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de enquete
func (uc *SendPollMessageUseCase) validateRequest(req message.SendPollMessageRequest) error {
	if req.To == "" {
		return fmt.Errorf("to is required")
	}

	if req.Name == "" {
		return fmt.Errorf("poll name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("poll name is too long (max 255 characters)")
	}

	// Validar opções
	if len(req.Options) < 2 {
		return fmt.Errorf("poll must have at least 2 options")
	}

	if len(req.Options) > 12 {
		return fmt.Errorf("poll can have maximum 12 options")
	}

	// Validar selectableOptionsCount
	if req.SelectableOptionsCount < 1 {
		return fmt.Errorf("selectableOptionsCount must be at least 1")
	}

	if req.SelectableOptionsCount > len(req.Options) {
		return fmt.Errorf("selectableOptionsCount cannot be greater than the number of options (%d)", len(req.Options))
	}

	// Validar cada opção
	optionTexts := make(map[string]bool)
	for i, option := range req.Options {
		if option == "" {
			return fmt.Errorf("option %d: text is required", i+1)
		}

		if len(option) > 100 {
			return fmt.Errorf("option %d: text is too long (max 100 characters)", i+1)
		}

		// Verificar opções únicas
		if optionTexts[option] {
			return fmt.Errorf("option %d: duplicate option text '%s'", i+1, option)
		}
		optionTexts[option] = true
	}

	// Validar formato do telefone ou JID de grupo
	if !uc.isValidPhoneNumberOrGroupJID(req.To) {
		return fmt.Errorf("invalid phone number or group JID format")
	}

	return nil
}

// isValidPhoneNumberOrGroupJID valida o formato do número de telefone ou JID de grupo
func (uc *SendPollMessageUseCase) isValidPhoneNumberOrGroupJID(input string) bool {
	// Se contém @, é um JID - validar como JID
	if strings.Contains(input, "@") {
		return uc.isValidJID(input)
	}

	// Se não contém @, é um número - validar como número
	return uc.isValidPhoneNumber(input)
}

// isValidJID valida o formato de um JID (tanto individual quanto grupo)
func (uc *SendPollMessageUseCase) isValidJID(jid string) bool {
	// Verificar formato básico
	if !strings.Contains(jid, "@") {
		return false
	}

	// Aceitar tanto @s.whatsapp.net (individual) quanto @g.us (grupo)
	if strings.HasSuffix(jid, "@s.whatsapp.net") || strings.HasSuffix(jid, "@g.us") {
		// Extrair a parte antes do @
		parts := strings.Split(jid, "@")
		if len(parts) != 2 {
			return false
		}

		numberPart := parts[0]

		// Para grupos, aceitar números longos
		if strings.HasSuffix(jid, "@g.us") {
			// Grupos podem ter números mais longos
			return len(numberPart) >= 10 && len(numberPart) <= 25
		}

		// Para indivíduos, usar validação normal
		return uc.isValidPhoneNumber(numberPart)
	}

	return false
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *SendPollMessageUseCase) isValidPhoneNumber(phone string) bool {
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
func (uc *SendPollMessageUseCase) normalizePhoneNumber(phone string) string {
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
