package session

import (
	"context"

	"github.com/google/uuid"

	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// GetQRCodeUseCase implementa o caso de uso para obter QR code de uma sessão
type GetQRCodeUseCase struct {
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewGetQRCodeUseCase cria uma nova instância do caso de uso
func NewGetQRCodeUseCase(
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *GetQRCodeUseCase {
	return &GetQRCodeUseCase{
		whatsappManager: whatsappManager,
		logger:          logger.WithComponent("get-qr-usecase"),
	}
}

// QRCodeResponse representa a resposta do QR code
type QRCodeResponse struct {
	QRCode string `json:"qrCode"`
	Status string `json:"status"`
}

// Execute executa o caso de uso para obter QR code
func (uc *GetQRCodeUseCase) Execute(ctx context.Context, sessionID uuid.UUID) (*QRCodeResponse, error) {
	uc.logger.WithField("sessionId", sessionID).Info().Msg("Getting QR code")

	// Verificar status da sessão
	status, err := uc.whatsappManager.GetSessionStatus(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get session status")
		return nil, err
	}

	// Se já estiver conectada, não precisa de QR code
	if status == "connected" {
		uc.logger.WithField("sessionId", sessionID).Info().Msg("Session already connected")
		return &QRCodeResponse{
			QRCode: "",
			Status: "connected",
		}, nil
	}

	// Obter QR code
	qrCode, err := uc.whatsappManager.GetQRCode(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get QR code")
		return nil, err
	}

	response := &QRCodeResponse{
		QRCode: qrCode,
		Status: status,
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"hasQRCode": qrCode != "",
		"status":    status,
	}).Info().Msg("QR code retrieved successfully")

	return response, nil
}
// QRCodeUseCase é um alias para GetQRCodeUseCase para compatibilidade
type QRCodeUseCase = GetQRCodeUseCase

// NewQRCodeUseCase é um alias para NewGetQRCodeUseCase para compatibilidade
var NewQRCodeUseCase = NewGetQRCodeUseCase