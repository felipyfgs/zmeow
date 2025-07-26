package session

import (
	"context"
	"time"

	"github.com/google/uuid"

	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// GetStatusUseCase implementa o caso de uso para obter o status de uma sessão
type GetStatusUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewGetStatusUseCase cria uma nova instância do caso de uso
func NewGetStatusUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *GetStatusUseCase {
	return &GetStatusUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger.WithComponent("get-status-usecase"),
	}
}

// StatusResponse representa a resposta do status
type StatusResponse struct {
	Status   string     `json:"status"`
	JID      string     `json:"jid,omitempty"`
	LastSeen *time.Time `json:"lastSeen,omitempty"`
}

// Execute executa o caso de uso para obter status da sessão
func (uc *GetStatusUseCase) Execute(ctx context.Context, sessionID uuid.UUID) (*StatusResponse, error) {
	uc.logger.WithField("sessionId", sessionID).Info().Msg("Getting session status")

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

	// Obter status do WhatsApp manager
	whatsappStatus, err := uc.whatsappManager.GetSessionStatus(sessionID)
	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to get WhatsApp status, using database status")
		whatsappStatus = string(sess.Status)
	}

	// Sincronizar status se necessário
	if whatsappStatus != string(sess.Status) {
		uc.logger.WithFields(map[string]interface{}{
			"sessionId":      sessionID,
			"dbStatus":       sess.Status,
			"whatsappStatus": whatsappStatus,
		}).Info().Msg("Synchronizing session status")

		// Atualizar status no banco
		sess.Status = session.WhatsAppSessionStatus(whatsappStatus)
		if whatsappStatus == string(session.WhatsAppStatusConnected) {
			now := time.Now()
			sess.LastSeen = &now
		}
		sess.UpdatedAt = time.Now()

		if updateErr := uc.sessionRepo.Update(ctx, sess); updateErr != nil {
			uc.logger.WithError(updateErr).Warn().Msg("Failed to update session status in database")
		}
	}

	response := &StatusResponse{
		Status:   whatsappStatus,
		JID:      sess.WaJID,
		LastSeen: sess.LastSeen,
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"status":    whatsappStatus,
		"jid":       sess.WaJID,
	}).Info().Msg("Session status retrieved successfully")

	return response, nil
}
