package core

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.mau.fi/whatsmeow/store/sqlstore"

	"zmeow/internal/app/config"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/infra/database"
	"zmeow/internal/infra/whatsapp/connection"
	"zmeow/internal/infra/whatsapp/events"
	"zmeow/internal/infra/whatsapp/services"
	"zmeow/internal/infra/whatsapp/session"
	"zmeow/pkg/logger"
)

// ServiceFactory cria e configura todos os serviços WhatsApp
type ServiceFactory struct {
	db        *bun.DB
	container *sqlstore.Container
	config    *config.Config
	logger    logger.Logger
}

// NewServiceFactory cria uma nova instância da factory
func NewServiceFactory(db *bun.DB, container *sqlstore.Container, cfg *config.Config, log logger.Logger) *ServiceFactory {
	return &ServiceFactory{
		db:        db,
		container: container,
		config:    cfg,
		logger:    log,
	}
}

// CreateServices cria todos os serviços necessários
func (f *ServiceFactory) CreateServices() (*WhatsAppServices, error) {
	// Criar repositórios
	sessionRepo := database.NewSessionRepository(f.db)

	// Criar serviços base
	sessionManager := session.NewSessionManager(f.container, sessionRepo, f.logger)
	qrManager := connection.NewQRCodeManager(f.logger)
	webhookService := services.NewWebhookService(f.logger)
	eventProcessor := events.NewEventProcessor(sessionManager, sessionRepo, webhookService, f.logger)
	connectionManager := connection.NewConnectionManager(sessionManager, qrManager, eventProcessor, f.logger)

	// Iniciar rotina de limpeza do QR Manager
	go qrManager.StartCleanupRoutine(context.Background())

	// Criar serviços de configuração e validação
	configService, err := NewManager(f.db, f.config, f.logger) // Use consolidated manager as config service
	if err != nil {
		return nil, err
	}
	validationService := services.NewValidationService(f.logger)
	securityService := services.NewSecurityService(f.logger)

	return &WhatsAppServices{
		SessionManager:    sessionManager,
		EventProcessor:    eventProcessor,
		QRManager:         qrManager,
		ConnectionManager: connectionManager,
		WebhookService:    webhookService,
		ConfigService:     configService,
		ValidationService: validationService,
		SecurityService:   securityService,
	}, nil
}

// WhatsAppServices agrupa todos os serviços WhatsApp
type WhatsAppServices struct {
	SessionManager    *session.SessionManager
	EventProcessor    *events.EventProcessor
	QRManager         *connection.QRCodeManager
	ConnectionManager *connection.ConnectionManager
	WebhookService    *services.WebhookServiceImpl
	ConfigService     whatsapp.ConfigService
	ValidationService whatsapp.ValidationService
	SecurityService   whatsapp.SecurityService
}

// RefactoredManager é uma versão refatorada do Manager que usa os novos serviços
type RefactoredManager struct {
	services *WhatsAppServices
	logger   logger.Logger
}

// NewRefactoredManager cria um manager refatorado usando os novos serviços
func NewRefactoredManager(db *bun.DB, cfg *config.Config, log logger.Logger) (*RefactoredManager, error) {
	// Criar container SQLStore usando a configuração principal
	dsn := cfg.GetDatabaseDSN()

	container, err := sqlstore.New(context.Background(), "postgres", dsn, nil)
	if err != nil {
		return nil, err
	}

	// Criar factory e serviços
	factory := NewServiceFactory(db, container, cfg, log)
	services, err := factory.CreateServices()
	if err != nil {
		return nil, err
	}

	return &RefactoredManager{
		services: services,
		logger:   log.WithComponent("refactored-whatsapp-manager"),
	}, nil
}

// Implementar interface whatsapp.WhatsAppManager
func (rm *RefactoredManager) RegisterSession(sessionID uuid.UUID) error {
	_, err := rm.services.SessionManager.CreateSession(sessionID)
	return err
}

func (rm *RefactoredManager) ConnectSession(ctx context.Context, sessionID uuid.UUID) error {
	return rm.services.ConnectionManager.Connect(ctx, sessionID)
}

func (rm *RefactoredManager) DisconnectSession(sessionID uuid.UUID) error {
	return rm.services.ConnectionManager.Disconnect(sessionID)
}

func (rm *RefactoredManager) GetQRCode(sessionID uuid.UUID) (string, error) {
	return rm.services.QRManager.GetQRCode(sessionID)
}

func (rm *RefactoredManager) PairPhone(sessionID uuid.UUID, phoneNumber string) (string, error) {
	// Validar número de telefone
	if err := rm.services.ValidationService.ValidatePhoneNumber(phoneNumber); err != nil {
		return "", err
	}

	return rm.services.ConnectionManager.PairPhone(context.Background(), sessionID, phoneNumber)
}

func (rm *RefactoredManager) IsConnected(sessionID uuid.UUID) bool {
	return rm.services.ConnectionManager.IsConnected(sessionID)
}

func (rm *RefactoredManager) SetProxy(sessionID uuid.UUID, proxyURL string) error {
	// TODO: Implementar configuração de proxy
	rm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"proxyUrl":  proxyURL,
	}).Info().Msg("Proxy configuration requested (not implemented)")
	return nil
}

func (rm *RefactoredManager) GetSessionStatus(sessionID uuid.UUID) (string, error) {
	sessionInfo, err := rm.services.SessionManager.GetSession(sessionID)
	if err != nil {
		return "", err
	}
	return sessionInfo.Status, nil
}

func (rm *RefactoredManager) GetSessionJID(sessionID uuid.UUID) (string, error) {
	sessionInfo, err := rm.services.SessionManager.GetSession(sessionID)
	if err != nil {
		return "", err
	}
	if sessionInfo.JID != nil {
		return sessionInfo.JID.String(), nil
	}
	return "", nil
}

func (rm *RefactoredManager) RemoveSession(sessionID uuid.UUID) error {
	return rm.services.SessionManager.RemoveSession(sessionID)
}

func (rm *RefactoredManager) RestoreSession(ctx context.Context, sessionID uuid.UUID, jid string) error {
	// Validar JID
	if err := rm.services.ValidationService.ValidateJID(jid); err != nil {
		return err
	}

	// TODO: Implementar restauração usando SessionManager
	rm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid,
	}).Info().Msg("Session restoration requested (not fully implemented)")
	return nil
}

func (rm *RefactoredManager) GetClient(sessionID uuid.UUID) (whatsapp.WhatsAppClient, error) {
	// TODO: Implementar usando os novos serviços
	return nil, fmt.Errorf("GetClient not implemented in refactored manager")
}

// Close encerra o manager refatorado
func (rm *RefactoredManager) Close() {
	rm.services.SessionManager.Close()
	rm.services.QRManager.Close()
	rm.services.ConnectionManager.Close()
	rm.services.WebhookService.Close()

	rm.logger.Info().Msg("Refactored WhatsApp Manager closed")
}

// GetServices retorna os serviços internos (para testes ou uso avançado)
func (rm *RefactoredManager) GetServices() *WhatsAppServices {
	return rm.services
}
