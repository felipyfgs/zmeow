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

// UpdateParticipantsUseCase implementa o caso de uso para gerenciar participantes
type UpdateParticipantsUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewUpdateParticipantsUseCase cria uma nova instância do caso de uso
func NewUpdateParticipantsUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *UpdateParticipantsUseCase {
	return &UpdateParticipantsUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para atualizar participantes
func (uc *UpdateParticipantsUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.UpdateParticipantsRequest) error {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId":        sessionID,
		"groupJid":         req.GroupJID,
		"action":           req.Action,
		"participantCount": len(req.Phones),
	}).Info().Msg("Updating group participants")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return err
	}

	// Verificar se a sessão existe
	_, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return fmt.Errorf("session not found: %w", err)
	}

	// Verificar se a sessão está conectada
	if !uc.whatsappManager.IsConnected(sessionID) {
		uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session is not connected")
		return fmt.Errorf("session %s is not connected", sessionID)
	}

	// Atualizar participantes via GroupService
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	err = groupService.UpdateParticipants(ctx, req.GroupJID, req.Phones, req.Action)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to update participants via GroupService")
		return fmt.Errorf("failed to %s participants: %w", req.Action, err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"action":    req.Action,
	}).Info().Msg("Participants updated successfully")

	return nil
}

// validateRequest valida a requisição de atualização de participantes
func (uc *UpdateParticipantsUseCase) validateRequest(req group.UpdateParticipantsRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	if req.Action == "" {
		return group.NewValidationError("action", req.Action, "action is required")
	}

	// Validar ação
	action := group.ParticipantAction(req.Action)
	switch action {
	case group.ParticipantActionAdd, group.ParticipantActionRemove,
		group.ParticipantActionPromote, group.ParticipantActionDemote:
		// Ação válida
	default:
		return group.NewValidationError("action", req.Action, "invalid action")
	}

	if len(req.Phones) == 0 {
		return group.NewValidationError("phones", "", "at least one phone is required")
	}

	if len(req.Phones) > 50 {
		return group.NewValidationError("phones", "", "maximum 50 participants per operation")
	}

	// Validar cada número de telefone
	for i, phone := range req.Phones {
		if !uc.isValidPhoneNumber(phone) {
			return group.NewValidationError(
				fmt.Sprintf("phones[%d]", i),
				phone,
				"invalid phone number format",
			)
		}
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *UpdateParticipantsUseCase) isValidPhoneNumber(phone string) bool {
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
func (uc *UpdateParticipantsUseCase) normalizePhoneNumber(phone string) string {
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
func (uc *UpdateParticipantsUseCase) convertPhonesToJIDs(phones []string) ([]types.JID, error) {
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

// parseGroupJID converte string para JID de grupo
func (uc *UpdateParticipantsUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
	// Se não contém @, assumir que é apenas o ID e adicionar o servidor de grupo
	if !strings.Contains(groupJIDStr, "@") {
		groupJIDStr = groupJIDStr + "@g.us"
	}

	// Parse do JID
	jid, err := types.ParseJID(groupJIDStr)
	if err != nil {
		return types.JID{}, fmt.Errorf("failed to parse JID: %w", err)
	}

	// Verificar se é um JID de grupo válido
	if jid.Server != types.GroupServer {
		return types.JID{}, fmt.Errorf("not a group JID: %s", jid.Server)
	}

	if jid.User == "" {
		return types.JID{}, fmt.Errorf("empty group ID")
	}

	return jid, nil
}

// addParticipants adiciona participantes ao grupo
func (uc *UpdateParticipantsUseCase) addParticipants(client whatsapp.WhatsAppClient, groupJID types.JID, participants []types.JID) error {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação
	uc.logger.WithFields(map[string]interface{}{
		"groupJid":         groupJID.String(),
		"participantCount": len(participants),
	}).Info().Msg("Adding participants (simulated)")

	return nil
}

// removeParticipants remove participantes do grupo
func (uc *UpdateParticipantsUseCase) removeParticipants(client whatsapp.WhatsAppClient, groupJID types.JID, participants []types.JID) error {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação
	uc.logger.WithFields(map[string]interface{}{
		"groupJid":         groupJID.String(),
		"participantCount": len(participants),
	}).Info().Msg("Removing participants (simulated)")

	return nil
}

// promoteParticipants promove participantes a admin
func (uc *UpdateParticipantsUseCase) promoteParticipants(client whatsapp.WhatsAppClient, groupJID types.JID, participants []types.JID) error {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação
	uc.logger.WithFields(map[string]interface{}{
		"groupJid":         groupJID.String(),
		"participantCount": len(participants),
	}).Info().Msg("Promoting participants (simulated)")

	return nil
}

// demoteParticipants rebaixa participantes de admin
func (uc *UpdateParticipantsUseCase) demoteParticipants(client whatsapp.WhatsAppClient, groupJID types.JID, participants []types.JID) error {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação
	uc.logger.WithFields(map[string]interface{}{
		"groupJid":         groupJID.String(),
		"participantCount": len(participants),
	}).Info().Msg("Demoting participants (simulated)")

	return nil
}
