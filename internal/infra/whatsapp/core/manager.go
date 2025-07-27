package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"zmeow/internal/app/config"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/infra/database"
	"zmeow/internal/infra/whatsapp/connection"
	sessionPkg "zmeow/internal/infra/whatsapp/session"
	"zmeow/pkg/logger"
)

// ============================================================================
// CONSTANTS - Consolidado de constants.go
// Todas as constantes do sistema WhatsApp centralizadas neste arquivo
// ============================================================================

// WhatsApp Configuration Constants
const (
	// Default timeouts
	DefaultConnectionTimeout      = 30 * time.Second
	DefaultQRCodeTimeout          = 30 * time.Second
	DefaultWebhookTimeoutDuration = 30 * time.Second
	DefaultReconnectDelay         = 2 * time.Second

	// QR Code settings
	QRCodeExpirationTime  = 30 * time.Second
	QRCodeCleanupInterval = 10 * time.Second

	// Session settings
	MaxSessionNameLength = 100
	MinSessionNameLength = 3
	MaxRetryAttempts     = 3

	// Webhook settings
	DefaultWebhookRetries        = 3
	DefaultWebhookTimeoutSeconds = 30 // seconds
	MaxWebhookURLLength          = 2048

	// Phone number validation
	MinPhoneNumberLength = 10
	MaxPhoneNumberLength = 15

	// JID validation patterns
	JIDPattern = `^\d+(:?\d+)?@s\.whatsapp\.net$`

	// Session name validation pattern
	SessionNamePattern = `^[a-zA-Z0-9_-]+$`

	// Encryption settings
	DefaultEncryptionKeyLength = 32                                 // AES-256
	DefaultEncryptionKey       = "zmeow-default-encryption-key-32b" // Para desenvolvimento apenas

	// HTTP settings
	DefaultUserAgent          = "ZMeow-Webhook/1.0"
	WebhookValidatorUserAgent = "ZMeow-Webhook-Validator/1.0"
)

// Status constants
const (
	StatusDisconnected  = "disconnected"
	StatusConnecting    = "connecting"
	StatusConnected     = "connected"
	StatusReconnecting  = "reconnecting"
	StatusError         = "error"
	StatusPairing       = "pairing"
	StatusQRCode        = "qr_code"
	StatusAuthenticated = "authenticated"
)

// Event types
const (
	EventConnected    = "connected"
	EventDisconnected = "disconnected"
	EventQRCode       = "qr_code"
	EventMessage      = "message"
	EventError        = "error"
	EventPairSuccess  = "pair_success"
	EventSessionReady = "session_ready"
)

// Component names for logging
const (
	ComponentSessionManager        = "session-manager"
	ComponentEventProcessor        = "event-processor"
	ComponentQRManager             = "qr-manager"
	ComponentConnectionManager     = "connection-manager"
	ComponentWebhookService        = "webhook-service"
	ComponentConfigService         = "config-service"
	ComponentValidationService     = "validation-service"
	ComponentSecurityService       = "security-service"
	ComponentDatabaseUpdater       = "database-updater"
	ComponentUnifiedClient         = "unified-whatsapp-client"
	ComponentManagerAdapter        = "manager-adapter"
	ComponentWhatsAppManager       = "whatsapp-manager"
	ComponentRefactoredManager     = "whatsapp-manager-refactored"
	ComponentDIContainer           = "di-container"
	ComponentRefactoredDIContainer = "refactored-di-container"
	ComponentConfigAdapter         = "config-adapter"
)

// Validation constants
const (
	// Regex patterns
	PhoneNumberPattern = `^\+?[1-9]\d{1,14}$`
	URLPattern         = `^https?://[^\s/$.?#].[^\s]*$`

	// Character sets
	SessionNameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
)

// ============================================================================
// TYPES AND STRUCTURES - Consolidado de manager.go e manager_refactored.go
// Estruturas principais do gerenciador WhatsApp unificado
// ============================================================================

// SessionState representa o estado atual de uma sessão
type SessionState struct {
	ID             uuid.UUID                      `json:"id"`
	JID            *types.JID                     `json:"jid,omitempty"`
	Status         string                         `json:"status"`
	LastSeen       *time.Time                     `json:"lastSeen,omitempty"`
	QRCode         string                         `json:"qrCode,omitempty"`
	ProxyURL       string                         `json:"proxyUrl,omitempty"`
	Webhook        string                         `json:"webhook,omitempty"`
	Client         *whatsmeow.Client              `json:"-"`
	EventHandlerID uint32                         `json:"-"`
	QRChan         <-chan whatsmeow.QRChannelItem `json:"-"`
}

// Manager gerencia múltiplas sessões WhatsApp
// VERSÃO CONSOLIDADA que integra:
// - Manager original (manager.go)
// - ManagerRefactored (manager_refactored.go)
// - ConfigAdapter (config_adapter.go)
// - Constantes do sistema (constants.go)
type Manager struct {
	// Core fields - Gerenciamento de sessões
	db            *bun.DB
	container     *sqlstore.Container
	sessionStates map[uuid.UUID]*SessionState
	mutex         sync.RWMutex
	logger        logger.Logger

	// Config adapter fields - Configurações integradas
	config *config.Config

	// Connection management
	connectionManager *connection.ConnectionManager
}

