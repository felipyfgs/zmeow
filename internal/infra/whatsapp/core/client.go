package core

import (
	"context"

	"github.com/google/uuid"

	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// UnifiedClient é uma implementação consolidada que substitui Client, ClientWrapper e SessionClient
type UnifiedClient struct {
	manager   whatsapp.WhatsAppManager
	sessionID *uuid.UUID // Opcional - se definido, operações usam este ID por padrão
	logger    logger.Logger
}

// NewUnifiedClient cria uma nova instância do cliente unificado
func NewUnifiedClient(manager whatsapp.WhatsAppManager, log logger.Logger) whatsapp.WhatsAppClient {
	return &UnifiedClient{
		manager: manager,
		logger:  log.WithComponent("unified-whatsapp-client"),
	}
}

// NewUnifiedClientForSession cria um cliente unificado para uma sessão específica
func NewUnifiedClientForSession(manager whatsapp.WhatsAppManager, sessionID uuid.UUID, log logger.Logger) whatsapp.WhatsAppClient {
	return &UnifiedClient{
		manager:   manager,
		sessionID: &sessionID,
		logger:    log.WithComponent("unified-whatsapp-client").WithField("sessionId", sessionID),
	}
}

// Connect conecta uma sessão ao WhatsApp
func (uc *UnifiedClient) Connect(ctx context.Context, sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Connecting session")

	err := uc.manager.ConnectSession(ctx, targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to connect session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session connected successfully")
	return nil
}

// Disconnect desconecta uma sessão do WhatsApp
func (uc *UnifiedClient) Disconnect(ctx context.Context, sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Disconnecting session")

	err := uc.manager.DisconnectSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to disconnect session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session disconnected successfully")
	return nil
}

// GetQRCode obtém o QR code para autenticação
func (uc *UnifiedClient) GetQRCode(ctx context.Context, sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Getting QR code")

	qrCode, err := uc.manager.GetQRCode(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get QR code")
		return "", err
	}

	if qrCode != "" {
		uc.logger.WithField("sessionId", targetSessionID).Info().Msg("QR code retrieved successfully")
	} else {
		uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("No QR code available")
	}

	return qrCode, nil
}

// PairPhone realiza pareamento via número de telefone
func (uc *UnifiedClient) PairPhone(ctx context.Context, sessionID uuid.UUID, phone string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
	}).Debug().Msg("Starting phone pairing")

	code, err := uc.manager.PairPhone(targetSessionID, phone)
	if err != nil {
		uc.logger.WithError(err).WithFields(map[string]interface{}{
			"sessionId": targetSessionID,
			"phone":     phone,
		}).Error().Msg("Failed to pair phone")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"code":      code,
	}).Info().Msg("Phone pairing code generated")

	return code, nil
}

// IsConnected verifica se uma sessão está conectada
func (uc *UnifiedClient) IsConnected(sessionID uuid.UUID) bool {
	targetSessionID := uc.resolveSessionID(sessionID)

	connected := uc.manager.IsConnected(targetSessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"connected": connected,
	}).Debug().Msg("Checked connection status")

	return connected
}

// SetProxy configura proxy para uma sessão
func (uc *UnifiedClient) SetProxy(sessionID uuid.UUID, proxyURL string) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"proxyUrl":  proxyURL,
	}).Debug().Msg("Setting proxy")

	err := uc.manager.SetProxy(targetSessionID, proxyURL)
	if err != nil {
		uc.logger.WithError(err).WithFields(map[string]interface{}{
			"sessionId": targetSessionID,
			"proxyUrl":  proxyURL,
		}).Error().Msg("Failed to set proxy")
		return err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"proxyUrl":  proxyURL,
	}).Info().Msg("Proxy configured successfully")

	return nil
}

// GetSessionStatus retorna o status de uma sessão
func (uc *UnifiedClient) GetSessionStatus(sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	status, err := uc.manager.GetSessionStatus(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get session status")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"status":    status,
	}).Debug().Msg("Retrieved session status")

	return status, nil
}

// GetSessionJID retorna o JID de uma sessão
func (uc *UnifiedClient) GetSessionJID(sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	jid, err := uc.manager.GetSessionJID(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get session JID")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"jid":       jid,
	}).Debug().Msg("Retrieved session JID")

	return jid, nil
}

// resolveSessionID resolve qual session ID usar (parâmetro ou padrão)
func (uc *UnifiedClient) resolveSessionID(sessionID uuid.UUID) uuid.UUID {
	// Se um sessionID foi fornecido e não é nil, usar ele
	if sessionID != uuid.Nil {
		return sessionID
	}

	// Se o cliente tem um sessionID padrão, usar ele
	if uc.sessionID != nil {
		return *uc.sessionID
	}

	// Caso contrário, usar o sessionID fornecido (mesmo que seja nil)
	return sessionID
}

