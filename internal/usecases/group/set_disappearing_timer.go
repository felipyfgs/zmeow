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

// SetDisappearingTimerUseCase implementa o caso de uso para configurar timer de desaparecimento
type SetDisappearingTimerUseCase struct {
	sessionRepo         session.SessionRepository
	whatsappManager     whatsapp.WhatsAppManager
	permissionValidator *group.PermissionValidator
	logger              logger.Logger
}

// NewSetDisappearingTimerUseCase cria uma nova instância do caso de uso
func NewSetDisappearingTimerUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	permissionValidator *group.PermissionValidator,
	logger logger.Logger,
) *SetDisappearingTimerUseCase {
	return &SetDisappearingTimerUseCase{
		sessionRepo:         sessionRepo,
		whatsappManager:     whatsappManager,
		permissionValidator: permissionValidator,
		logger:              logger,
	}
}

// Execute executa o caso de uso para configurar timer de desaparecimento
func (uc *SetDisappearingTimerUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req group.SetDisappearingTimerRequest) error {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"duration":  req.Duration,
	}).Info().Msg("Setting disappearing timer")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return err
	}

	// Converter duração string para time.Duration
	duration, err := uc.parseDuration(req.Duration)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid duration format")
		return fmt.Errorf("invalid duration: %w", err)
	}

	// Configurar timer via GroupService (que já inclui validações de permissão)
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	if err := groupService.SetDisappearingTimer(ctx, req.GroupJID, duration); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to set disappearing timer via GroupService")
		return fmt.Errorf("failed to set disappearing timer: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"groupJid":  req.GroupJID,
		"duration":  req.Duration,
		"seconds":   duration.Seconds(),
	}).Info().Msg("Disappearing timer set successfully")

	return nil
}

// validateRequest valida a requisição
func (uc *SetDisappearingTimerUseCase) validateRequest(req group.SetDisappearingTimerRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "group JID is required")
	}

	if req.Duration == "" {
		return group.NewValidationError("duration", req.Duration, "duration is required")
	}

	// Validar formato do JID
	if !strings.Contains(req.GroupJID, "@") || !strings.HasSuffix(req.GroupJID, "@g.us") {
		return group.NewValidationError("groupJid", req.GroupJID, "invalid group JID format")
	}

	// Validar duração
	if _, err := uc.parseDuration(req.Duration); err != nil {
		return group.NewValidationError("duration", req.Duration, "invalid duration format")
	}

	return nil
}

// validateSession valida a sessão
func (uc *SetDisappearingTimerUseCase) validateSession(ctx context.Context, sessionID uuid.UUID) error {
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
func (uc *SetDisappearingTimerUseCase) parseGroupJID(groupJIDStr string) (types.JID, error) {
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
func (uc *SetDisappearingTimerUseCase) validatePermissions(ctx context.Context, client whatsapp.WhatsAppClient, groupJID types.JID) error {
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
	return uc.permissionValidator.CanSetDisappearingTimer(groupInfo, userJID)
}

// parseDuration converte string de duração para time.Duration
func (uc *SetDisappearingTimerUseCase) parseDuration(durationStr string) (time.Duration, error) {
	// Normalizar string
	durationStr = strings.ToLower(strings.TrimSpace(durationStr))

	switch durationStr {
	case "off", "0", "disable", "disabled":
		return 0, nil
	case "24h", "1d", "1day", "day":
		return 24 * time.Hour, nil
	case "7d", "7days", "week", "1week":
		return 7 * 24 * time.Hour, nil
	case "90d", "90days", "3months":
		return 90 * 24 * time.Hour, nil
	default:
		// Tentar parsing direto para casos como "2h", "30m", etc.
		if duration, err := time.ParseDuration(durationStr); err == nil {
			// Validar se está dentro dos limites permitidos
			if duration < 0 {
				return 0, fmt.Errorf("duration cannot be negative")
			}
			if duration > 90*24*time.Hour {
				return 0, fmt.Errorf("duration cannot exceed 90 days")
			}
			return duration, nil
		}

		return 0, fmt.Errorf("invalid duration format: %s (valid: off, 24h, 7d, 90d)", durationStr)
	}
}

// setDisappearingTimerViaWhatsApp configura o timer via WhatsApp
func (uc *SetDisappearingTimerUseCase) setDisappearingTimerViaWhatsApp(client whatsapp.WhatsAppClient, groupJID types.JID, duration time.Duration) error {
	// TODO: Implementar método específico no WhatsAppClient
	uc.logger.WithFields(map[string]interface{}{
		"groupJid": groupJID.String(),
		"duration": duration.String(),
		"seconds":  duration.Seconds(),
	}).Info().Msg("Disappearing timer set (simulated)")

	return nil
}

// GetValidDurations retorna as durações válidas para timer de desaparecimento
func (uc *SetDisappearingTimerUseCase) GetValidDurations() []string {
	return []string{
		"off", // Desabilitar timer
		"24h", // 24 horas
		"7d",  // 7 dias
		"90d", // 90 dias
	}
}

// FormatDuration formata uma duração para string legível
func (uc *SetDisappearingTimerUseCase) FormatDuration(duration time.Duration) string {
	if duration == 0 {
		return "off"
	}

	hours := duration.Hours()
	if hours == 24 {
		return "24h"
	}
	if hours == 24*7 {
		return "7d"
	}
	if hours == 24*90 {
		return "90d"
	}

	// Para durações customizadas
	if hours < 24 {
		return fmt.Sprintf("%.0fh", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%.0fd", days)
}
