package session

import (
	"context"

	"github.com/google/uuid"

	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// DisconnectSessionUseCase implementa o caso de uso para desconectar uma sessão
type DisconnectSessionUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewDisconnectSessionUseCase cria uma nova instância do caso de uso
func NewDisconnectSessionUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *DisconnectSessionUseCase {
	return &DisconnectSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger.WithComponent("disconnect-session-usecase"),
	}
}

// DisconnectResponse representa a resposta da desconexão
type DisconnectResponse struct {
	Status string `json:"status"`
}

// Execute executa o caso de uso para desconectar uma sessão
func (uc *DisconnectSessionUseCase) Execute(ctx context.Context, sessionID uuid.UUID) (*DisconnectResponse, error) {
	uc.logger.WithField("sessionId", sessionID).Info().Msg("Disconnecting session")

	// Buscar a sessão
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if err == session.ErrSessionNotFound {
			uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session not found")
			return nil, session.ErrSessionNotFound
		}
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return nil, err
	}

	// Verificar se a sessão está conectada
	if sess.Status == session.WhatsAppStatusDisconnected {
		uc.logger.WithField("sessionId", sessionID).Info().Msg("Session already disconnected")
		return &DisconnectResponse{
			Status: string(session.WhatsAppStatusDisconnected),
		}, nil
	}

	// Desconectar do WhatsApp
	if err := uc.whatsappManager.DisconnectSession(sessionID); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to disconnect WhatsApp session")
		return nil, err
	}

	// Atualizar status no banco
	sess.SetDisconnected()
	if err := uc.sessionRepo.Update(ctx, sess); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to update session status")
		return nil, err
	}

	uc.logger.WithField("sessionId", sessionID).Info().Msg("Session disconnected successfully")

	return &DisconnectResponse{
		Status: string(session.WhatsAppStatusDisconnected),
	}, nil
}