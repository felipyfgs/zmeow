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

// SetGroupNameUseCase implementa o caso de uso para definir nome do grupo
type SetGroupNameUseCase struct {
	sessionRepo         session.SessionRepository
	whatsappManager     whatsapp.WhatsAppManager
	permissionValidator *group.PermissionValidator
	logger              logger.Logger
}

// NewSetGroupNameUseCase cria uma nova instância do caso de uso
func NewSetGroupNameUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	permissionValidator *group.PermissionValidator,
	logger logger.Logger,
) *SetGroupNameUseCase {
	return &SetGroupNameUseCase{
		sessionRepo:         sessionRepo,
		whatsappManager:     whatsappManager,
		permissionValidator: permissionValidator,
		logger:              logger,
	}
}

// Execute executa o caso de uso para definir nome do grupo
func (uc *SetGroupNameUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.SetGroupNameRequest) error {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"newName":   req.Name,
	}).Info().Msg("Setting group name")

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

	// Definir nome do grupo via GroupService (que já inclui validações de permissão)
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	if err := groupService.SetGroupName(ctx, req.GroupJID, req.Name); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to set group name via GroupService")
		return fmt.Errorf("failed to set group name: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"newName":   req.Name,
	}).Info().Msg("Group name set successfully")

	return nil
}

// validateRequest valida a requisição de definir nome do grupo
func (uc *SetGroupNameUseCase) validateRequest(req group.SetGroupNameRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	if req.Name == "" {
		return group.NewValidationError("name", req.Name, "group name is required")
	}

	if len(req.Name) > 25 {
		return group.NewValidationError("name", req.Name, "group name must be at most 25 characters")
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
func (uc *SetGroupNameUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
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

// getGroupInfo obtém informações do grupo
func (uc *SetGroupNameUseCase) getGroupInfo(ctx context.Context, client whatsapp.WhatsAppClient, groupJID types.JID) (*group.Group, error) {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a obtenção de informações

	groupInfo := &group.Group{
		JID:              groupJID,
		Name:             "Nome Atual do Grupo",
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

	return groupInfo, nil
}

// getUserJID obtém o JID do usuário atual
func (uc *SetGroupNameUseCase) getUserJID(client whatsapp.WhatsAppClient) types.JID {
	// TODO: Implementar método para obter JID do usuário no WhatsAppClient
	// Por enquanto, retornar JID simulado
	return types.NewJID("5511999999999", types.DefaultUserServer)
}

// setGroupNameViaWhatsApp define o nome do grupo usando o cliente WhatsApp
func (uc *SetGroupNameUseCase) setGroupNameViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID, name string) error {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação

	// Em uma implementação real, seria algo como:
	// return client.SetGroupName(ctx, sessionID, groupJID, name)

	uc.logger.WithFields(map[string]interface{}{
		"groupJid": groupJID.String(),
		"newName":  name,
	}).Info().Msg("Group name set (simulated)")

	return nil
}