// ============================================================================
// CONSTRUCTOR
// ============================================================================

// NewManager cria uma nova instância do Manager consolidado
func NewManager(db *bun.DB, cfg *config.Config, log logger.Logger) (*Manager, error) {
	// Criar container SQLStore usando a configuração
	dsn := cfg.GetDatabaseDSN()
	container, err := sqlstore.New(context.Background(), "postgres", dsn, nil)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		db:            db,
		container:     container,
		sessionStates: make(map[uuid.UUID]*SessionState),
		logger:        log.WithComponent(ComponentWhatsAppManager),
		config:        cfg,
	}

	// Inicializar ConnectionManager
	manager.initConnectionManager()

	return manager, nil
}

// initConnectionManager inicializa o ConnectionManager
func (m *Manager) initConnectionManager() {
	// Criar SessionManager wrapper para o ConnectionManager
	sessionManager := &SessionManagerWrapper{manager: m}

	// Criar QRCodeManager
	qrManager := connection.NewQRCodeManager(m.logger)

	// Criar EventProcessor
	eventProcessor := &EventProcessorWrapper{manager: m}

	// Criar ConnectionManager
	m.connectionManager = connection.NewConnectionManager(
		sessionManager,
		qrManager,
		eventProcessor,
		m.logger,
	)
}

// ============================================================================
// CONFIG ADAPTER METHODS (from config_adapter.go)
// ============================================================================

// Get obtém uma configuração por chave das variáveis de ambiente
func (m *Manager) Get(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("configuration key '%s' not found", key)
	}
	return value, nil
}

// Set define uma configuração (apenas para runtime)
func (m *Manager) Set(key, value string) error {
	os.Setenv(key, value)
	m.logger.WithFields(map[string]interface{}{
		"key":   key,
		"value": value,
	}).Debug().Msg("Configuration set")
	return nil
}

// GetDatabaseDSN retorna a DSN do banco de dados usando a configuração principal
func (m *Manager) GetDatabaseDSN() string {
	return m.config.GetDatabaseDSN()
}

// GetWhatsAppConfig retorna configurações do WhatsApp usando a configuração principal
func (m *Manager) GetWhatsAppConfig() *whatsapp.WhatsAppConfig {
	return &whatsapp.WhatsAppConfig{
		DebugLevel:  m.config.WhatsApp.DebugLevel,
		StorePrefix: m.config.WhatsApp.StorePrefix,
		DatabaseDSN: m.config.GetDatabaseDSN(),
	}
}

// ============================================================================
// HELPER FUNCTIONS (from constants.go)
// ============================================================================

// IsValidStatus verifica se um status é válido
func IsValidStatus(status string) bool {
	switch status {
	case StatusDisconnected, StatusConnecting, StatusConnected, StatusReconnecting, StatusError:
		return true
	default:
		return false
	}
}

// ============================================================================
// CORE MANAGER METHODS
// ============================================================================

// CreateSession cria uma nova sessão WhatsApp
func (m *Manager) CreateSession(ctx context.Context, sessionName string) (uuid.UUID, error) {
	if sessionName == "" {
		return uuid.Nil, errors.New("session name cannot be empty")
	}

	sessionID := uuid.New()

	// Criar device store para a sessão
	deviceStore := m.container.NewDevice()
	client := whatsmeow.NewClient(deviceStore, nil) // TODO: Fix logger compatibility

	// Criar estado inicial da sessão
	sessionState := &SessionState{
		ID:     sessionID,
		Status: StatusDisconnected,
		Client: client,
	}

	// Armazenar estado
	m.mutex.Lock()
	m.sessionStates[sessionID] = sessionState
	m.mutex.Unlock()

	// Persistir no banco de dados
	repo := database.NewSessionRepository(m.db)
	dbSession := &session.Session{
		ID:        sessionID,
		Name:      sessionName,
		Status:    StatusDisconnected,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.Create(ctx, dbSession); err != nil {
		// Remover do estado em caso de erro
		m.mutex.Lock()
		delete(m.sessionStates, sessionID)
		m.mutex.Unlock()
		return uuid.Nil, fmt.Errorf("failed to persist session: %w", err)
	}

	m.logger.WithFields(map[string]interface{}{
		"session_id":   sessionID,
		"session_name": sessionName,
	}).Info().Msg("Session created successfully")

	return sessionID, nil
}

// ConnectSession conecta uma sessão específica
func (m *Manager) ConnectSession(ctx context.Context, sessionID uuid.UUID) error {
	m.mutex.RLock()
	_, exists := m.sessionStates[sessionID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Usar ConnectionManager para conectar (inclui lógica de QR code)
	if m.connectionManager != nil {
		return m.connectionManager.Connect(ctx, sessionID)
	}

	// Fallback para método antigo se ConnectionManager não estiver disponível
	m.logger.WithField("session_id", sessionID).Warn().Msg("ConnectionManager not available, using fallback connection")
	return m.connectSessionFallback(ctx, sessionID)
}

// connectSessionFallback método de fallback para conexão direta
func (m *Manager) connectSessionFallback(_ context.Context, sessionID uuid.UUID) error {
	m.mutex.RLock()
	state, exists := m.sessionStates[sessionID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if state.Client == nil {
		return fmt.Errorf("client not initialized for session %s", sessionID)
	}

	// Atualizar status para connecting
	m.updateSessionStatus(sessionID, StatusConnecting)

	// Conectar cliente
	err := state.Client.Connect()
	if err != nil {
		m.updateSessionStatus(sessionID, StatusError)
		return fmt.Errorf("failed to connect session %s: %w", sessionID, err)
	}

	m.updateSessionStatus(sessionID, StatusConnected)

	m.logger.WithField("session_id", sessionID).Debug().Msg("Session connected successfully")
	return nil
}

// DisconnectSessionWithContext desconecta uma sessão específica com contexto
func (m *Manager) DisconnectSessionWithContext(ctx context.Context, sessionID uuid.UUID) error {
	return m.DisconnectSession(sessionID)
}

// DeleteSession remove uma sessão
func (m *Manager) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	// Primeiro desconectar se estiver conectada
	if err := m.DisconnectSession(sessionID); err != nil {
		m.logger.WithError(err).Warn().Msg("Failed to disconnect session before deletion")
	}

	// Remover do estado
	m.mutex.Lock()
	delete(m.sessionStates, sessionID)
	m.mutex.Unlock()

	// Remover do banco de dados
	repo := database.NewSessionRepository(m.db)
	if err := repo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session from database: %w", err)
	}

	m.logger.WithField("session_id", sessionID).Info().Msg("Session deleted successfully")
	return nil
}

// GetSessionStatus retorna o status de uma sessão
func (m *Manager) GetSessionStatus(sessionID uuid.UUID) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.sessionStates[sessionID]
	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	return state.Status, nil
}

