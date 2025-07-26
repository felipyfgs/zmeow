package connection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"

	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/infra/whatsapp/session"
	"zmeow/pkg/logger"
)

// SessionManagerInterface define os métodos necessários do session manager
type SessionManagerInterface interface {
	GetSession(sessionID uuid.UUID) (*session.SessionState, error)
	SetQRChannel(sessionID uuid.UUID, qrChan <-chan whatsmeow.QRChannelItem) error
	UpdateSessionJID(sessionID uuid.UUID, jid *types.JID) error
	UpdateSessionStatus(sessionID uuid.UUID, status string) error
	UpdateSessionQRCode(sessionID uuid.UUID, qrCode string) error
	SetEventHandler(sessionID uuid.UUID, handlerID uint32) error
}

// EventProcessorInterface define os métodos necessários do event processor
type EventProcessorInterface interface {
	ProcessEvent(sessionID uuid.UUID, evt interface{})
}

// ConnectionStatus representa o status de uma conexão
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusReconnecting ConnectionStatus = "reconnecting"
	StatusError        ConnectionStatus = "error"
)

// ConnectionInfo contém informações sobre uma conexão
type ConnectionInfo struct {
	SessionID       uuid.UUID        `json:"sessionId"`
	Status          ConnectionStatus `json:"status"`
	ConnectedAt     *time.Time       `json:"connectedAt,omitempty"`
	LastSeen        *time.Time       `json:"lastSeen,omitempty"`
	RetryCount      int              `json:"retryCount"`
	LastError       string           `json:"lastError,omitempty"`
	IsAuthenticated bool             `json:"isAuthenticated"`
}

// ConnectionManager gerencia conexões WhatsApp
type ConnectionManager struct {
	sessionManager SessionManagerInterface
	qrManager      *QRCodeManager
	eventProcessor EventProcessorInterface
	connections    map[uuid.UUID]*ConnectionInfo
	mutex          sync.RWMutex
	logger         logger.Logger
}

// NewConnectionManager cria uma nova instância do ConnectionManager
func NewConnectionManager(
	sessionManager SessionManagerInterface,
	qrManager *QRCodeManager,
	eventProcessor EventProcessorInterface,
	log logger.Logger,
) *ConnectionManager {
	return &ConnectionManager{
		sessionManager: sessionManager,
		qrManager:      qrManager,
		eventProcessor: eventProcessor,
		connections:    make(map[uuid.UUID]*ConnectionInfo),
		logger:         log.WithComponent("connection-manager"),
	}
}

// ConnectWithRetry conecta uma sessão ao WhatsApp com retry automático
func (cm *ConnectionManager) ConnectWithRetry(ctx context.Context, sessionID uuid.UUID) error {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		cm.logger.WithFields(map[string]interface{}{
			"sessionId":  sessionID,
			"attempt":    attempt,
			"maxRetries": maxRetries,
		}).Info().Msg("Attempting connection")

		err := cm.Connect(ctx, sessionID)
		if err == nil {
			cm.logger.WithField("sessionId", sessionID).Info().Msg("Connection successful")
			return nil
		}

		lastErr = err
		cm.logger.WithError(err).WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"attempt":   attempt,
		}).Warn().Msg("Connection attempt failed")

		// Se não é o último attempt, aguardar antes de tentar novamente
		if attempt < maxRetries {
			cm.logger.WithFields(map[string]interface{}{
				"sessionId": sessionID,
				"delay":     retryDelay,
			}).Info().Msg("Waiting before retry")

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continuar para próximo attempt
			}
		}
	}

	cm.logger.WithError(lastErr).WithField("sessionId", sessionID).Error().Msg("All connection attempts failed")
	return fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, lastErr)
}

