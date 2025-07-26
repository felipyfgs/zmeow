package session

import (
	"context"

	"github.com/google/uuid"

	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// ConnectSessionUseCase implementa o caso de uso para conectar uma sessão
type ConnectSessionUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewConnectSessionUseCase cria uma nova instância do caso de uso
func NewConnectSessionUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *ConnectSessionUseCase {
	return &ConnectSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// ConnectResponse representa a resposta da conexão
type ConnectResponse struct {
	Status string `json:"status"`
	QRCode string `json:"qrCode,omitempty"`
}

// Execute executa o caso de uso para conectar uma sessão
func (uc *ConnectSessionUseCase) Execute(ctx context.Context, sessionID uuid.UUID) (*ConnectResponse, error) {
	uc.logger.WithField("sessionId", sessionID).Info().Msg("Connecting session")

	// Buscar a sessão
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return nil, session.ErrSessionNotFound
	}

	// Verificar se a sessão pode ser conectada
	if !sess.CanConnect() {
		uc.logger.WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"status":    sess.Status,
			"isActive":  sess.IsActive,
		}).Warn().Msg("Session cannot be connected")

		if !sess.IsActive {
			return nil, session.ErrSessionInactive
		}
		if sess.Status == session.WhatsAppStatusConnected {
			return nil, session.ErrSessionAlreadyConnected
		}
		if sess.Status == session.WhatsAppStatusConnecting {
			return nil, session.ErrSessionConnecting
		}
	}

	// Atualizar status para connecting
	sess.SetConnecting()
	if err := uc.sessionRepo.Update(ctx, sess); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to update session status to connecting")
		return nil, err
	}

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to get WhatsApp client, attempting to register session")
		// Tentar registrar a sessão novamente se não existir
		if regErr := uc.whatsappManager.RegisterSession(sessionID); regErr != nil {
			uc.logger.WithError(regErr).Error().Msg("Failed to register session in WhatsApp manager")
			// Reverter status
			sess.SetDisconnected()
			uc.sessionRepo.Update(ctx, sess)
			return nil, regErr
		}
		// Tentar obter cliente novamente após registro
		client, err = uc.whatsappManager.GetClient(sessionID)
		if err != nil {
			uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client after registration")
			// Reverter status
			sess.SetDisconnected()
			uc.sessionRepo.Update(ctx, sess)
			return nil, err
		}
	}

	// Iniciar conexão
	if err := client.Connect(context.Background(), sessionID); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to connect WhatsApp client")
		// Reverter status
		sess.SetDisconnected()
		uc.sessionRepo.Update(ctx, sess)
		return nil, err
	}

	response := &ConnectResponse{
		Status: string(session.WhatsAppStatusConnecting),
	}

	// Sempre tentar obter o QR Code após conectar
	// Isso é similar ao wuzapi que chama GetQRChannel logo após Connect
	qrCode, err := client.GetQRCode(ctx, sessionID)
	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to get QR code, but connection started")
	} else if qrCode != "" {
		response.QRCode = qrCode
		uc.logger.WithField("sessionId", sessionID).Info().Msg("QR Code generated and displayed in terminal")
	} else {
		uc.logger.WithField("sessionId", sessionID).Info().Msg("No QR Code needed - session may already be authenticated")
	}

	uc.logger.WithField("sessionId", sessionID).Info().Msg("Session connection initiated")

	return response, nil
}