// GetAllSessions retorna todas as sessões ativas
func (m *Manager) GetAllSessions() map[uuid.UUID]*SessionState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessions := make(map[uuid.UUID]*SessionState)
	for id, state := range m.sessionStates {
		stateCopy := *state
		sessions[id] = &stateCopy
	}

	return sessions
}

// updateSessionStatus atualiza o status de uma sessão
func (m *Manager) updateSessionStatus(sessionID uuid.UUID, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if state, exists := m.sessionStates[sessionID]; exists {
		state.Status = status
		now := time.Now()
		state.LastSeen = &now
	}
}

// ============================================================================
// WHATSAPP MANAGER INTERFACE IMPLEMENTATION
// ============================================================================

// RegisterSession registra uma nova sessão no manager
func (m *Manager) RegisterSession(sessionID uuid.UUID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.sessionStates[sessionID]; exists {
		return errors.New("session already exists")
	}

	// Criar device store para a sessão
	deviceStore := m.container.NewDevice()
	client := whatsmeow.NewClient(deviceStore, nil)

	// Criar estado inicial da sessão
	sessionState := &SessionState{
		ID:     sessionID,
		Status: StatusDisconnected,
		Client: client,
	}

	// Armazenar estado
	m.sessionStates[sessionID] = sessionState

	m.logger.WithField("session_id", sessionID).Info().Msg("Session registered successfully")
	return nil
}

// DisconnectSession desconecta uma sessão específica (interface WhatsAppManager)
func (m *Manager) DisconnectSession(sessionID uuid.UUID) error {
	m.mutex.RLock()
	state, exists := m.sessionStates[sessionID]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if state.Client != nil {
		state.Client.Disconnect()
	}

	m.updateSessionStatus(sessionID, StatusDisconnected)

	m.logger.WithField("session_id", sessionID).Debug().Msg("Session disconnected successfully")
	return nil
}

// GetQRCode retorna o QR code de uma sessão
func (m *Manager) GetQRCode(sessionID uuid.UUID) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.sessionStates[sessionID]
	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	return state.QRCode, nil
}

// PairPhone realiza pareamento por telefone
func (m *Manager) PairPhone(sessionID uuid.UUID, phoneNumber string) (string, error) {
	m.mutex.RLock()
	state, exists := m.sessionStates[sessionID]
	m.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if state.Client == nil {
		return "", fmt.Errorf("client not initialized for session %s", sessionID)
	}

	// TODO: Implementar pareamento por telefone
	m.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"phone":      phoneNumber,
	}).Info().Msg("Phone pairing requested")

	return "pairing-code-placeholder", nil
}

// IsConnected verifica se uma sessão está conectada
func (m *Manager) IsConnected(sessionID uuid.UUID) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.sessionStates[sessionID]
	if !exists {
		return false
	}

	return state.Client != nil && state.Client.IsConnected()
}

// SetProxy configura proxy para uma sessão
func (m *Manager) SetProxy(sessionID uuid.UUID, proxyURL string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	state, exists := m.sessionStates[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	state.ProxyURL = proxyURL

	m.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"proxy_url":  proxyURL,
	}).Info().Msg("Proxy configured for session")

	return nil
}

