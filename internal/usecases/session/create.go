package session

import (
	"context"
	"time"

	"github.com/google/uuid"

	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// CreateSessionUseCase implementa o caso de uso para criar uma nova sessão
type CreateSessionUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewCreateSessionUseCase cria uma nova instância do caso de uso
func NewCreateSessionUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// CreateSessionRequest representa os dados necessários para criar uma sessão
type CreateSessionRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Webhook  string `json:"webhook,omitempty" validate:"omitempty,url"`
	ProxyURL string `json:"proxyUrl,omitempty" validate:"omitempty,url"`
}

// Execute executa o caso de uso para criar uma sessão
func (uc *CreateSessionUseCase) Execute(ctx context.Context, req CreateSessionRequest) (*session.Session, error) {
	uc.logger.WithFields(map[string]interface{}{
		"name":     req.Name,
		"webhook":  req.Webhook,
		"proxyUrl": req.ProxyURL,
	}).Info().Msg("Creating new session")

	// Verificar se uma sessão com esse nome já existe
	exists, err := uc.sessionRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to check if session exists")
		return nil, err
	}

	if exists {
		uc.logger.WithField("name", req.Name).Warn().Msg("Session with name already exists")
		return nil, session.ErrSessionAlreadyExists
	}

	// Criar nova sessão
	newSession := &session.Session{
		ID:        uuid.New(),
		Name:      req.Name,
		Status:    session.WhatsAppStatusDisconnected,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Webhook:   req.Webhook,
		ProxyURL:  req.ProxyURL,
	}

	// Salvar no banco de dados
	if err := uc.sessionRepo.Create(ctx, newSession); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to create session in database")
		return nil, err
	}

	// Registrar sessão no WhatsApp manager
	if err := uc.whatsappManager.RegisterSession(newSession.ID); err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to register session in WhatsApp manager")
		// Tentar remover a sessão do banco se falhou ao registrar
		if deleteErr := uc.sessionRepo.Delete(ctx, newSession.ID); deleteErr != nil {
			uc.logger.WithError(deleteErr).Error().Msg("Failed to rollback session creation")
		}
		return nil, err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": newSession.ID,
		"name":      newSession.Name,
	}).Info().Msg("Session created successfully")

	return newSession, nil
}