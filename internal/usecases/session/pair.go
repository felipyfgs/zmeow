package session

import (
	"context"

	"github.com/google/uuid"

	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// PairPhoneUseCase implementa o caso de uso para pareamento por telefone
type PairPhoneUseCase struct {
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewPairPhoneUseCase cria uma nova instância do caso de uso
func NewPairPhoneUseCase(
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *PairPhoneUseCase {
	return &PairPhoneUseCase{
		whatsappManager: whatsappManager,
		logger:          logger.WithComponent("pair-phone-usecase"),
	}
}

// PairPhoneResponse representa a resposta do pareamento
type PairPhoneResponse struct {
	PairingCode string `json:"pairingCode"`
	PhoneNumber string `json:"phoneNumber"`
}

// Execute executa o caso de uso para pareamento por telefone
func (uc *PairPhoneUseCase) Execute(ctx context.Context, sessionID uuid.UUID, phoneNumber string) (*PairPhoneResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     phoneNumber,
	}).Info().Msg("Initiating phone pairing")

	// Validar número de telefone
	if phoneNumber == "" {
		uc.logger.Error().Msg("Phone number is required")
		return nil, whatsapp.ErrInvalidPhoneNumber
	}

	// Verificar se a sessão já está conectada
	if uc.whatsappManager.IsConnected(sessionID) {
		uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session already connected")
		return nil, whatsapp.ErrSessionAlreadyConnected
	}

	// Realizar pareamento
	pairingCode, err := uc.whatsappManager.PairPhone(sessionID, phoneNumber)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to pair phone")
		return nil, err
	}

	response := &PairPhoneResponse{
		PairingCode: pairingCode,
		PhoneNumber: phoneNumber,
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     phoneNumber,
		"hasCode":   pairingCode != "",
	}).Info().Msg("Phone pairing initiated successfully")

	return response, nil
}