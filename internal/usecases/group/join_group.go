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

// JoinGroupUseCase implementa o caso de uso para entrar em um grupo via convite
type JoinGroupUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewJoinGroupUseCase cria uma nova instância do caso de uso
func NewJoinGroupUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *JoinGroupUseCase {
	return &JoinGroupUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para entrar em um grupo
func (uc *JoinGroupUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.JoinGroupRequest) (*group.Group, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"code":      req.Code,
	}).Info().Msg("Joining group via invite")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return nil, err
	}

	// Verificar sessão e conectividade
	if err := uc.validateSession(ctx, sessionID); err != nil {
		return nil, err
	}

	// Extrair código do link se necessário
	code, err := uc.extractAndValidateCode(req.Code)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid invite code")
		return nil, fmt.Errorf("invalid invite code: %w", err)
	}

	// Obter serviço de grupos
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)

	// Entrar no grupo via GroupService
	groupJID, err := groupService.JoinGroupWithLink(ctx, code)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to join group via GroupService")
		return nil, fmt.Errorf("failed to join group: %w", err)
	}

	// Obter informações do grupo após entrar
	groupInfo, err := groupService.GetGroupInfo(ctx, groupJID)
	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to get group info after joining, returning basic info")
		// Se não conseguir obter informações completas, retornar informações básicas
		groupInfo = &group.Group{
			JID:              groupJID,
			Name:             "Grupo Entrado via Convite",
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
			InviteCode:       code,
			ParticipantCount: 0,
		}
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"code":      code,
		"groupJid":  groupJID.String(),
		"groupName": groupInfo.Name,
	}).Info().Msg("Joined group successfully")

	return groupInfo, nil
}

// validateRequest valida a requisição
func (uc *JoinGroupUseCase) validateRequest(req group.JoinGroupRequest) error {
	if req.Code == "" {
		return group.NewValidationError("code", req.Code, "invite code is required")
	}

	return nil
}

// validateSession valida a sessão
func (uc *JoinGroupUseCase) validateSession(ctx context.Context, sessionID uuid.UUID) error {
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

// extractAndValidateCode extrai e valida o código de convite
func (uc *JoinGroupUseCase) extractAndValidateCode(input string) (string, error) {
	// Se parece com um link, extrair o código
	if strings.Contains(input, "chat.whatsapp.com") || strings.Contains(input, "wa.me") {
		return uc.extractCodeFromLink(input)
	}

	// Caso contrário, validar como código direto
	return uc.validateInviteCode(input)
}

// extractCodeFromLink extrai o código de um link de convite
func (uc *JoinGroupUseCase) extractCodeFromLink(inviteLink string) (string, error) {
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

	return "", group.NewInviteError(inviteLink, types.JID{}, "extract_code",
		fmt.Errorf("invalid invite link format"))
}

// validateInviteCode valida o formato do código de convite
func (uc *JoinGroupUseCase) validateInviteCode(code string) (string, error) {
	// Remover caracteres especiais
	code = strings.TrimSpace(code)
	code = strings.TrimSuffix(code, "/")

	// Validar comprimento
	if len(code) != 22 {
		return "", group.NewInviteError(code, types.JID{}, "validate",
			fmt.Errorf("invalid code length: expected 22 characters, got %d", len(code)))
	}

	// Validar caracteres (apenas alfanuméricos)
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return "", group.NewInviteError(code, types.JID{}, "validate",
				fmt.Errorf("invalid code format: contains non-alphanumeric characters"))
		}
	}

	return code, nil
}

// validateInvite valida se o convite é válido e não expirou
func (uc *JoinGroupUseCase) validateInvite(ctx context.Context, client whatsapp.WhatsAppClient, code string) error {
	// TODO: Implementar validação real do convite
	// Por enquanto, simular validação

	// Em uma implementação real, seria algo como:
	// inviteInfo, err := client.GetGroupInfoFromLink(ctx, sessionID, code)
	// if err != nil {
	//     return group.NewInviteError(code, types.JID{}, "validate", err)
	// }
	//
	// // Verificar se não expirou
	// if inviteInfo.ExpiresAt != nil && time.Now().Unix() > *inviteInfo.ExpiresAt {
	//     return group.ErrInviteExpired
	// }

	uc.logger.WithField("code", code).Info().Msg("Invite validated (simulated)")

	return nil
}

// joinGroupViaWhatsApp entra no grupo via WhatsApp
func (uc *JoinGroupUseCase) joinGroupViaWhatsApp(client whatsapp.WhatsAppClient, code string) (*group.Group, error) {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação

	// Em uma implementação real, seria algo como:
	// return client.JoinGroupWithLink(ctx, sessionID, code)

	// Simular entrada no grupo
	groupJID := types.NewJID(fmt.Sprintf("group_%d", time.Now().Unix()), types.GroupServer)
	userJID := types.NewJID("5511999999999", types.DefaultUserServer)

	groupInfo := &group.Group{
		JID:   groupJID,
		Name:  "Grupo Entrado via Convite",
		Topic: "Bem-vindo ao grupo!",
		Participants: []group.Participant{
			{
				JID:          userJID,
				IsAdmin:      false,
				IsSuperAdmin: false,
				JoinedAt:     time.Now(),
			},
		},
		Admins:           []types.JID{},
		Owner:            types.NewJID("5511888888888", types.DefaultUserServer),
		CreatedAt:        time.Now().Add(-24 * time.Hour),
		IsAnnounce:       false,
		IsLocked:         false,
		IsEphemeral:      false,
		EphemeralTimer:   0,
		PictureID:        "",
		InviteCode:       code,
		ParticipantCount: 1,
	}

	uc.logger.WithFields(map[string]interface{}{
		"code":      code,
		"groupJid":  groupJID.String(),
		"groupName": groupInfo.Name,
		"userJid":   userJID.String(),
	}).Info().Msg("Joined group (simulated)")

	return groupInfo, nil
}

// CanJoinGroup verifica se é possível entrar em um grupo
func (uc *JoinGroupUseCase) CanJoinGroup(ctx context.Context, client whatsapp.WhatsAppClient, code string) error {
	// Validar formato do código
	if _, err := uc.validateInviteCode(code); err != nil {
		return err
	}

	// Validar se o convite existe e é válido
	return uc.validateInvite(ctx, client, code)
}

// GetJoinPreview obtém preview do grupo antes de entrar
func (uc *JoinGroupUseCase) GetJoinPreview(ctx context.Context, client whatsapp.WhatsAppClient, code string) (*group.InviteInfo, error) {
	// Validar código
	validCode, err := uc.validateInviteCode(code)
	if err != nil {
		return nil, err
	}

	// TODO: Implementar obtenção de preview
	// Por enquanto, simular
	groupJID := types.NewJID(fmt.Sprintf("group_%s", validCode[:8]), types.GroupServer)
	createdBy := types.NewJID("5511888888888", types.DefaultUserServer)

	return &group.InviteInfo{
		Code:        validCode,
		GroupJID:    groupJID,
		GroupName:   "Preview do Grupo",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now().Add(-24 * time.Hour).Unix(),
		Description: "Este é um preview do grupo que você pode entrar",
	}, nil
}
