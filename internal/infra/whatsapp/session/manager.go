package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"

	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// ============================================================================
// TYPES AND STRUCTURES - Consolidado de session_manager.go e database_updater.go
// Estruturas para gerenciamento completo de sessões WhatsApp
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

// SessionManager gerencia o ciclo de vida das sessões WhatsApp
// VERSÃO CONSOLIDADA que integra:
// - SessionManager original (services/session_manager.go)
// - DatabaseUpdater (database_updater.go)
type SessionManager struct {
	// Session management fields - Gerenciamento em memória
	sessions  map[uuid.UUID]*SessionState
	container *sqlstore.Container
	mutex     sync.RWMutex
	logger    logger.Logger

	// Database operations fields - Persistência no banco
	repo session.SessionRepository
}

// ============================================================================
// DATABASE UPDATE TYPES (from database_updater.go)
// ============================================================================

// UpdateType representa o tipo de atualização
type UpdateType string

const (
	UpdateTypeStatus     UpdateType = "status"
	UpdateTypeJID        UpdateType = "jid"
	UpdateTypeLastSeen   UpdateType = "last_seen"
	UpdateTypeConnect    UpdateType = "connect"
	UpdateTypeDisconnect UpdateType = "disconnect"
	UpdateTypeLogout     UpdateType = "logout"
)

// SessionUpdate representa uma atualização de sessão
type SessionUpdate struct {
	SessionID uuid.UUID
	Type      UpdateType
	Status    *session.WhatsAppSessionStatus
	JID       *string
}

// ============================================================================
// CONSTRUCTOR
// ============================================================================

// NewSessionManager cria uma nova instância do SessionManager consolidado
func NewSessionManager(container *sqlstore.Container, repo session.SessionRepository, log logger.Logger) *SessionManager {
	return &SessionManager{
		sessions:  make(map[uuid.UUID]*SessionState),
		container: container,
		repo:      repo,
		logger:    log.WithComponent("session-manager"),
	}
}

// GetRepository retorna o repository de sessões
func (sm *SessionManager) GetRepository() session.SessionRepository {
	return sm.repo
}

// GetContainer retorna o container SQLStore
func (sm *SessionManager) GetContainer() *sqlstore.Container {
	return sm.container
}

// ============================================================================
// SESSION MANAGEMENT METHODS (from session_manager.go)
// ============================================================================

// CreateSession cria uma nova sessão
func (sm *SessionManager) CreateSession(sessionID uuid.UUID) (*SessionState, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		return nil, whatsapp.ErrSessionAlreadyExists
	}

	// Criar novo device store
	deviceStore := sm.container.NewDevice()

	// Criar cliente WhatsApp
	client := whatsmeow.NewClient(deviceStore, nil)

	// Criar estado da sessão
	state := &SessionState{
		ID:     sessionID,
		Status: "disconnected",
		Client: client,
	}

	sm.sessions[sessionID] = state

	sm.logger.WithField("sessionId", sessionID).Info().Msg("Session created successfully")
	return state, nil
}

// GetSession retorna uma sessão pelo ID
func (sm *SessionManager) GetSession(sessionID uuid.UUID) (*SessionState, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return nil, whatsapp.ErrSessionNotFound
	}

	return state, nil
}

// UpdateSessionStatus atualiza o status de uma sessão
func (sm *SessionManager) UpdateSessionStatus(sessionID uuid.UUID, status string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return whatsapp.ErrSessionNotFound
	}

	state.Status = status
	now := time.Now()
	state.LastSeen = &now

	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"status":    status,
	}).Debug().Msg("Session status updated")

	return nil
}

// RemoveSession remove uma sessão
func (sm *SessionManager) RemoveSession(sessionID uuid.UUID) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return whatsapp.ErrSessionNotFound
	}

	// Desconectar se estiver conectado
	if state.Client != nil && state.Client.IsConnected() {
		state.Client.Disconnect()
	}

	// Remover event handler se existir
	if state.Client != nil && state.EventHandlerID != 0 {
		state.Client.RemoveEventHandler(state.EventHandlerID)
	}

	delete(sm.sessions, sessionID)

	sm.logger.WithField("sessionId", sessionID).Info().Msg("Session removed")
	return nil
}