// GetSessionJID retorna o JID de uma sessão
func (m *Manager) GetSessionJID(sessionID uuid.UUID) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.sessionStates[sessionID]
	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if state.JID == nil {
		return "", fmt.Errorf("session %s not authenticated", sessionID)
	}

	return state.JID.String(), nil
}

// RemoveSession remove uma sessão
func (m *Manager) RemoveSession(sessionID uuid.UUID) error {
	// Primeiro desconectar se estiver conectada
	if err := m.DisconnectSession(sessionID); err != nil {
		m.logger.WithError(err).Warn().Msg("Failed to disconnect session before removal")
	}

	// Remover do estado
	m.mutex.Lock()
	delete(m.sessionStates, sessionID)
	m.mutex.Unlock()

	m.logger.WithField("session_id", sessionID).Info().Msg("Session removed successfully")
	return nil
}

// RestoreSession restaura uma sessão a partir do WaJID usando device existente
func (m *Manager) RestoreSession(ctx context.Context, sessionID uuid.UUID, waJID string) error {
	// Parse JID
	parsedJID, err := types.ParseJID(waJID)
	if err != nil {
		return fmt.Errorf("invalid WaJID format: %w", err)
	}

	m.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"waJid":      waJID,
	}).Info().Msg("Restoring session from saved WaJID")

	// Tentar obter device existente usando o JID
	deviceStore, err := m.container.GetDevice(ctx, parsedJID)
	if err != nil {
		m.logger.WithError(err).WithField("waJid", waJID).Warn().Msg("Failed to get existing device, creating new one")
		// Se não conseguir obter device existente, criar novo (sessão precisará de nova autenticação)
		deviceStore = m.container.NewDevice()
	}

	// Criar cliente WhatsApp com o device store
	client := whatsmeow.NewClient(deviceStore, nil)

	// Criar estado da sessão
	sessionState := &SessionState{
		ID:     sessionID,
		JID:    &parsedJID,
		Status: StatusDisconnected,
		Client: client,
	}

	// Armazenar estado
	m.mutex.Lock()
	m.sessionStates[sessionID] = sessionState
	m.mutex.Unlock()

	if deviceStore.ID != nil {
		m.logger.WithFields(map[string]interface{}{
			"session_id": sessionID,
			"waJid":      waJID,
			"device_id":  deviceStore.ID.String(),
		}).Debug().Msg("Session restored with existing device")
	} else {
		m.logger.WithFields(map[string]interface{}{
			"session_id": sessionID,
			"waJid":      waJID,
		}).Debug().Msg("Session restored with new device")
	}

	return nil
}

// RestoreSessions restaura todas as sessões autenticadas do banco de dados
func (m *Manager) RestoreSessions(ctx context.Context) error {
	m.logger.Info().Msg("Starting session restoration from database")

	// Buscar apenas sessões que têm WaJID (foram autenticadas)
	repo := database.NewSessionRepository(m.db)
	sessions, err := repo.GetSessionsWithWhatsAppJID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated sessions from database: %w", err)
	}

	if len(sessions) == 0 {
		m.logger.Info().Msg("No authenticated sessions found to restore")
		return nil
	}

	m.logger.WithField("sessions_count", len(sessions)).Info().Msg("Found authenticated sessions to restore")

	restoredCount := 0
	for _, sess := range sessions {
		if err := m.RestoreSession(ctx, sess.ID, sess.WaJID); err != nil {
			m.logger.WithError(err).WithFields(map[string]interface{}{
				"session_id":   sess.ID,
				"session_name": sess.Name,
				"waJid":        sess.WaJID,
			}).Error().Msg("Failed to restore session")
			continue
		}
		restoredCount++

		// Pequena pausa entre restaurações para evitar sobrecarga
		time.Sleep(100 * time.Millisecond)
	}

	m.logger.WithField("restored_count", restoredCount).Info().Msg("Session restoration completed successfully")
	return nil
}

