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

// GetInviteLinkUseCase implementa o caso de uso para obter link de convite do grupo
type GetInviteLinkUseCase struct {
	sessionRepo         session.SessionRepository
	whatsappManager     whatsapp.WhatsAppManager
	permissionValidator *group.PermissionValidator
	logger              logger.Logger
}

// NewGetInviteLinkUseCase cria uma nova instância do caso de uso
func NewGetInviteLinkUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	permissionValidator *group.PermissionValidator,
	logger logger.Logger,
) *GetInviteLinkUseCase {
	return &GetInviteLinkUseCase{
		sessionRepo:         sessionRepo,
		whatsappManager:     whatsappManager,
		permissionValidator: permissionValidator,
		logger:              logger,
	}
}

// Execute executa o caso de uso para obter link de convite
func (uc *GetInviteLinkUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.GetGroupInviteLinkRequest) (*group.InviteLinkResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"reset":     req.Reset,
	}).Info().Msg("Getting group invite link")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return nil, err
	}

	// Verificar sessão e conectividade
	if err := uc.validateSession(ctx, sessionID); err != nil {
		return nil, err
	}

	// Obter serviço de grupos
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)

	// Obter link de convite via GroupService
	inviteLink, err := groupService.GetGroupInviteLink(ctx, req.GroupJID, req.Reset)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get invite link via GroupService")
		return nil, fmt.Errorf("failed to get invite link: %w", err)
	}

	// Extrair código do link
	code, _ := uc.ExtractCodeFromLink(inviteLink)

	response := &group.InviteLinkResponse{
		Details:    "Link de convite obtido com sucesso",
		InviteLink: inviteLink,
		Code:       code,
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":  sessionID,
		"groupJid":   req.GroupJID,
		"inviteLink": inviteLink,
		"code":       code,
		"reset":      req.Reset,
	}).Info().Msg("Invite link obtained successfully")

	return response, nil
}

// validateRequest valida a requisição
func (uc *GetInviteLinkUseCase) validateRequest(req group.GetGroupInviteLinkRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	// Validar formato do JID
	if !strings.Contains(req.GroupJID, "@") || !strings.HasSuffix(req.GroupJID, "@g.us") {
		return group.NewValidationError("groupJid", req.GroupJID, "invalid group JID format")
	}

	return nil
}

// validateSession valida a sessão
func (uc *GetInviteLinkUseCase) validateSession(ctx context.Context, sessionID uuid.UUID) error {
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
func (uc *GetInviteLinkUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
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
func (uc *GetInviteLinkUseCase) validatePermissions(ctx context.Context, client whatsapp.WhatsAppClient, groupJID types.JID) error {
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
	return uc.permissionValidator.CanGetGroupInviteLink(groupInfo, userJID)
}

// getInviteLinkViaWhatsApp obtém o link de convite via WhatsApp
func (uc *GetInviteLinkUseCase) getInviteLinkViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID, reset bool) (string, string, error) {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação

	// Em uma implementação real, seria algo como:
	// return client.GetGroupInviteLink(ctx, sessionID, groupJID, reset)

	// Simular código e link de convite
	code := uc.generateInviteCode()
	inviteLink := fmt.Sprintf("https://chat.whatsapp.com/%s", code)

	uc.logger.WithFields(map[string]interface{}{
		"groupJid":   groupJID.String(),
		"inviteLink": inviteLink,
		"code":       code,
		"reset":      reset,
	}).Info().Msg("Invite link obtained (simulated)")

	return inviteLink, code, nil
}

// generateInviteCode gera um código de convite simulado
func (uc *GetInviteLinkUseCase) generateInviteCode() string {
	// Simular código de convite (22 caracteres alfanuméricos)
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	code := make([]byte, 22)

	// Usar timestamp para gerar código único
	timestamp := time.Now().UnixNano()
	for i := range code {
		code[i] = chars[timestamp%int64(len(chars))]
		timestamp = timestamp / int64(len(chars))
	}

	return string(code)
}

// ExtractCodeFromLink extrai o código de um link de convite
func (uc *GetInviteLinkUseCase) ExtractCodeFromLink(inviteLink string) (string, error) {
	// Formatos válidos:
	// https://chat.whatsapp.com/CODIGO
	// https://wa.me/CODIGO
	// chat.whatsapp.com/CODIGO

	// Normalizar URL
	link := strings.TrimSpace(inviteLink)
	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "http://")

	// Extrair código
	if strings.HasPrefix(link, "chat.whatsapp.com/") {
		code := strings.TrimPrefix(link, "chat.whatsapp.com/")
		return uc.validateInviteCode(code)
	}

	if strings.HasPrefix(link, "wa.me/") {
		code := strings.TrimPrefix(link, "wa.me/")
		return uc.validateInviteCode(code)
	}

	// Se não tem prefixo, assumir que é apenas o código
	return uc.validateInviteCode(link)
}

// validateInviteCode valida o formato do código de convite
func (uc *GetInviteLinkUseCase) validateInviteCode(code string) (string, error) {
	// Remover caracteres especiais
	code = strings.TrimSpace(code)
	code = strings.TrimSuffix(code, "/")

	// Validar comprimento (códigos do WhatsApp têm 22 caracteres)
	if len(code) != 22 {
		return "", fmt.Errorf("invalid invite code length: expected 22 characters, got %d", len(code))
	}

	// Validar caracteres (apenas alfanuméricos)
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return "", fmt.Errorf("invalid invite code format: contains non-alphanumeric characters")
		}
	}

	return code, nil
}
