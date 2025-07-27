package group

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types"

	"zmeow/internal/domain/group"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/infra/whatsapp/services"
	"zmeow/pkg/logger"
)

// CreateGroupUseCase implementa o caso de uso para criar um novo grupo
type CreateGroupUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewCreateGroupUseCase cria uma nova instância do caso de uso
func NewCreateGroupUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para criar um grupo
func (uc *CreateGroupUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.CreateGroupRequest) (*group.Group, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId":        sessionID,
		"groupName":        req.Name,
		"participantCount": len(req.Participants),
	}).Info().Msg("Creating new group")

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

	// Converter telefones para JIDs
	participantJIDs, err := uc.convertPhonesToJIDs(req.Participants)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to convert phones to JIDs")
		return nil, fmt.Errorf("failed to convert participants: %w", err)
	}

	// Criar grupo via WhatsApp
	groupInfo, err := uc.createGroupViaWhatsApp(sessionID, req.Name, participantJIDs)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to create group via WhatsApp")
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  groupInfo.JID.String(),
		"groupName": groupInfo.Name,
	}).Info().Msg("Group created successfully")

	return groupInfo, nil
}

// validateRequest valida a requisição de criação de grupo
func (uc *CreateGroupUseCase) validateRequest(req group.CreateGroupRequest) error {
	if req.Name == "" {
		return group.NewValidationError("name", req.Name, "group name is required")
	}

	if len(req.Name) > 25 {
		return group.NewValidationError("name", req.Name, "group name must be at most 25 characters")
	}

	if len(req.Participants) == 0 {
		return group.NewValidationError("participants", "", "at least one participant is required")
	}

	if len(req.Participants) > 256 {
		return group.NewValidationError("participants", "", "maximum 256 participants allowed")
	}

	// Validar cada número de telefone
	for i, phone := range req.Participants {
		if !uc.isValidPhoneNumber(phone) {
			return group.NewValidationError(
				fmt.Sprintf("participants[%d]", i),
				phone,
				"invalid phone number format",
			)
		}
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *CreateGroupUseCase) isValidPhoneNumber(phone string) bool {
	// Remover caracteres não numéricos e símbolos comuns
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
func (uc *CreateGroupUseCase) normalizePhoneNumber(phone string) string {
	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Se não tem código de país, assumir Brasil (55)
	if len(cleaned) == 10 || len(cleaned) == 11 {
		if !strings.HasPrefix(cleaned, "55") {
			cleaned = "55" + cleaned
		}
	}

	return cleaned
}

// convertPhonesToJIDs converte uma lista de telefones para JIDs
func (uc *CreateGroupUseCase) convertPhonesToJIDs(phones []string) ([]types.JID, error) {
	jids := make([]types.JID, len(phones))

	for i, phone := range phones {
		normalizedPhone := uc.normalizePhoneNumber(phone)

		// Criar JID para usuário individual
		jid := types.NewJID(normalizedPhone, types.DefaultUserServer)
		if jid.IsEmpty() {
			return nil, fmt.Errorf("failed to create JID for phone %s", phone)
		}

		jids[i] = jid
	}

	return jids, nil
}

// createGroupViaWhatsApp cria o grupo usando o cliente WhatsApp
func (uc *CreateGroupUseCase) createGroupViaWhatsApp(sessionID uuid.UUID, name string, participants []types.JID) (*group.Group, error) {
	// Criar serviço de grupo para a sessão
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)

	// Criar grupo usando o serviço
	groupInfo, err := groupService.CreateGroup(context.Background(), name, participants)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to create group via GroupService")
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return groupInfo, nil
}