// ConnectRestoredSessions conecta automaticamente todas as sessões restauradas
func (m *Manager) ConnectRestoredSessions(ctx context.Context) {
	m.logger.Debug().Msg("Starting automatic connection of restored sessions")

	// Aguardar um pouco para garantir que todas as sessões foram restauradas
	time.Sleep(5 * time.Second)

	m.mutex.RLock()
	sessions := make(map[uuid.UUID]*SessionState)
	for id, state := range m.sessionStates {
		if state.JID != nil { // Só conectar sessões que têm JID
			sessions[id] = state
		}
	}
	m.mutex.RUnlock()

	if len(sessions) == 0 {
		m.logger.Debug().Msg("No restored sessions found to connect")
		return
	}

	m.logger.WithField("sessions_count", len(sessions)).Info().Msg("Connecting restored sessions")

	connectedCount := 0
	for sessionID, state := range sessions {
		m.logger.WithFields(map[string]interface{}{
			"session_id": sessionID,
			"jid":        state.JID.String(),
		}).Debug().Msg("Connecting restored session")

		// Criar contexto com timeout para cada sessão
		sessionCtx, cancel := context.WithTimeout(ctx, DefaultConnectionTimeout)

		// Executar conexão em goroutine separada para cada sessão
		go func(id uuid.UUID, sessionState *SessionState) {
			defer cancel()

			// Primeiro garantir que a sessão está criada no banco
			repo := database.NewSessionRepository(m.db)
			existingSession, err := repo.GetByID(sessionCtx, id)
			if err != nil {
				m.logger.WithError(err).WithField("session_id", id).Error().Msg("Failed to get session from database")
				return
			}

			// Se a sessão não existe no banco, criar
			if existingSession == nil {
				newSession := &session.Session{
					ID:        id,
					Name:      fmt.Sprintf("restored-%s", id.String()[:8]),
					WaJID:     sessionState.JID.String(),
					Status:    StatusDisconnected,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				if err := repo.Create(sessionCtx, newSession); err != nil {
					m.logger.WithError(err).WithField("session_id", id).Error().Msg("Failed to create restored session in database")
					return
				}

				m.logger.WithField("session_id", id).Info().Msg("Restored session created in database")
			}

			// Agora tentar conectar
			if err := m.ConnectSession(sessionCtx, id); err != nil {
				m.logger.WithError(err).WithField("session_id", id).Error().Msg("Failed to connect restored session")
			} else {
				m.logger.WithField("session_id", id).Debug().Msg("Restored session connected successfully")
			}
		}(sessionID, state)

		connectedCount++

		// Aguardar um pouco entre conexões para evitar sobrecarga
		time.Sleep(2 * time.Second)
	}

	m.logger.WithField("attempted_connections", connectedCount).Info().Msg("Automatic connection of restored sessions completed")
}

// GetClient retorna o cliente WhatsApp de uma sessão específica
func (m *Manager) GetClient(sessionID uuid.UUID) (whatsapp.WhatsAppClient, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.sessionStates[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Retornar um UnifiedClient para a sessão específica
	client := NewUnifiedClientForSession(m, sessionID, m.logger)
	return client, nil
}

// ============================================================================
// WRAPPERS PARA CONNECTIONMANAGER
// ============================================================================

// SessionManagerWrapper adapta o Manager para implementar SessionManagerInterface
type SessionManagerWrapper struct {
	manager *Manager
}

func (smw *SessionManagerWrapper) GetSession(sessionID uuid.UUID) (*sessionPkg.SessionState, error) {
	smw.manager.mutex.RLock()
	defer smw.manager.mutex.RUnlock()

	state, exists := smw.manager.sessionStates[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Converter SessionState do core para sessionPkg.SessionState
	return &sessionPkg.SessionState{
		ID:     state.ID,
		JID:    state.JID,
		Status: state.Status,
		Client: state.Client,
	}, nil
}

func (smw *SessionManagerWrapper) SetQRChannel(sessionID uuid.UUID, qrChan <-chan whatsmeow.QRChannelItem) error {
	smw.manager.mutex.Lock()
	defer smw.manager.mutex.Unlock()

	state, exists := smw.manager.sessionStates[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	state.QRChan = qrChan
	return nil
}

func (smw *SessionManagerWrapper) UpdateSessionJID(sessionID uuid.UUID, jid *types.JID) error {
	smw.manager.mutex.Lock()
	defer smw.manager.mutex.Unlock()

	state, exists := smw.manager.sessionStates[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	state.JID = jid
	return nil
}

func (smw *SessionManagerWrapper) UpdateSessionStatus(sessionID uuid.UUID, status string) error {
	smw.manager.updateSessionStatus(sessionID, status)
	return nil
}

func (smw *SessionManagerWrapper) UpdateSessionQRCode(sessionID uuid.UUID, qrCode string) error {
	smw.manager.mutex.Lock()
	defer smw.manager.mutex.Unlock()

	state, exists := smw.manager.sessionStates[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	state.QRCode = qrCode
	return nil
}

func (smw *SessionManagerWrapper) SetEventHandler(sessionID uuid.UUID, handlerID uint32) error {
	smw.manager.mutex.Lock()
	defer smw.manager.mutex.Unlock()

	state, exists := smw.manager.sessionStates[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	state.EventHandlerID = handlerID
	return nil
}

// EventProcessorWrapper adapta o Manager para implementar EventProcessorInterface
type EventProcessorWrapper struct {
	manager *Manager
}

func (epw *EventProcessorWrapper) ProcessEvent(sessionID uuid.UUID, evt interface{}) {
	eventType := fmt.Sprintf("%T", evt)

	// Log do evento para debugging
	epw.manager.logger.WithFields(map[string]interface{}{
		"sessionId":  sessionID,
		"eventType":  eventType,
		"rawPayload": evt,
	}).Trace().Msg("Raw WhatsApp Event Payload")

	// Processar eventos específicos
	switch e := evt.(type) {
	case *events.Connected:
		epw.handleConnected(sessionID, e)
	case *events.Disconnected:
		epw.handleDisconnected(sessionID, e)
	case *events.LoggedOut:
		epw.handleLoggedOut(sessionID, e)
	case *events.PairSuccess:
		epw.handlePairSuccess(sessionID, e)
	case *events.Message:
		epw.handleMessage(sessionID, e)
	case *events.Receipt:
		// Processar recibo (lógica futura aqui)
	default:
		epw.manager.logger.WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"eventType": eventType,
		}).Debug().Msg("Unhandled WhatsApp event")
	}

	// Log consolidado do evento com mensagens descritivas
	rawData := extractEventData(evt)
	message := getEventMessage(eventType, evt)

	epw.manager.logger.WithFields(map[string]interface{}{
		"eventType": eventType,
		"sessionId": sessionID,
	}).Info().Msgf("%s raw=%s", message, rawData)
}

// handleConnected processa evento de conexão
func (epw *EventProcessorWrapper) handleConnected(sessionID uuid.UUID, _ *events.Connected) {
	epw.manager.mutex.Lock()
	defer epw.manager.mutex.Unlock()

	state, exists := epw.manager.sessionStates[sessionID]
	if !exists {
		epw.manager.logger.WithField("sessionId", sessionID).Error().Msg("Session not found for connected event")
		return
	}

	// Atualizar status na memória
	state.Status = StatusConnected
	now := time.Now()
	state.LastSeen = &now

	// Atualizar JID se disponível
	if state.Client != nil && state.Client.Store.ID != nil {
		state.JID = state.Client.Store.ID
	}

	epw.manager.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       state.JID,
	}).Info().Msg("Session connected")

	// Atualizar banco de dados
	go epw.updateDatabaseOnConnect(sessionID, state.JID)
}

// handleDisconnected processa evento de desconexão
func (epw *EventProcessorWrapper) handleDisconnected(sessionID uuid.UUID, _ *events.Disconnected) {
	epw.manager.mutex.Lock()
	defer epw.manager.mutex.Unlock()

	state, exists := epw.manager.sessionStates[sessionID]
	if !exists {
		epw.manager.logger.WithField("sessionId", sessionID).Error().Msg("Session not found for disconnected event")
		return
	}

	// Atualizar status na memória
	state.Status = StatusDisconnected
	now := time.Now()
	state.LastSeen = &now

	epw.manager.logger.WithField("sessionId", sessionID).Info().Msg("Session disconnected")

	// Atualizar banco de dados
	go epw.updateDatabaseStatus(sessionID, session.WhatsAppStatusDisconnected)
}

// handleLoggedOut processa evento de logout
func (epw *EventProcessorWrapper) handleLoggedOut(sessionID uuid.UUID, _ *events.LoggedOut) {
	epw.manager.mutex.Lock()
	defer epw.manager.mutex.Unlock()

	state, exists := epw.manager.sessionStates[sessionID]
	if !exists {
		epw.manager.logger.WithField("sessionId", sessionID).Error().Msg("Session not found for logged out event")
		return
	}

	// Atualizar status na memória
	state.Status = StatusDisconnected
	state.JID = nil
	now := time.Now()
	state.LastSeen = &now

	epw.manager.logger.WithField("sessionId", sessionID).Warn().Msg("Session logged out")

	// Atualizar banco de dados
	go epw.updateDatabaseOnLogout(sessionID)
}

// handlePairSuccess processa sucesso no pareamento
func (epw *EventProcessorWrapper) handlePairSuccess(sessionID uuid.UUID, evt *events.PairSuccess) {
	epw.manager.mutex.Lock()
	defer epw.manager.mutex.Unlock()

	state, exists := epw.manager.sessionStates[sessionID]
	if !exists {
		epw.manager.logger.WithField("sessionId", sessionID).Error().Msg("Session not found for pair success event")
		return
	}

	// Atualizar JID na memória
	state.JID = &evt.ID

	epw.manager.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       evt.ID.String(),
	}).Info().Msg("Phone pairing successful")

	// Salvar WaJID no banco de dados
	go epw.saveWaJIDToDatabase(sessionID, evt.ID.String())
}

// handleMessage processa mensagens recebidas
func (epw *EventProcessorWrapper) handleMessage(sessionID uuid.UUID, evt *events.Message) {
	epw.manager.mutex.Lock()
	defer epw.manager.mutex.Unlock()

	state, exists := epw.manager.sessionStates[sessionID]
	if !exists {
		epw.manager.logger.WithField("sessionId", sessionID).Error().Msg("Session not found for message event")
		return
	}

	// Atualizar último visto
	now := time.Now()
	state.LastSeen = &now

	epw.manager.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"messageId": evt.Info.ID,
		"from":      evt.Info.Sender.String(),
		"timestamp": evt.Info.Timestamp,
	}).Info().Msg("Message received")

	// Atualizar último visto no banco
	go epw.updateLastSeen(sessionID)
}