// Connect conecta uma sessão ao WhatsApp
func (cm *ConnectionManager) Connect(ctx context.Context, sessionID uuid.UUID) error {
	cm.logger.WithField("sessionId", sessionID).Info().Msg("Starting connection")

	// Obter sessão
	state, err := cm.sessionManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	if state.Client == nil {
		return fmt.Errorf("client is nil for session %s", sessionID)
	}

	// Verificar se já está conectado
	if state.Client.IsConnected() {
		cm.logger.WithField("sessionId", sessionID).Warn().Msg("Session already connected")
		return whatsapp.ErrSessionAlreadyConnected
	}

	// Atualizar status de conexão
	cm.updateConnectionStatus(sessionID, StatusConnecting)

	// Verificar se precisa de autenticação baseado no WaJID salvo
	needsAuthentication, err := cm.needsAuthentication(ctx, sessionID)
	if err != nil {
		cm.updateConnectionStatus(sessionID, StatusError)
		return fmt.Errorf("failed to check authentication status: %w", err)
	}

	if needsAuthentication {
		// Nova sessão - precisa de QR code ou pareamento
		return cm.connectNewSession(ctx, sessionID, state)
	} else {
		// Sessão existente - conectar diretamente
		return cm.connectExistingSession(ctx, sessionID, state)
	}
}

// connectNewSession conecta uma nova sessão que precisa de autenticação via QR code
func (cm *ConnectionManager) connectNewSession(ctx context.Context, sessionID uuid.UUID, state *session.SessionState) error {
	cm.logger.WithField("sessionId", sessionID).Info().Msg("Connecting new session - QR authentication required")

	// Verificar se o cliente está pronto para autenticação
	if state.Client.Store.ID != nil {
		cm.logger.WithField("sessionId", sessionID).Warn().Msg("Client already has device ID, should use existing session flow")
		return cm.connectExistingSession(ctx, sessionID, state)
	}

	// Gerar QR code para nova autenticação
	qrChan, err := cm.qrManager.GenerateQRCode(ctx, sessionID, state.Client)
	if err != nil {
		cm.updateConnectionStatus(sessionID, StatusError)
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Definir canal QR na sessão
	if err := cm.sessionManager.SetQRChannel(sessionID, qrChan); err != nil {
		cm.updateConnectionStatus(sessionID, StatusError)
		return err
	}

	// Registrar event handler para processar eventos WhatsApp
	handlerID := state.Client.AddEventHandler(func(evt interface{}) {
		cm.eventProcessor.ProcessEvent(sessionID, evt)
	})

	// Armazenar handler ID na sessão
	if err := cm.sessionManager.SetEventHandler(sessionID, handlerID); err != nil {
		cm.logger.WithError(err).Warn().Msg("Failed to store event handler ID")
	}

	// Conectar cliente para iniciar processo de autenticação
	if err := state.Client.Connect(); err != nil {
		cm.updateConnectionStatus(sessionID, StatusError)
		return fmt.Errorf("failed to connect client for authentication: %w", err)
	}

	// Iniciar timeout para conexão
	go cm.startConnectionTimeout(ctx, sessionID)

	cm.logger.WithField("sessionId", sessionID).Info().Msg("New session connection initiated, waiting for QR scan")
	return nil
}

// connectExistingSession conecta uma sessão existente que já foi autenticada
func (cm *ConnectionManager) connectExistingSession(ctx context.Context, sessionID uuid.UUID, state *session.SessionState) error {
	cm.logger.WithField("sessionId", sessionID).Info().Msg("Connecting existing authenticated session")

	// Verificar se o cliente tem device ID
	if state.Client.Store.ID == nil {
		cm.logger.WithField("sessionId", sessionID).Error().Msg("Existing session has no device ID - this should not happen")
		return fmt.Errorf("existing session missing device ID")
	}

	// Registrar event handler para processar eventos WhatsApp
	handlerID := state.Client.AddEventHandler(func(evt interface{}) {
		cm.eventProcessor.ProcessEvent(sessionID, evt)
	})

	// Armazenar handler ID na sessão
	if err := cm.sessionManager.SetEventHandler(sessionID, handlerID); err != nil {
		cm.logger.WithError(err).Warn().Msg("Failed to store event handler ID")
	}

	// Conectar cliente diretamente (sem QR code)
	if err := state.Client.Connect(); err != nil {
		cm.updateConnectionStatus(sessionID, StatusError)
		return fmt.Errorf("failed to connect existing session: %w", err)
	}

	// Atualizar status para conectado
	cm.updateConnectionStatus(sessionID, StatusConnected)

	// Salvar WaJID no banco se ainda não estiver salvo
	if state.Client.Store.ID != nil {
		sessionManager, ok := cm.sessionManager.(*session.SessionManager)
		if ok {
			repo := sessionManager.GetRepository()
			if err := repo.UpdateWhatsAppJID(ctx, sessionID, state.Client.Store.ID.String()); err != nil {
				cm.logger.WithError(err).Warn().Msg("Failed to update WaJID in database")
			}
		}
	}

	cm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"deviceId":  state.Client.Store.ID.String(),
	}).Info().Msg("Existing session connected successfully")
	return nil
}

