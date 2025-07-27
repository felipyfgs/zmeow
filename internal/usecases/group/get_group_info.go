package group

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types"

	"zmeow/internal/domain/group"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/infra/whatsapp/services"
	"zmeow/pkg/logger"
)

// GetGroupInfoUseCase implementa o caso de uso para obter informações de um grupo
type GetGroupInfoUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewGetGroupInfoUseCase cria uma nova instância do caso de uso
func NewGetGroupInfoUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *GetGroupInfoUseCase {
	return &GetGroupInfoUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para obter informações de um grupo
func (uc *GetGroupInfoUseCase) Execute(ctx context.Context, sessionID uuid.UUID, groupJIDStr string) (*group.Group, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  groupJIDStr,
	}).Info().Msg("Getting group info")

	// Validar entrada
	if err := uc.validateRequest(groupJIDStr); err != nil {
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

	// Obter informações do grupo via GroupService
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	groupInfo, err := groupService.GetGroupInfoByString(ctx, groupJIDStr)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get group info via WhatsApp")
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  groupJIDStr,
		"groupName": groupInfo.Name,
	}).Info().Msg("Group info retrieved successfully")

	return groupInfo, nil
}

// validateRequest valida a requisição de obter informações do grupo
func (uc *GetGroupInfoUseCase) validateRequest(groupJIDStr string) error {
	if groupJIDStr == "" {
		return group.NewValidationError("groupJid", groupJIDStr, "group JID is required")
	}

	// Validar formato básico do JID
	if !strings.Contains(groupJIDStr, "@") {
		return group.NewValidationError("groupJid", groupJIDStr, "invalid JID format")
	}

	// Verificar se é um JID de grupo
	if !strings.HasSuffix(groupJIDStr, "@g.us") {
		return group.NewValidationError("groupJid", groupJIDStr, "not a group JID")
	}

	return nil
}

// parseGroupJID converte string para JID de grupo
func (uc *GetGroupInfoUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
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

// getGroupInfoViaWhatsApp obtém informações do grupo usando o cliente WhatsApp
func (uc *GetGroupInfoUseCase) getGroupInfoViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID) (*group.Group, error) {
	// TODO: Implementar método específico de obter informações do grupo no WhatsAppClient
	// Por enquanto, vamos simular a obtenção de informações

	// Em uma implementação real, usaríamos algo como client.GetGroupInfo(groupJID)

	// Retornar grupo simulado para desenvolvimento
	groupInfo := &group.Group{
		JID:              groupJID,
		Name:             "Grupo Simulado",
		Topic:            "Este é um grupo simulado para desenvolvimento",
		Participants:     []group.Participant{},
		Admins:           []types.JID{},
		Owner:            types.JID{},
		CreatedAt:        time.Now(),
		IsAnnounce:       false,
		IsLocked:         false,
		IsEphemeral:      false,
		EphemeralTimer:   0,
		PictureID:        "",
		InviteCode:       "",
		ParticipantCount: 0,
	}

	return groupInfo, nil
}