// IsConnected verifica se uma sessão está conectada
func (sm *SessionManager) IsConnected(sessionID uuid.UUID) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return false
	}

	return state.Client != nil && state.Client.IsConnected()
}

// GetAllSessions retorna todas as sessões
func (sm *SessionManager) GetAllSessions() map[uuid.UUID]*SessionState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessions := make(map[uuid.UUID]*SessionState)
	for id, state := range sm.sessions {
		stateCopy := *state
		sessions[id] = &stateCopy
	}

	return sessions
}

// RestoreSession restaura uma sessão a partir do WaJID usando device existente
func (sm *SessionManager) RestoreSession(sessionID uuid.UUID, waJID string) (*SessionState, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		return nil, whatsapp.ErrSessionAlreadyExists
	}

	// Parse JID
	parsedJID, err := types.ParseJID(waJID)
	if err != nil {
		return nil, fmt.Errorf("invalid WaJID format: %w", err)
	}

	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"waJid":     waJID,
	}).Info().Msg("Restoring session from WaJID")

	// Tentar obter device existente usando o JID
	deviceStore, err := sm.container.GetDevice(context.Background(), parsedJID)
	if err != nil {
		sm.logger.WithError(err).WithField("waJid", waJID).Warn().Msg("Failed to get existing device, creating new one")
		// Se não conseguir obter device existente, criar novo (sessão precisará de nova autenticação)
		deviceStore = sm.container.NewDevice()
	}

	// Criar cliente WhatsApp
	client := whatsmeow.NewClient(deviceStore, nil)

	// Criar estado da sessão
	state := &SessionState{
		ID:     sessionID,
		JID:    &parsedJID,
		Status: "disconnected",
		Client: client,
	}

	sm.sessions[sessionID] = state

	if deviceStore.ID != nil {
		sm.logger.WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"waJid":     waJID,
			"deviceId":  deviceStore.ID.String(),
		}).Info().Msg("Session restored successfully with existing device")
	} else {
		sm.logger.WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"waJid":     waJID,
		}).Info().Msg("Session restored with new device - will need authentication")
	}

	return state, nil
}

// RestoreSessionWithDevice restaura uma sessão usando um device store existente
func (sm *SessionManager) RestoreSessionWithDevice(sessionID uuid.UUID, jid types.JID, deviceStore *store.Device) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		return whatsapp.ErrSessionAlreadyExists
	}

	// Criar cliente WhatsApp com o device store existente
	client := whatsmeow.NewClient(deviceStore, nil)

	// Criar estado da sessão
	state := &SessionState{
		ID:     sessionID,
		JID:    &jid,
		Status: "disconnected",
		Client: client,
	}

	sm.sessions[sessionID] = state

	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid.String(),
	}).Info().Msg("Session restored with existing device")

	return nil
}

// SetEventHandler define o event handler para uma sessão
func (sm *SessionManager) SetEventHandler(sessionID uuid.UUID, handlerID uint32) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return whatsapp.ErrSessionNotFound
	}

	state.EventHandlerID = handlerID
	return nil
}

// SetQRChannel define o canal de QR code para uma sessão
func (sm *SessionManager) SetQRChannel(sessionID uuid.UUID, qrChan <-chan whatsmeow.QRChannelItem) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return whatsapp.ErrSessionNotFound
	}

	state.QRChan = qrChan
	return nil
}

// ============================================================================
// DATABASE OPERATIONS METHODS (from database_updater.go)
// ============================================================================

// UpdateSessionStatusDB atualiza o status de uma sessão no banco de dados
func (sm *SessionManager) UpdateSessionStatusDB(ctx context.Context, sessionID uuid.UUID, status session.WhatsAppSessionStatus) error {
	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"status":    status,
	}).Debug().Msg("Updating session status in database")

	return sm.repo.UpdateStatus(ctx, sessionID, status)
}