// Disconnect desconecta uma sessão
func (cm *ConnectionManager) Disconnect(sessionID uuid.UUID) error {
	cm.logger.WithField("sessionId", sessionID).Info().Msg("Disconnecting session")

	// Obter sessão
	state, err := cm.sessionManager.GetSession(sessionID)
	if err != nil {
		return err
	}

	if state.Client == nil {
		return fmt.Errorf("client is nil for session %s", sessionID)
	}

	// Verificar se está conectado
	if !state.Client.IsConnected() {
		cm.logger.WithField("sessionId", sessionID).Warn().Msg("Session already disconnected")
		return whatsapp.ErrSessionNotConnected
	}

	// Desconectar cliente
	state.Client.Disconnect()

	// Atualizar status
	cm.updateConnectionStatus(sessionID, StatusDisconnected)

	// Limpar QR code se existir
	cm.qrManager.ClearQRCode(sessionID)

	cm.logger.WithField("sessionId", sessionID).Info().Msg("Session disconnected successfully")
	return nil
}

// Reconnect reconecta uma sessão
func (cm *ConnectionManager) Reconnect(ctx context.Context, sessionID uuid.UUID) error {
	cm.logger.WithField("sessionId", sessionID).Info().Msg("Reconnecting session")

	// Incrementar contador de retry
	cm.incrementRetryCount(sessionID)

	// Atualizar status
	cm.updateConnectionStatus(sessionID, StatusReconnecting)

	// Tentar desconectar primeiro (se estiver conectado)
	if err := cm.Disconnect(sessionID); err != nil && err != whatsapp.ErrSessionNotConnected {
		cm.logger.WithError(err).Warn().Msg("Failed to disconnect before reconnect")
	}

	// Aguardar um pouco antes de reconectar
	time.Sleep(2 * time.Second)

	// Conectar novamente
	return cm.Connect(ctx, sessionID)
}

// PairPhone realiza pareamento via telefone
func (cm *ConnectionManager) PairPhone(ctx context.Context, sessionID uuid.UUID, phoneNumber string) (string, error) {
	cm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     phoneNumber,
	}).Info().Msg("Starting phone pairing")

	// Obter sessão
	state, err := cm.sessionManager.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	if state.Client == nil {
		return "", fmt.Errorf("client is nil for session %s", sessionID)
	}

	// Verificar se já está autenticado
	if state.Client.Store.ID != nil {
		return "", whatsapp.ErrSessionAlreadyAuthenticated
	}

	// Realizar pareamento
	code, err := state.Client.PairPhone(ctx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		cm.updateConnectionStatus(sessionID, StatusError)
		return "", fmt.Errorf("failed to pair phone: %w", err)
	}

	cm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     phoneNumber,
		"code":      code,
	}).Info().Msg("Phone pairing code generated")

	return code, nil
}

