package session

import (
	"context"

	"github.com/google/uuid"

	"zmeow/internal/domain/session"
	"zmeow/pkg/logger"
)

// GetSessionUseCase implementa o caso de uso para obter uma sessão específica
type GetSessionUseCase struct {
	sessionRepo session.SessionRepository
	logger      logger.Logger
}

// NewGetSessionUseCase cria uma nova instância do caso de uso
func NewGetSessionUseCase(
	sessionRepo session.SessionRepository,
	logger logger.Logger,
) *GetSessionUseCase {
	return &GetSessionUseCase{
		sessionRepo: sessionRepo,
		logger:      logger.WithComponent("get-session-usecase"),
	}
}

// Execute executa o caso de uso para obter uma sessão
func (uc *GetSessionUseCase) Execute(ctx context.Context, sessionID uuid.UUID) (*session.Session, error) {
	uc.logger.WithField("sessionId", sessionID).Info().Msg("Getting session")

	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if err == session.ErrSessionNotFound {
			uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session not found")
			return nil, session.ErrSessionNotFound
		}
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return nil, err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"name":      sess.Name,
		"status":    sess.Status,
	}).Info().Msg("Session retrieved successfully")

	return sess, nil
}