// UpdateSessionQRCode atualiza o QR code de uma sessão
func (sm *SessionManager) UpdateSessionQRCode(sessionID uuid.UUID, qrCode string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return whatsapp.ErrSessionNotFound
	}

	state.QRCode = qrCode

	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"qrCode":    qrCode != "",
	}).Debug().Msg("Session QR code updated")

	return nil
}

// UpdateSessionJID atualiza o JID de uma sessão em memória (versão com types.JID)
func (sm *SessionManager) UpdateSessionJID(sessionID uuid.UUID, jid *types.JID) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	state, exists := sm.sessions[sessionID]
	if !exists {
		return whatsapp.ErrSessionNotFound
	}

	state.JID = jid

	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid,
	}).Debug().Msg("Session JID updated in memory")

	return nil
}

// UpdateSessionJIDDB atualiza o JID de uma sessão no banco de dados
func (sm *SessionManager) UpdateSessionJIDDB(ctx context.Context, sessionID uuid.UUID, jid string) error {
	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid,
	}).Debug().Msg("Updating session JID in database")

	return sm.repo.UpdateJID(ctx, sessionID, jid)
}

// UpdateLastSeen atualiza o último visto de uma sessão
func (sm *SessionManager) UpdateLastSeen(ctx context.Context, sessionID uuid.UUID) error {
	sm.logger.WithField("sessionId", sessionID).Debug().Msg("Updating session last seen")

	// Buscar a sessão atual
	sess, err := sm.repo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Atualizar last seen
	now := time.Now()
	sess.LastSeen = &now
	sess.UpdatedAt = now

	return sm.repo.Update(ctx, sess)
}

// UpdateSessionOnConnect atualiza sessão quando conecta (status + JID + last seen)
func (sm *SessionManager) UpdateSessionOnConnect(ctx context.Context, sessionID uuid.UUID, jid string) error {
	sm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid,
	}).Debug().Msg("Updating session on connect")

	// Atualizar status para conectado
	if err := sm.UpdateSessionStatusDB(ctx, sessionID, session.WhatsAppStatusConnected); err != nil {
		return err
	}

	// Atualizar JID se fornecido
	if jid != "" {
		if err := sm.UpdateSessionJIDDB(ctx, sessionID, jid); err != nil {
			return err
		}
	}

	// Atualizar último visto
	if err := sm.UpdateLastSeen(ctx, sessionID); err != nil {
		return err
	}

	// Atualizar estado em memória também
	sm.mutex.Lock()
	if state, exists := sm.sessions[sessionID]; exists {
		state.Status = string(session.WhatsAppStatusConnected)
		if jid != "" {
			parsedJID, err := types.ParseJID(jid)
			if err == nil {
				state.JID = &parsedJID
			}
		}
		now := time.Now()
		state.LastSeen = &now
	}
	sm.mutex.Unlock()

	return nil
}

// UpdateSessionOnDisconnect atualiza sessão quando desconecta
func (sm *SessionManager) UpdateSessionOnDisconnect(ctx context.Context, sessionID uuid.UUID) error {
	sm.logger.WithField("sessionId", sessionID).Debug().Msg("Updating session on disconnect")

	// Atualizar status para desconectado
	if err := sm.UpdateSessionStatusDB(ctx, sessionID, session.WhatsAppStatusDisconnected); err != nil {
		return err
	}

	// Atualizar último visto
	if err := sm.UpdateLastSeen(ctx, sessionID); err != nil {
		return err
	}

	// Atualizar estado em memória também
	sm.mutex.Lock()
	if state, exists := sm.sessions[sessionID]; exists {
		state.Status = string(session.WhatsAppStatusDisconnected)
		now := time.Now()
		state.LastSeen = &now
	}
	sm.mutex.Unlock()

	return nil
}

