package group

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"zmeow/internal/domain/group"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/infra/whatsapp/services"
	"zmeow/pkg/logger"
)

// ListGroupsUseCase implementa o caso de uso para listar grupos
type ListGroupsUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewListGroupsUseCase cria uma nova instância do caso de uso
func NewListGroupsUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *ListGroupsUseCase {
	return &ListGroupsUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para listar grupos
func (uc *ListGroupsUseCase) Execute(ctx context.Context, sessionID uuid.UUID) ([]group.Group, error) {
	uc.logger.WithField("sessionId", sessionID).Info().Msg("Listing groups")

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

	// Listar grupos via WhatsApp usando GroupService
	groupService := services.NewGroupService(uc.whatsappManager, sessionID, uc.logger)
	groupPointers, err := groupService.ListGroups(ctx)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to list groups via WhatsApp")
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	// Converter ponteiros para valores
	groups := make([]group.Group, len(groupPointers))
	for i, groupPtr := range groupPointers {
		if groupPtr != nil {
			groups[i] = *groupPtr
		}
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":  sessionID,
		"groupCount": len(groups),
	}).Info().Msg("Groups listed successfully")

	return groups, nil
}

// listGroupsViaWhatsApp lista os grupos usando o cliente WhatsApp
func (uc *ListGroupsUseCase) listGroupsViaWhatsApp(client whatsapp.WhatsAppClient) ([]group.Group, error) {
	// TODO: Implementar método específico de listar grupos no WhatsAppClient
	// Por enquanto, vamos simular a listagem de grupos

	// Em uma implementação real, usaríamos algo como client.GetJoinedGroups()

	// Retornar lista vazia para desenvolvimento
	// Quando implementarmos o cliente real, isso será substituído
	groups := []group.Group{}

	return groups, nil
}
