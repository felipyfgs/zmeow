package session

import (
	"context"

	"github.com/google/uuid"

	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// DeleteSessionUseCase implementa o caso de uso para deletar uma sessão
type DeleteSessionUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewDeleteSessionUseCase cria uma nova instância do caso de uso
func NewDeleteSessionUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *DeleteSessionUseCase {
	return &DeleteSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger.WithComponent("delete-session-usecase"),
	}
}

// Execute executa o caso de uso para deletar uma sessão
func (uc *DeleteSessionUseCase) Execute(ctx context.Context, sessionID uuid.UUID) error {
	uc.logger.WithField("sessionId", sessionID).Info().Msg("Deleting session")

	// Verificar se a sessão existe
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if err == session.ErrSessionNotFound {
			uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session not found")
			return session.ErrSessionNotFound
		}
		uc.logger.WithError(err).Error().Msg("Failed to get session for deletion")
		return err
	}

	// Desconectar sessão do WhatsApp se estiver conectada
	if uc.whatsappManager.IsConnected(sessionID) {
		if err := uc.whatsappManager.DisconnectSession(sessionID); err != nil {
			uc.logger.WithError(err).Warn().Msg("Failed to disconnect WhatsApp session")
		}
	}

	// Remover sessão do manager do WhatsApp
	if err := uc.whatsappManager.RemoveSession(sessionID); err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to remove session from WhatsApp manager")
	}

	// Deletar do banco de dados
	if err := uc.sessionRepo.Delete(ctx, sessionID); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to delete session from database")
		return err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"name":      sess.Name,
	}).Info().Msg("Session deleted successfully")

	return nil
}