// Métodos auxiliares para atualização do banco de dados
func (epw *EventProcessorWrapper) updateDatabaseOnConnect(sessionID uuid.UUID, jid *types.JID) {
	ctx := context.Background()
	repo := database.NewSessionRepository(epw.manager.db)

	// Atualizar status para conectado
	if err := repo.UpdateStatus(ctx, sessionID, session.WhatsAppStatusConnected); err != nil {
		epw.manager.logger.WithError(err).Error().Msg("Failed to update status to connected in database")
	}

	// Atualizar JID se disponível
	if jid != nil {
		if err := repo.UpdateJID(ctx, sessionID, jid.String()); err != nil {
			epw.manager.logger.WithError(err).Error().Msg("Failed to update JID in database")
		}
	}

	epw.manager.logger.WithField("sessionId", sessionID).Debug().Msg("Database updated on connect")
}

func (epw *EventProcessorWrapper) updateDatabaseStatus(sessionID uuid.UUID, status session.WhatsAppSessionStatus) {
	ctx := context.Background()
	repo := database.NewSessionRepository(epw.manager.db)

	if err := repo.UpdateStatus(ctx, sessionID, status); err != nil {
		epw.manager.logger.WithError(err).Error().Msg("Failed to update status in database")
	}

	epw.manager.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"status":    status,
	}).Debug().Msg("Database status updated")
}

