package session

import (
	"context"

	"zmeow/internal/domain/session"
	"zmeow/pkg/logger"
)

// ListSessionsUseCase implementa o caso de uso para listar sessões
type ListSessionsUseCase struct {
	sessionRepo session.SessionRepository
	logger      logger.Logger
}

// NewListSessionsUseCase cria uma nova instância do caso de uso
func NewListSessionsUseCase(
	sessionRepo session.SessionRepository,
	logger logger.Logger,
) *ListSessionsUseCase {
	return &ListSessionsUseCase{
		sessionRepo: sessionRepo,
		logger:      logger.WithComponent("list-sessions-usecase"),
	}
}

// Execute executa o caso de uso para listar sessões
func (uc *ListSessionsUseCase) Execute(ctx context.Context) ([]*session.Session, error) {
	uc.logger.Info().Msg("Listing all sessions")

	sessions, err := uc.sessionRepo.ListActive(ctx)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to list sessions")
		return nil, err
	}

	uc.logger.WithField("count", len(sessions)).Info().Msg("Sessions listed successfully")

	return sessions, nil
}