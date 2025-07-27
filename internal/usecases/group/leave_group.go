package group

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types"

	"zmeow/internal/domain/group"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/infra/whatsapp/services"
	"zmeow/pkg/logger"
)

// LeaveGroupUseCase implementa o caso de uso para sair de um grupo
type LeaveGroupUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewLeaveGroupUseCase cria uma nova instância do caso de uso
func NewLeaveGroupUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *LeaveGroupUseCase {
	return &LeaveGroupUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para sair de um grupo
func (uc *LeaveGroupUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.LeaveGroupRequest) error {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
	}).Info().Msg("Leaving group")

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

	// Sair do grupo via GroupService
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	err = groupService.LeaveGroupByString(ctx, req.GroupJID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to leave group via GroupService")
		return fmt.Errorf("failed to leave group: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
	}).Info().Msg("Left group successfully")

	return nil
}

// validateRequest valida a requisição de sair do grupo
func (uc *LeaveGroupUseCase) validateRequest(req group.LeaveGroupRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	// Validar formato básico do JID
	if !strings.Contains(req.GroupJID, "@") {
		return group.NewValidationError("groupJid", req.GroupJID, "invalid JID format")
	}

	// Verificar se é um JID de grupo
	if !strings.HasSuffix(req.GroupJID, "@g.us") {
		return group.NewValidationError("groupJid", req.GroupJID, "not a group JID")
	}

	return nil
}

// parseGroupJID converte string para JID de grupo
func (uc *LeaveGroupUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
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

// validateUserCanLeave verifica se o usuário pode sair do grupo
func (uc *LeaveGroupUseCase) validateUserCanLeave(ctx context.Context, client whatsapp.WhatsAppClient, groupJID types.JID) error {
	// TODO: Implementar verificação se o usuário é o dono do grupo
	// Por enquanto, permitir que qualquer usuário saia

	// Em uma implementação real, seria algo como:
	// groupInfo, err := client.GetGroupInfo(ctx, sessionID, groupJID)
	// if err != nil {
	//     return fmt.Errorf("failed to get group info: %w", err)
	// }
	//
	// userJID := client.GetOwnJID()
	// if groupInfo.IsUserOwner(userJID) {
	//     return group.ErrCannotLeaveAsOwner
	// }

	uc.logger.WithField("groupJid", groupJID.String()).Info().Msg("User can leave group (validation simulated)")

	return nil
}

// leaveGroupViaWhatsApp sai do grupo usando o cliente WhatsApp
func (uc *LeaveGroupUseCase) leaveGroupViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID) error {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação

	// Em uma implementação real, seria algo como:
	// return client.LeaveGroup(ctx, sessionID, groupJID)

	uc.logger.WithField("groupJid", groupJID.String()).Info().Msg("Left group (simulated)")

	return nil
}