// UpdateSessionOnLogout atualiza sessão quando faz logout
func (sm *SessionManager) UpdateSessionOnLogout(ctx context.Context, sessionID uuid.UUID) error {
	sm.logger.WithField("sessionId", sessionID).Debug().Msg("Updating session on logout")

	// Atualizar status para desconectado
	if err := sm.UpdateSessionStatusDB(ctx, sessionID, session.WhatsAppStatusDisconnected); err != nil {
		return err
	}

	// Limpar JID
	if err := sm.UpdateSessionJIDDB(ctx, sessionID, ""); err != nil {
		return err
	}

	// Atualizar último visto
	if err := sm.UpdateLastSeen(ctx, sessionID); err != nil {
		return err
	}

	// Atualizar estado em memória também
	sm.mutex.Lock()
	if state, exists := sm.sessions[sessionID]; exists {
		state.Status = string(session.WhatsAppStatusDisconnected)
		state.JID = nil
		now := time.Now()
		state.LastSeen = &now
	}
	sm.mutex.Unlock()

	return nil
}

// BatchUpdate executa múltiplas atualizações em lote
func (sm *SessionManager) BatchUpdate(ctx context.Context, updates []SessionUpdate) error {
	sm.logger.WithField("count", len(updates)).Debug().Msg("Executing batch updates")

	for _, update := range updates {
		if err := sm.performSingleUpdate(ctx, update); err != nil {
			sm.logger.WithError(err).WithField("sessionId", update.SessionID).Error().Msg("Failed to execute update")
			return err
		}
	}

	return nil
}

// performSingleUpdate executa uma única atualização
func (sm *SessionManager) performSingleUpdate(ctx context.Context, update SessionUpdate) error {
	switch update.Type {
	case UpdateTypeStatus:
		if update.Status != nil {
			return sm.UpdateSessionStatusDB(ctx, update.SessionID, *update.Status)
		}
	case UpdateTypeJID:
		if update.JID != nil {
			return sm.UpdateSessionJIDDB(ctx, update.SessionID, *update.JID)
		}
	case UpdateTypeLastSeen:
		return sm.UpdateLastSeen(ctx, update.SessionID)
	case UpdateTypeConnect:
		jid := ""
		if update.JID != nil {
			jid = *update.JID
		}
		return sm.UpdateSessionOnConnect(ctx, update.SessionID, jid)
	case UpdateTypeDisconnect:
		return sm.UpdateSessionOnDisconnect(ctx, update.SessionID)
	case UpdateTypeLogout:
		return sm.UpdateSessionOnLogout(ctx, update.SessionID)
	default:
		return fmt.Errorf("unknown update type: %s", update.Type)
	}
	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// CreateStatusUpdate cria um SessionUpdate para status
func CreateStatusUpdate(sessionID uuid.UUID, status session.WhatsAppSessionStatus) SessionUpdate {
	return SessionUpdate{
		SessionID: sessionID,
		Type:      UpdateTypeStatus,
		Status:    &status,
	}
}

// CreateJIDUpdate cria um SessionUpdate para JID
func CreateJIDUpdate(sessionID uuid.UUID, jid string) SessionUpdate {
	return SessionUpdate{
		SessionID: sessionID,
		Type:      UpdateTypeJID,
		JID:       &jid,
	}
}

// CreateConnectUpdate cria um SessionUpdate para conexão
func CreateConnectUpdate(sessionID uuid.UUID, jid string) SessionUpdate {
	return SessionUpdate{
		SessionID: sessionID,
		Type:      UpdateTypeConnect,
		JID:       &jid,
	}
}

// Close encerra o session manager e limpa recursos
func (sm *SessionManager) Close() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.logger.Info().Msg("Closing Session Manager")

	// Fechar todas as sessões ativas
	for sessionID, state := range sm.sessions {
		if state.Client != nil && state.Client.IsConnected() {
			state.Client.Disconnect()
		}
		// Remover event handler se existir
		if state.Client != nil && state.EventHandlerID != 0 {
			state.Client.RemoveEventHandler(state.EventHandlerID)
		}
		delete(sm.sessions, sessionID)
	}

	sm.logger.Info().Msg("Session Manager closed successfully")
	return nil
}
