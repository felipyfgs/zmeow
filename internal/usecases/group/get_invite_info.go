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

// GetInviteInfoUseCase implementa o caso de uso para obter informações de um convite
type GetInviteInfoUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewGetInviteInfoUseCase cria uma nova instância do caso de uso
func NewGetInviteInfoUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *GetInviteInfoUseCase {
	return &GetInviteInfoUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para obter informações do convite
func (uc *GetInviteInfoUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.GetGroupInviteInfoRequest) (*group.InviteInfoResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"code":      req.Code,
	}).Info().Msg("Getting invite info")

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

	// Obter informações do convite via GroupService
	inviteInfo, err := groupService.GetGroupInviteInfo(ctx, code)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get invite info via GroupService")
		return nil, fmt.Errorf("failed to get invite info: %w", err)
	}

	response := &group.InviteInfoResponse{
		Details:     "Informações do convite obtidas com sucesso",
		GroupName:   inviteInfo.GroupName,
		GroupJID:    inviteInfo.GroupJID.String(),
		CreatedBy:   inviteInfo.CreatedBy.String(),
		CreatedAt:   time.Unix(inviteInfo.CreatedAt, 0).Format(time.RFC3339),
		Description: inviteInfo.Description,
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"code":      code,
		"groupName": inviteInfo.GroupName,
		"groupJid":  inviteInfo.GroupJID.String(),
	}).Info().Msg("Invite info obtained successfully")

	return response, nil
}

// validateRequest valida a requisição
func (uc *GetInviteInfoUseCase) validateRequest(req group.GetGroupInviteInfoRequest) error {
	if req.Code == "" {
		return group.NewValidationError("code", req.Code, "invite code is required")
	}

	return nil
}

// validateSession valida a sessão
func (uc *GetInviteInfoUseCase) validateSession(ctx context.Context, sessionID uuid.UUID) error {
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
func (uc *GetInviteInfoUseCase) extractAndValidateCode(input string) (string, error) {
	// Se parece com um link, extrair o código
	if strings.Contains(input, "chat.whatsapp.com") || strings.Contains(input, "wa.me") {
		return uc.extractCodeFromLink(input)
	}

	// Caso contrário, validar como código direto
	return uc.validateInviteCode(input)
}

// extractCodeFromLink extrai o código de um link de convite
func (uc *GetInviteInfoUseCase) extractCodeFromLink(inviteLink string) (string, error) {
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

	return "", fmt.Errorf("invalid invite link format")
}

// validateInviteCode valida o formato do código de convite
func (uc *GetInviteInfoUseCase) validateInviteCode(code string) (string, error) {
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

// getInviteInfoViaWhatsApp obtém informações do convite via WhatsApp
func (uc *GetInviteInfoUseCase) getInviteInfoViaWhatsApp(client whatsapp.WhatsAppClient, code string) (*group.InviteInfo, error) {
	// TODO: Implementar método específico no WhatsAppClient
	// Por enquanto, simular a operação

	// Em uma implementação real, seria algo como:
	// return client.GetGroupInfoFromLink(ctx, sessionID, code)

	// Simular informações do convite
	groupJID := types.NewJID(fmt.Sprintf("group_%d", time.Now().Unix()), types.GroupServer)
	createdBy := types.NewJID("5511999999999", types.DefaultUserServer)

	inviteInfo := &group.InviteInfo{
		Code:        code,
		GroupJID:    groupJID,
		GroupName:   "Grupo de Exemplo",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now().Unix(),
		Description: "Este é um grupo de exemplo criado para demonstração",
	}

	uc.logger.WithFields(map[string]interface{}{
		"code":      code,
		"groupJid":  groupJID.String(),
		"groupName": inviteInfo.GroupName,
		"createdBy": createdBy.String(),
	}).Info().Msg("Invite info obtained (simulated)")

	return inviteInfo, nil
}

// IsValidInviteCode verifica se um código de convite é válido
func (uc *GetInviteInfoUseCase) IsValidInviteCode(code string) bool {
	_, err := uc.validateInviteCode(code)
	return err == nil
}

// IsValidInviteLink verifica se um link de convite é válido
func (uc *GetInviteInfoUseCase) IsValidInviteLink(link string) bool {
	_, err := uc.extractCodeFromLink(link)
	return err == nil
}

// ParseInviteInput analisa entrada de convite (link ou código)
func (uc *GetInviteInfoUseCase) ParseInviteInput(input string) (string, string, error) {
	input = strings.TrimSpace(input)

	// Se contém domínio do WhatsApp, é um link
	if strings.Contains(input, "chat.whatsapp.com") || strings.Contains(input, "wa.me") {
		code, err := uc.extractCodeFromLink(input)
		if err != nil {
			return "", "", err
		}
		return "link", code, nil
	}

	// Caso contrário, assumir que é código direto
	code, err := uc.validateInviteCode(input)
	if err != nil {
		return "", "", err
	}
	return "code", code, nil
}
