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
	"zmeow/internal/infra/media"
	"zmeow/internal/infra/whatsapp/services"
	"zmeow/pkg/logger"
)

// SetGroupPhotoUseCase implementa o caso de uso para definir foto do grupo
type SetGroupPhotoUseCase struct {
	sessionRepo         session.SessionRepository
	whatsappManager     whatsapp.WhatsAppManager
	imageProcessor      *media.ImageProcessor
	permissionValidator *group.PermissionValidator
	logger              logger.Logger
}

// NewSetGroupPhotoUseCase cria uma nova instância do caso de uso
func NewSetGroupPhotoUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	imageProcessor *media.ImageProcessor,
	permissionValidator *group.PermissionValidator,
	logger logger.Logger,
) *SetGroupPhotoUseCase {
	return &SetGroupPhotoUseCase{
		sessionRepo:         sessionRepo,
		whatsappManager:     whatsappManager,
		imageProcessor:      imageProcessor,
		permissionValidator: permissionValidator,
		logger:              logger,
	}
}

// Execute executa o caso de uso para definir foto do grupo
func (uc *SetGroupPhotoUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.SetGroupPhotoRequest) (string, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
	}).Info().Msg("Setting group photo")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return "", err
	}

	// Verificar se a sessão existe
	_, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return "", fmt.Errorf("session not found: %w", err)
	}

	// Processar imagem
	var imageInfo *media.ImageInfo

	if req.Image != "" {
		// Processar imagem base64
		imageInfo, err = uc.imageProcessor.ProcessBase64Image(req.Image)
	} else if req.ImageURL != "" {
		// Processar imagem via URL
		imageInfo, err = uc.imageProcessor.ProcessImageURL(req.ImageURL)
	}

	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to process image")
		return "", fmt.Errorf("failed to process image: %w", err)
	}

	// Definir foto do grupo via GroupService (que já inclui validações de permissão)
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	pictureID, err := groupService.SetGroupPhoto(ctx, req.GroupJID, imageInfo.Data)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to set group photo via GroupService")
		return "", fmt.Errorf("failed to set group photo: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"groupJid":    req.GroupJID,
		"pictureId":   pictureID,
		"imageSize":   imageInfo.Size,
		"imageFormat": imageInfo.Format,
	}).Info().Msg("Group photo set successfully")

	return pictureID, nil
}

// validateRequest valida a requisição de definir foto do grupo
func (uc *SetGroupPhotoUseCase) validateRequest(req group.SetGroupPhotoRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	if req.Image == "" && req.ImageURL == "" {
		return group.NewValidationError("image", "", "image data or image URL is required")
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
func (uc *SetGroupPhotoUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
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
func (uc *SetGroupPhotoUseCase) getGroupInfo(ctx context.Context, client whatsapp.WhatsAppClient, groupJID types.JID) (*group.Group, error) {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a obtenção de informações

	// Em uma implementação real, seria algo como:
	// return client.GetGroupInfo(ctx, sessionID, groupJID)

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

	return groupInfo, nil
}

// getUserJID obtém o JID do usuário atual
func (uc *SetGroupPhotoUseCase) getUserJID(client whatsapp.WhatsAppClient) types.JID {
	// TODO: Implementar método para obter JID do usuário no WhatsAppClient
	// Por enquanto, retornar JID simulado

	// Em uma implementação real, seria algo como:
	// return client.GetOwnJID()

	return types.NewJID("5511999999999", types.DefaultUserServer)
}

// setGroupPhotoViaWhatsApp define a foto do grupo usando o cliente WhatsApp
func (uc *SetGroupPhotoUseCase) setGroupPhotoViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID, imageData []byte) (string, error) {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação

	// Em uma implementação real, seria algo como:
	// return client.SetGroupPhoto(ctx, sessionID, groupJID, imageData)

	// Simular ID da foto
	pictureID := fmt.Sprintf("pic_%d", time.Now().Unix())

	uc.logger.WithFields(map[string]interface{}{
		"groupJid":  groupJID.String(),
		"pictureId": pictureID,
		"imageSize": len(imageData),
	}).Info().Msg("Group photo set (simulated)")

	return pictureID, nil
}