// IsConnected verifica se uma sessão está conectada
func (cm *ConnectionManager) IsConnected(sessionID uuid.UUID) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connInfo, exists := cm.connections[sessionID]
	if !exists {
		return false
	}

	return connInfo.Status == StatusConnected
}

// GetConnectionInfo retorna informações de conexão de uma sessão
func (cm *ConnectionManager) GetConnectionInfo(sessionID uuid.UUID) (*ConnectionInfo, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connInfo, exists := cm.connections[sessionID]
	if !exists {
		return nil, fmt.Errorf("connection info not found for session %s", sessionID)
	}

	// Retornar cópia
	infoCopy := *connInfo
	return &infoCopy, nil
}

// GetAllConnections retorna informações de todas as conexões
func (cm *ConnectionManager) GetAllConnections() map[uuid.UUID]*ConnectionInfo {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[uuid.UUID]*ConnectionInfo)
	for sessionID, connInfo := range cm.connections {
		infoCopy := *connInfo
		result[sessionID] = &infoCopy
	}

	return result
}

// updateConnectionStatus atualiza o status de uma conexão
func (cm *ConnectionManager) updateConnectionStatus(sessionID uuid.UUID, status ConnectionStatus) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connInfo, exists := cm.connections[sessionID]
	if !exists {
		connInfo = &ConnectionInfo{
			SessionID: sessionID,
		}
		cm.connections[sessionID] = connInfo
	}

	connInfo.Status = status
	now := time.Now()
	connInfo.LastSeen = &now

	switch status {
	case StatusConnected:
		connInfo.ConnectedAt = &now
		connInfo.IsAuthenticated = true
		connInfo.LastError = ""
	case StatusError:
		connInfo.IsAuthenticated = false
	}

	cm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"status":    status,
	}).Debug().Msg("Connection status updated")
}

// incrementRetryCount incrementa o contador de tentativas
func (cm *ConnectionManager) incrementRetryCount(sessionID uuid.UUID) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connInfo, exists := cm.connections[sessionID]
	if !exists {
		connInfo = &ConnectionInfo{
			SessionID: sessionID,
		}
		cm.connections[sessionID] = connInfo
	}

	connInfo.RetryCount++
}

// SetConnectionError define um erro de conexão
func (cm *ConnectionManager) SetConnectionError(sessionID uuid.UUID, err error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connInfo, exists := cm.connections[sessionID]
	if !exists {
		connInfo = &ConnectionInfo{
			SessionID: sessionID,
		}
		cm.connections[sessionID] = connInfo
	}

	connInfo.Status = StatusError
	connInfo.LastError = err.Error()
	now := time.Now()
	connInfo.LastSeen = &now
}

// OnConnectionSuccess deve ser chamado quando uma conexão é bem-sucedida
func (cm *ConnectionManager) OnConnectionSuccess(sessionID uuid.UUID, jid *types.JID) {
	cm.updateConnectionStatus(sessionID, StatusConnected)

	// Limpar QR code após sucesso
	cm.qrManager.ClearQRCode(sessionID)

	cm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid.String(),
	}).Info().Msg("Connection successful")
}

// OnConnectionLost deve ser chamado quando uma conexão é perdida
func (cm *ConnectionManager) OnConnectionLost(sessionID uuid.UUID) {
	cm.updateConnectionStatus(sessionID, StatusDisconnected)

	cm.logger.WithField("sessionId", sessionID).Warn().Msg("Connection lost")

	// Iniciar reconexão automática para sessões autenticadas
	go cm.startAutoReconnect(sessionID)
}

// startAutoReconnect inicia processo de reconexão automática
func (cm *ConnectionManager) startAutoReconnect(sessionID uuid.UUID) {
	// Aguardar um pouco antes de tentar reconectar
	time.Sleep(5 * time.Second)

	// Só reconectar sessões que já foram autenticadas (têm WaJID)
	needsAuth, err := cm.needsAuthentication(context.Background(), sessionID)
	if err != nil {
		cm.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("Failed to check authentication for auto-reconnect")
		return
	}

	if needsAuth {
		cm.logger.WithField("sessionId", sessionID).Debug().Msg("Skipping auto-reconnect for unauthenticated session")
		return
	}

	cm.logger.WithField("sessionId", sessionID).Info().Msg("Starting auto-reconnect")

	ctx := context.Background()
	if err := cm.ConnectWithRetry(ctx, sessionID); err != nil {
		cm.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("Auto-reconnect failed")
	} else {
		cm.logger.WithField("sessionId", sessionID).Info().Msg("Auto-reconnect successful")
	}
}