func (epw *EventProcessorWrapper) updateDatabaseOnLogout(sessionID uuid.UUID) {
	ctx := context.Background()
	repo := database.NewSessionRepository(epw.manager.db)

	// Atualizar status para desconectado
	if err := repo.UpdateStatus(ctx, sessionID, session.WhatsAppStatusDisconnected); err != nil {
		epw.manager.logger.WithError(err).Error().Msg("Failed to update status to disconnected in database")
	}

	// Limpar JID
	if err := repo.UpdateJID(ctx, sessionID, ""); err != nil {
		epw.manager.logger.WithError(err).Error().Msg("Failed to clear JID in database")
	}

	epw.manager.logger.WithField("sessionId", sessionID).Debug().Msg("Database updated on logout")
}

func (epw *EventProcessorWrapper) updateLastSeen(sessionID uuid.UUID) {
	ctx := context.Background()
	repo := database.NewSessionRepository(epw.manager.db)

	if err := repo.UpdateLastSeen(ctx, sessionID); err != nil {
		epw.manager.logger.WithError(err).Error().Msg("Failed to update last seen in database")
	}
}

// getEventMessage retorna uma mensagem descritiva para cada tipo de evento
func getEventMessage(eventType string, evt interface{}) string {
	switch eventType {
	case "*events.Connected":
		return "Conexão estabelecida com sucesso"
	case "*events.Disconnected":
		return "Conexão encerrada"
	case "*events.Message":
		if msg, ok := evt.(*events.Message); ok {
			if msg.Info.IsFromMe {
				return "Mensagem enviada"
			}
			return "Mensagem recebida"
		}
		return "Evento de mensagem"
	case "*events.Receipt":
		if receipt, ok := evt.(*events.Receipt); ok {
			switch receipt.Type {
			case types.ReceiptTypeDelivered:
				return "Mensagem entregue"
			case types.ReceiptTypeRead:
				return "Mensagem lida"
			case types.ReceiptTypePlayed:
				return "Áudio/vídeo reproduzido"
			default:
				return "Recibo de mensagem"
			}
		}
		return "Recibo de mensagem"
	case "*events.Presence":
		if presence, ok := evt.(*events.Presence); ok {
			switch presence.Unavailable {
			case false:
				return "Usuário online"
			default:
				return "Usuário offline"
			}
		}
		return "Status de presença"
	case "*events.ChatPresence":
		if chatPresence, ok := evt.(*events.ChatPresence); ok {
			switch chatPresence.State {
			case "composing":
				return "Usuário digitando"
			case "recording":
				return "Usuário gravando áudio"
			case "paused":
				return "Usuário parou de digitar"
			default:
				return "Status de digitação"
			}
		}
		return "Status de digitação"
	case "*events.PairSuccess":
		return "Dispositivo pareado com sucesso"
	case "*events.PairError":
		return "Erro no pareamento do dispositivo"
	case "*events.LoggedOut":
		return "Sessão desconectada remotamente"
	case "*events.OfflineSyncPreview":
		if sync, ok := evt.(*events.OfflineSyncPreview); ok {
			return fmt.Sprintf("Sincronização offline iniciada (%d mensagens pendentes)", sync.Total)
		}
		return "Sincronização offline iniciada"
	case "*events.OfflineSyncCompleted":
		if sync, ok := evt.(*events.OfflineSyncCompleted); ok {
			return fmt.Sprintf("Sincronização offline concluída (%d mensagens processadas)", sync.Count)
		}
		return "Sincronização offline concluída"
	case "*events.HistorySync":
		return "Histórico sincronizado"
	case "*events.AppState":
		return "Estado da aplicação atualizado"
	case "*events.KeepAliveTimeout":
		return "Timeout de keep-alive"
	case "*events.KeepAliveRestored":
		return "Keep-alive restaurado"
	case "*events.Blocklist":
		return "Lista de bloqueios atualizada"
	case "*events.Contact":
		return "Contato atualizado"
	case "*events.PushName":
		return "Nome do contato atualizado"
	case "*events.GroupInfo":
		return "Informações do grupo atualizadas"
	case "*events.JoinedGroup":
		return "Adicionado ao grupo"
	case "*events.Newsletter":
		return "Newsletter atualizada"
	case "*events.CallOffer":
		return "Chamada recebida"
	case "*events.CallAccept":
		return "Chamada aceita"
	case "*events.CallPreAccept":
		return "Chamada pré-aceita"
	case "*events.CallTransport":
		return "Transporte de chamada"
	case "*events.CallRelayLatency":
		return "Latência de chamada"
	case "*events.CallTerminate":
		return "Chamada encerrada"
	case "*events.UnknownCallEvent":
		return "Evento de chamada desconhecido"
	case "*events.UndecryptableMessage":
		return "Mensagem não descriptografável"
	case "*events.MediaRetry":
		return "Tentativa de reenvio de mídia"
	case "*events.AppStateSyncComplete":
		return "Sincronização de estado completa"
	case "*events.PictureUpdate":
		return "Foto de perfil atualizada"
	case "*events.IdentityChange":
		return "Identidade alterada"
	case "*events.PrivacySettings":
		return "Configurações de privacidade atualizadas"
	case "*events.TempBan":
		return "Banimento temporário"
	case "*events.ConnectFailure":
		return "Falha na conexão"
	case "*events.ClientOutdated":
		return "Cliente desatualizado"
	case "*events.StreamReplaced":
		return "Stream substituído"
	case "*events.StreamError":
		return "Erro no stream"
	case "*events.QRScanned":
		return "QR Code escaneado"
	case "*events.PairCode":
		return "Código de pareamento gerado"
	default:
		return "Evento do WhatsApp"
	}
}

