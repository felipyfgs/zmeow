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

// SetGroupTopicUseCase implementa o caso de uso para definir tópico do grupo
type SetGroupTopicUseCase struct {
	sessionRepo         session.SessionRepository
	whatsappManager     whatsapp.WhatsAppManager
	permissionValidator *group.PermissionValidator
	logger              logger.Logger
}

// NewSetGroupTopicUseCase cria uma nova instância do caso de uso
func NewSetGroupTopicUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	permissionValidator *group.PermissionValidator,
	logger logger.Logger,
) *SetGroupTopicUseCase {
	return &SetGroupTopicUseCase{
		sessionRepo:         sessionRepo,
		whatsappManager:     whatsappManager,
		permissionValidator: permissionValidator,
		logger:              logger,
	}
}

// Execute executa o caso de uso para definir tópico do grupo
func (uc *SetGroupTopicUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.SetGroupTopicRequest) error {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"topicLen":  len(req.Topic),
	}).Info().Msg("Setting group topic")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return err
	}

	// Definir tópico do grupo via GroupService (que já inclui validações de permissão)
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	if err := groupService.SetGroupTopic(ctx, req.GroupJID, req.Topic); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to set group topic via GroupService")
		return fmt.Errorf("failed to set group topic: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"topicLen":  len(req.Topic),
	}).Info().Msg("Group topic set successfully")

	return nil
}

// validateRequest valida a requisição
func (uc *SetGroupTopicUseCase) validateRequest(req group.SetGroupTopicRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	if len(req.Topic) > 512 {
		return group.NewValidationError("topic", req.Topic, "topic must be at most 512 characters")
	}

	if !strings.Contains(req.GroupJID, "@") || !strings.HasSuffix(req.GroupJID, "@g.us") {
		return group.NewValidationError("groupJid", req.GroupJID, "invalid group JID format")
	}

	return nil
}

// validateSession valida a sessão
func (uc *SetGroupTopicUseCase) validateSession(ctx context.Context, sessionID uuid.UUID) error {
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
func (uc *SetGroupTopicUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
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

// validatePermissions valida permissões do usuário
func (uc *SetGroupTopicUseCase) validatePermissions(ctx context.Context, client whatsapp.WhatsAppClient, groupJID types.JID) error {
	// Simular obtenção de informações do grupo
	groupInfo := &group.Group{
		JID:              groupJID,
		Name:             "Grupo Simulado",
		Topic:            "Tópico atual",
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
	return uc.permissionValidator.CanChangeGroupTopic(groupInfo, userJID)
}

// setGroupTopicViaWhatsApp define o tópico do grupo
func (uc *SetGroupTopicUseCase) setGroupTopicViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID, topic string) error {
	// TODO: Implementar método específico no WhatsAppClient
	uc.logger.WithFields(map[string]interface{}{
		"groupJid": groupJID.String(),
		"topic":    topic,
	}).Info().Msg("Group topic set (simulated)")

	return nil
}