// startConnectionTimeout inicia um timeout para conexão
func (cm *ConnectionManager) startConnectionTimeout(ctx context.Context, sessionID uuid.UUID) {
	// Timeout de 5 minutos para conexão
	timeout := 5 * time.Minute

	select {
	case <-ctx.Done():
		// Contexto cancelado
		return
	case <-time.After(timeout):
		// Timeout atingido - verificar se ainda está connecting
		cm.mutex.RLock()
		connInfo, exists := cm.connections[sessionID]
		cm.mutex.RUnlock()

		if exists && connInfo.Status == StatusConnecting {
			cm.logger.WithField("sessionId", sessionID).Warn().Msg("Connection timeout - reverting to disconnected")
			cm.updateConnectionStatus(sessionID, StatusDisconnected)

			// Limpar QR code após timeout
			cm.qrManager.ClearQRCode(sessionID)
		}
	}
}

// needsAuthentication verifica se uma sessão precisa de autenticação
func (cm *ConnectionManager) needsAuthentication(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	// Para simplificar, vamos verificar diretamente se o device store tem ID
	state, err := cm.sessionManager.GetSession(sessionID)
	if err != nil {
		return false, fmt.Errorf("failed to get session state: %w", err)
	}

	// Se o device store não tem ID, precisa de autenticação
	if state.Client.Store.ID == nil {
		cm.logger.WithField("sessionId", sessionID).Debug().Msg("Session needs authentication - no device ID found")
		return true, nil
	}

	cm.logger.WithField("sessionId", sessionID).Debug().Msg("Session already authenticated")
	return false, nil
}

// restoreSessionFromWaJID restaura uma sessão usando o WaJID salvo
func (cm *ConnectionManager) restoreSessionFromWaJID(ctx context.Context, sessionID uuid.UUID, waJID string) error {
	cm.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"waJid":     waJID,
	}).Info().Msg("Restoring session from saved WaJID")

	// Parse do JID
	parsedJID, err := types.ParseJID(waJID)
	if err != nil {
		cm.logger.WithError(err).WithField("waJid", waJID).Error().Msg("Failed to parse WaJID")
		return fmt.Errorf("invalid WaJID format: %w", err)
	}

	// Obter session manager para acessar o container
	sessionManager, ok := cm.sessionManager.(*session.SessionManager)
	if !ok {
		return fmt.Errorf("invalid session manager type")
	}

	// Tentar obter device existente usando o JID
	container := sessionManager.GetContainer()
	deviceStore, err := container.GetDevice(ctx, parsedJID)
	if err != nil {
		cm.logger.WithError(err).WithField("waJid", waJID).Warn().Msg("Failed to get existing device, will need new authentication")
		return fmt.Errorf("device not found for JID: %w", err)
	}

	// Restaurar a sessão com o device existente
	err = sessionManager.RestoreSessionWithDevice(sessionID, parsedJID, deviceStore)
	if err != nil {
		cm.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("Failed to restore session with device")
		return fmt.Errorf("failed to restore session: %w", err)
	}

	cm.logger.WithField("sessionId", sessionID).Info().Msg("Session restored successfully from WaJID")
	return nil
}

// Close encerra o ConnectionManager
func (cm *ConnectionManager) Close() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	count := len(cm.connections)
	cm.connections = make(map[uuid.UUID]*ConnectionInfo)

	cm.logger.WithField("clearedCount", count).Info().Msg("ConnectionManager closed")
}