// SetDefaultSessionID define um session ID padrão para este cliente
func (uc *UnifiedClient) SetDefaultSessionID(sessionID uuid.UUID) {
	uc.sessionID = &sessionID
	uc.logger = uc.logger.WithField("sessionId", sessionID)
}

// GetDefaultSessionID retorna o session ID padrão (se definido)
func (uc *UnifiedClient) GetDefaultSessionID() *uuid.UUID {
	return uc.sessionID
}

// Clone cria uma cópia do cliente para uma sessão específica
func (uc *UnifiedClient) Clone(sessionID uuid.UUID) whatsapp.WhatsAppClient {
	return NewUnifiedClientForSession(uc.manager, sessionID, uc.logger)
}

// CreateSession cria uma nova sessão WhatsApp
func (uc *UnifiedClient) CreateSession(sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Creating session")

	err := uc.manager.RegisterSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to create session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session created successfully")
	return nil
}

// DeleteSession remove uma sessão WhatsApp
func (uc *UnifiedClient) DeleteSession(sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Deleting session")

	err := uc.manager.RemoveSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to delete session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session deleted successfully")
	return nil
}

// GetJID obtém o JID (WhatsApp ID) de uma sessão conectada
func (uc *UnifiedClient) GetJID(sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	jid, err := uc.manager.GetSessionJID(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get session JID")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"jid":       jid,
	}).Debug().Msg("Retrieved session JID")

	return jid, nil
}

// Logout realiza logout de uma sessão
func (uc *UnifiedClient) Logout(ctx context.Context, sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Logging out session")

	// Primeiro desconectar
	err := uc.Disconnect(ctx, targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Warn().Msg("Failed to disconnect during logout")
	}

	// Depois remover a sessão
	err = uc.DeleteSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to logout session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session logged out successfully")
	return nil
}

// GetUnderlyingManager retorna o manager subjacente (para uso avançado)
func (uc *UnifiedClient) GetUnderlyingManager() whatsapp.WhatsAppManager {
	return uc.manager
}

// Adapter functions para compatibilidade com implementações antigas

// NewClientAdapter cria um adapter que substitui o Client antigo
func NewClientAdapter(manager *Manager, log logger.Logger) whatsapp.WhatsAppManager {
	// Criar um wrapper que implementa WhatsAppManager usando UnifiedClient
	return &ManagerAdapter{
		client:  NewUnifiedClient(manager, log),
		manager: manager,
		logger:  log,
	}
}

// ManagerAdapter adapta UnifiedClient para implementar WhatsAppManager
type ManagerAdapter struct {
	client  whatsapp.WhatsAppClient
	manager whatsapp.WhatsAppManager
	logger  logger.Logger
}

// Implementar interface WhatsAppManager delegando para o manager subjacente
func (ma *ManagerAdapter) RegisterSession(sessionID uuid.UUID) error {
	return ma.manager.RegisterSession(sessionID)
}

func (ma *ManagerAdapter) ConnectSession(ctx context.Context, sessionID uuid.UUID) error {
	return ma.client.Connect(ctx, sessionID)
}

func (ma *ManagerAdapter) DisconnectSession(sessionID uuid.UUID) error {
	return ma.client.Disconnect(context.Background(), sessionID)
}

func (ma *ManagerAdapter) GetQRCode(sessionID uuid.UUID) (string, error) {
	return ma.client.GetQRCode(context.Background(), sessionID)
}

func (ma *ManagerAdapter) PairPhone(sessionID uuid.UUID, phoneNumber string) (string, error) {
	return ma.client.PairPhone(context.Background(), sessionID, phoneNumber)
}

func (ma *ManagerAdapter) IsConnected(sessionID uuid.UUID) bool {
	return ma.client.IsConnected(sessionID)
}

func (ma *ManagerAdapter) SetProxy(sessionID uuid.UUID, proxyURL string) error {
	return ma.client.SetProxy(sessionID, proxyURL)
}

func (ma *ManagerAdapter) GetSessionStatus(sessionID uuid.UUID) (string, error) {
	return ma.manager.GetSessionStatus(sessionID)
}

func (ma *ManagerAdapter) GetSessionJID(sessionID uuid.UUID) (string, error) {
	return ma.manager.GetSessionJID(sessionID)
}

func (ma *ManagerAdapter) RemoveSession(sessionID uuid.UUID) error {
	return ma.manager.RemoveSession(sessionID)
}

func (ma *ManagerAdapter) RestoreSession(ctx context.Context, sessionID uuid.UUID, jid string) error {
	return ma.manager.RestoreSession(ctx, sessionID, jid)
}

func (ma *ManagerAdapter) GetClient(sessionID uuid.UUID) (whatsapp.WhatsAppClient, error) {
	// Criar um novo cliente unificado para a sessão específica
	if unifiedClient, ok := ma.client.(*UnifiedClient); ok {
		return unifiedClient.Clone(sessionID), nil
	}

	// Fallback: criar um novo cliente usando o manager
	return NewUnifiedClientForSession(ma.manager, sessionID, ma.logger), nil
}
