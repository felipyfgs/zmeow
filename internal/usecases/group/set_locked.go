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

// SetGroupLockedUseCase implementa o caso de uso para configurar modo bloqueado do grupo
type SetGroupLockedUseCase struct {
	sessionRepo         session.SessionRepository
	whatsappManager     whatsapp.WhatsAppManager
	permissionValidator *group.PermissionValidator
	logger              logger.Logger
}

// NewSetGroupLockedUseCase cria uma nova instância do caso de uso
func NewSetGroupLockedUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	permissionValidator *group.PermissionValidator,
	logger logger.Logger,
) *SetGroupLockedUseCase {
	return &SetGroupLockedUseCase{
		sessionRepo:         sessionRepo,
		whatsappManager:     whatsappManager,
		permissionValidator: permissionValidator,
		logger:              logger,
	}
}

// Execute executa o caso de uso para configurar modo bloqueado do grupo
func (uc *SetGroupLockedUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.SetGroupLockedRequest) error {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"locked":    req.Locked,
	}).Info().Msg("Setting group locked mode")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return err
	}

	// Configurar modo bloqueado via GroupService (que já inclui validações de permissão)
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	if err := groupService.SetGroupLocked(ctx, req.GroupJID, req.Locked); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to set group locked mode via GroupService")
		return fmt.Errorf("failed to set group locked mode: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"locked":    req.Locked,
	}).Info().Msg("Group locked mode set successfully")

	return nil
}

// validateRequest valida a requisição
func (uc *SetGroupLockedUseCase) validateRequest(req group.SetGroupLockedRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	if !strings.Contains(req.GroupJID, "@") || !strings.HasSuffix(req.GroupJID, "@g.us") {
		return group.NewValidationError("groupJid", req.GroupJID, "invalid group JID format")
	}

	return nil
}

// validateSession valida a sessão
func (uc *SetGroupLockedUseCase) validateSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return fmt.Errorf("session not found: %w", err)
	}

	if !uc.whatsappManager.IsConnected(sessionID) {
		uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session is not connected")
		return fmt.Errorf("session %s is not connected", sessionID)
	}

	return nil
}

// parseGroupJID converte string para JID de grupo
func (uc *SetGroupLockedUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
	if !strings.Contains(groupJIDStr, "@") {
		groupJIDStr = groupJIDStr + "@g.us"
	}

	jid, err := types.ParseJID(groupJIDStr)
	if err != nil {
		return types.JID{}, fmt.Errorf("failed to parse JID: %w", err)
	}

	if jid.Server != types.GroupServer || jid.User == "" {
		return types.JID{}, fmt.Errorf("invalid group JID")
	}

	return jid, nil
}

// validatePermissions valida permissões do usuário (apenas dono pode alterar)
func (uc *SetGroupLockedUseCase) validatePermissions(ctx context.Context, client whatsapp.WhatsAppClient, groupJID types.JID) error {
	// Simular obtenção de informações do grupo
	groupInfo := &group.Group{
		JID:              groupJID,
		Name:             "Grupo Simulado",
		Topic:            "",
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

	userJID := types.NewJID("5511999999999", types.DefaultUserServer)
	return uc.permissionValidator.CanChangeGroupLockedMode(groupInfo, userJID)
}

// setGroupLockedViaWhatsApp configura o modo bloqueado do grupo
func (uc *SetGroupLockedUseCase) setGroupLockedViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID, locked bool) error {
	// TODO: Implementar método específico no WhatsAppClient
	uc.logger.WithFields(map[string]interface{}{
		"groupJid": groupJID.String(),
		"locked":   locked,
	}).Info().Msg("Group locked mode set (simulated)")

	return nil
}
