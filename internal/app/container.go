package app

import (
	"github.com/uptrace/bun"

	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/http/handlers"
	"zmeow/internal/infra/database"
	sessionUseCases "zmeow/internal/usecases/session"
	"zmeow/pkg/logger"
)

// Container gerencia todas as dependências da aplicação
type Container struct {
	// Database
	DB *bun.DB

	// Repositories
	SessionRepo session.SessionRepository

	// WhatsApp
	WhatsAppManager whatsapp.WhatsAppManager

	// Use Cases
	CreateSessionUC     *sessionUseCases.CreateSessionUseCase
	ListSessionsUC      *sessionUseCases.ListSessionsUseCase
	GetSessionUC        *sessionUseCases.GetSessionUseCase
	DeleteSessionUC     *sessionUseCases.DeleteSessionUseCase
	ConnectSessionUC    *sessionUseCases.ConnectSessionUseCase
	DisconnectSessionUC *sessionUseCases.DisconnectSessionUseCase
	QRCodeUC            *sessionUseCases.GetQRCodeUseCase
	PairPhoneUC         *sessionUseCases.PairPhoneUseCase
	SetProxyUC          *sessionUseCases.SetProxyUseCase
	GetStatusUC         *sessionUseCases.GetStatusUseCase

	// Handlers
	SessionHandler *handlers.SessionHandler
	HealthHandler  *handlers.HealthHandler

	// Logger
	Logger logger.Logger
}

// NewContainer cria um novo container de dependências
func NewContainer(db *bun.DB, whatsappManager whatsapp.WhatsAppManager) (*Container, error) {
	c := &Container{
		DB:              db,
		WhatsAppManager: whatsappManager,
		Logger:          logger.WithComponent("di-container"),
	}

	// Inicializar repositórios
	if err := c.initRepositories(); err != nil {
		return nil, err
	}

	// Inicializar use cases
	c.initUseCases()

	// Inicializar handlers
	c.initHandlers()

	c.Logger.Info().Msg("Container initialized successfully")
	return c, nil
}

// initRepositories inicializa os repositórios
func (c *Container) initRepositories() error {
	c.SessionRepo = database.NewSessionRepository(c.DB)
	return nil
}

// initUseCases inicializa os casos de uso
func (c *Container) initUseCases() {
	c.CreateSessionUC = sessionUseCases.NewCreateSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.ListSessionsUC = sessionUseCases.NewListSessionsUseCase(
		c.SessionRepo,
		c.Logger,
	)

	c.GetSessionUC = sessionUseCases.NewGetSessionUseCase(
		c.SessionRepo,
		c.Logger,
	)

	c.DeleteSessionUC = sessionUseCases.NewDeleteSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.ConnectSessionUC = sessionUseCases.NewConnectSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.QRCodeUC = sessionUseCases.NewQRCodeUseCase(
		c.WhatsAppManager,
		c.Logger,
	)

	c.PairPhoneUC = sessionUseCases.NewPairPhoneUseCase(
		c.WhatsAppManager,
		c.Logger,
	)

	c.DisconnectSessionUC = sessionUseCases.NewDisconnectSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SetProxyUC = sessionUseCases.NewSetProxyUseCase(
		c.WhatsAppManager,
		c.Logger,
	)

	c.GetStatusUC = sessionUseCases.NewGetStatusUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)
}

// initHandlers inicializa os handlers
func (c *Container) initHandlers() {
	c.SessionHandler = handlers.NewSessionHandler(
		c.CreateSessionUC,
		c.ListSessionsUC,
		c.GetSessionUC,
		c.DeleteSessionUC,
		c.ConnectSessionUC,
		c.DisconnectSessionUC,
		c.QRCodeUC,
		c.PairPhoneUC,
		c.SetProxyUC,
		c.GetStatusUC,
		c.Logger,
	)

	c.HealthHandler = handlers.NewHealthHandler()
}

// Close encerra o container e todos os seus recursos
func (c *Container) Close() error {
	c.Logger.Info().Msg("Closing container")

	// Fechar WhatsApp Manager se implementar interface de Close
	if closer, ok := c.WhatsAppManager.(interface{ Close() }); ok {
		closer.Close()
	}

	// Fechar banco de dados
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			c.Logger.WithError(err).Error().Msg("Failed to close database")
			return err
		}
	}

	c.Logger.Info().Msg("Container closed successfully")
	return nil
}