// extractEventData extrai dados estruturados dos eventos do WhatsApp
func extractEventData(evt interface{}) string {
	// Primeiro tentar serialização JSON padrão
	if eventJSON, err := json.Marshal(evt); err == nil {
		// Se o JSON não estiver vazio, retornar
		jsonStr := string(eventJSON)
		if jsonStr != "{}" && jsonStr != "null" {
			return jsonStr
		}
	}

	// Para eventos vazios, adicionar informações contextuais
	eventType := fmt.Sprintf("%T", evt)
	switch eventType {
	case "*events.Connected":
		return `{"status":"connected","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
	case "*events.Disconnected":
		return `{"status":"disconnected","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
	default:
		// Para outros eventos vazios, usar reflection
		return extractEventDataWithReflection(evt)
	}
}

// extractEventDataWithReflection usa reflection para extrair dados de eventos
func extractEventDataWithReflection(evt interface{}) string {
	if evt == nil {
		return "{}"
	}

	// Usar reflection para extrair campos
	v := reflect.ValueOf(evt)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "{}"
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// Se não for struct, tentar converter para string
		return fmt.Sprintf(`{"value": %v}`, evt)
	}

	// Extrair campos da struct
	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Pular campos não exportados
		if !field.CanInterface() {
			continue
		}

		// Obter nome do campo (usar tag json se disponível)
		fieldName := fieldType.Name
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			if commaIdx := strings.Index(jsonTag, ","); commaIdx > 0 {
				fieldName = jsonTag[:commaIdx]
			} else {
				fieldName = jsonTag
			}
		}

		// Obter valor do campo
		fieldValue := field.Interface()

		// Processar valores especiais
		switch v := fieldValue.(type) {
		case time.Time:
			if !v.IsZero() {
				result[fieldName] = v.Format(time.RFC3339)
			}
		case *time.Time:
			if v != nil && !v.IsZero() {
				result[fieldName] = v.Format(time.RFC3339)
			}
		default:
			// Verificar se o valor não é zero
			if !reflect.ValueOf(fieldValue).IsZero() {
				result[fieldName] = fieldValue
			}
		}
	}

	// Se não encontrou nenhum campo, retornar objeto vazio
	if len(result) == 0 {
		return "{}"
	}

	// Serializar resultado
	if resultJSON, err := json.Marshal(result); err == nil {
		return string(resultJSON)
	}

	return "{}"
}

// saveWaJIDToDatabase salva o WaJID no banco de dados após autenticação bem-sucedida
func (epw *EventProcessorWrapper) saveWaJIDToDatabase(sessionID uuid.UUID, waJID string) {
	ctx := context.Background()

	// Obter repository
	repo := database.NewSessionRepository(epw.manager.db)

	// Salvar WaJID
	err := repo.UpdateWhatsAppJID(ctx, sessionID, waJID)
	if err != nil {
		epw.manager.logger.WithError(err).WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"waJid":     waJID,
		}).Error().Msg("Failed to save WaJID to database")
		return
	}

	// Atualizar status para connected
	err = repo.UpdateStatus(ctx, sessionID, session.WhatsAppStatusConnected)
	if err != nil {
		epw.manager.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("Failed to update session status to connected")
	}

	epw.manager.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"waJid":     waJID,
	}).Info().Msg("WaJID saved to database and status updated to connected")
}

// Close encerra o manager e todos os seus recursos
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.logger.Info().Msg("Closing WhatsApp Manager")

	// Fechar todas as sessões ativas
	for sessionID, state := range m.sessionStates {
		if state.Client != nil {
			state.Client.Disconnect()
		}
		delete(m.sessionStates, sessionID)
	}

	// Fechar container
	if m.container != nil {
		m.container.Close()
	}

	m.logger.Info().Msg("WhatsApp Manager closed successfully")
	return nil
}
