package events

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	"zmeow/internal/domain/session"
	sessionpkg "zmeow/internal/infra/whatsapp/session"
	"zmeow/pkg/logger"
)

// SessionManagerInterface define os métodos necessários do session manager
type SessionManagerInterface interface {
	GetSession(sessionID uuid.UUID) (*sessionpkg.SessionState, error)
	UpdateSessionStatus(sessionID uuid.UUID, status string) error
	UpdateSessionQRCode(sessionID uuid.UUID, qrCode string) error
	UpdateSessionJID(sessionID uuid.UUID, jid *types.JID) error
}

// EventProcessor processa eventos do WhatsApp
type EventProcessor struct {
	sessionManager SessionManagerInterface
	dbRepo         session.SessionRepository
	webhookService WebhookService
	logger         logger.Logger
}

// WebhookService interface para envio de webhooks
type WebhookService interface {
	SendWebhook(sessionID uuid.UUID, event string, data map[string]interface{}) error
}

// NewEventProcessor cria uma nova instância do EventProcessor
func NewEventProcessor(
	sessionManager SessionManagerInterface,
	dbRepo session.SessionRepository,
	webhookService WebhookService,
	log logger.Logger,
) *EventProcessor {
	return &EventProcessor{
		sessionManager: sessionManager,
		dbRepo:         dbRepo,
		webhookService: webhookService,
		logger:         log.WithComponent("event-processor"),
	}
}

// ProcessEvent processa um evento do WhatsApp
func (ep *EventProcessor) ProcessEvent(sessionID uuid.UUID, evt interface{}) {
	// Log do evento para debugging
	ep.logRawEvent(sessionID, evt)

	switch v := evt.(type) {
	case *events.Connected:
		ep.handleConnected(sessionID, v)
	case *events.Disconnected:
		ep.handleDisconnected(sessionID, v)
	case *events.LoggedOut:
		ep.handleLoggedOut(sessionID, v)
	case *events.Message:
		ep.handleMessage(sessionID, v)
	case *events.QR:
		// QR events são processados pelo QRCodeManager via canal QR
		// Não processamos aqui para evitar duplicação
		ep.logger.WithField("sessionId", sessionID).Debug().Msg("QR event received - handled by QRCodeManager")
	case *events.PairSuccess:
		ep.handlePairSuccess(sessionID, v)
	case *events.PairError:
		ep.handlePairError(sessionID, v)
	default:
		ep.logger.WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"eventType": GetEventTypeName(evt),
		}).Debug().Msg("Unhandled WhatsApp event")
	}
}

// handleConnected processa evento de conexão
func (ep *EventProcessor) handleConnected(sessionID uuid.UUID, _ *events.Connected) {
	state, err := ep.sessionManager.GetSession(sessionID)
	if err != nil {
		ep.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("Failed to get session for connected event")
		return
	}

	// Atualizar estado da sessão
	if err := ep.sessionManager.UpdateSessionStatus(sessionID, "connected"); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update session status")
	}

	// Atualizar JID se disponível
	if state.Client != nil && state.Client.Store.ID != nil {
		if err := ep.sessionManager.UpdateSessionJID(sessionID, state.Client.Store.ID); err != nil {
			ep.logger.WithError(err).Error().Msg("Failed to update session JID")
		}
	}

	// Limpar QR code após conexão
	if err := ep.sessionManager.UpdateSessionQRCode(sessionID, ""); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to clear QR code")
	}

	ep.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       state.JID,
	}).Info().Msg("Session connected")

	// Atualizar banco de dados
	jid := ""
	if state.JID != nil {
		jid = state.JID.String()
	}
	go ep.updateDatabaseOnConnect(sessionID, jid)

	// Enviar webhook
	go ep.sendConnectedWebhook(sessionID, jid)
}

// handleDisconnected processa evento de desconexão
func (ep *EventProcessor) handleDisconnected(sessionID uuid.UUID, _ *events.Disconnected) {
	if err := ep.sessionManager.UpdateSessionStatus(sessionID, "disconnected"); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update session status")
	}

	ep.logger.WithField("sessionId", sessionID).Info().Msg("Session disconnected")

	// Atualizar banco de dados
	go ep.updateDatabaseStatus(sessionID, session.WhatsAppStatusDisconnected)

	// Enviar webhook
	go ep.sendDisconnectedWebhook(sessionID)
}

// handleLoggedOut processa evento de logout
func (ep *EventProcessor) handleLoggedOut(sessionID uuid.UUID, _ *events.LoggedOut) {
	if err := ep.sessionManager.UpdateSessionStatus(sessionID, "disconnected"); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update session status")
	}

	// TODO: Implementar limpeza de JID quando necessário

	ep.logger.WithField("sessionId", sessionID).Warn().Msg("Session logged out")

	// Atualizar banco de dados
	go ep.updateDatabaseOnLogout(sessionID)

	// Enviar webhook
	go ep.sendLoggedOutWebhook(sessionID)
}

// handleMessage processa mensagens recebidas
func (ep *EventProcessor) handleMessage(sessionID uuid.UUID, evt *events.Message) {
	ep.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"messageId":   evt.Info.ID,
		"from":        evt.Info.Sender.String(),
		"timestamp":   evt.Info.Timestamp,
		"messageType": getMessageType(evt.Message),
	}).Info().Msg("Message received")

	// Atualizar último visto
	if err := ep.sessionManager.UpdateSessionStatus(sessionID, "connected"); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update last seen")
	}

	// Enviar webhook
	go ep.sendMessageWebhook(sessionID, evt)
}

// handlePairSuccess processa sucesso no pareamento
func (ep *EventProcessor) handlePairSuccess(sessionID uuid.UUID, _ *events.PairSuccess) {
	state, err := ep.sessionManager.GetSession(sessionID)
	if err != nil {
		ep.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("Failed to get session for pair success event")
		return
	}

	// Atualizar estado da sessão
	if err := ep.sessionManager.UpdateSessionStatus(sessionID, "connected"); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update session status")
	}

	// Atualizar JID se disponível
	if state.Client != nil && state.Client.Store.ID != nil {
		if err := ep.sessionManager.UpdateSessionJID(sessionID, state.Client.Store.ID); err != nil {
			ep.logger.WithError(err).Error().Msg("Failed to update session JID")
		}
	}

	// Limpar QR code após pareamento
	if err := ep.sessionManager.UpdateSessionQRCode(sessionID, ""); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to clear QR code")
	}

	ep.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"jid":       state.JID,
	}).Info().Msg("Phone pairing successful")

	// Atualizar banco de dados
	jid := ""
	if state.JID != nil {
		jid = state.JID.String()
	}
	go ep.updateDatabaseOnConnect(sessionID, jid)

	// Enviar webhook
	go ep.sendPairSuccessWebhook(sessionID, jid)
}

// handlePairError processa erro no pareamento
func (ep *EventProcessor) handlePairError(sessionID uuid.UUID, evt *events.PairError) {
	ep.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"error":     evt.Error.Error(),
	}).Error().Msg("Phone pairing failed")

	// Enviar webhook
	go ep.sendPairErrorWebhook(sessionID, evt.Error.Error())
}

// Métodos auxiliares para atualização do banco de dados
func (ep *EventProcessor) updateDatabaseOnConnect(sessionID uuid.UUID, jid string) {
	ctx := context.Background()
	if err := ep.dbRepo.UpdateJID(ctx, sessionID, jid); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update JID in database")
	}
	if err := ep.dbRepo.UpdateStatus(ctx, sessionID, session.WhatsAppStatusConnected); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update status to connected in database")
	}
}

func (ep *EventProcessor) updateDatabaseStatus(sessionID uuid.UUID, status session.WhatsAppSessionStatus) {
	ctx := context.Background()
	if err := ep.dbRepo.UpdateStatus(ctx, sessionID, status); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update status in database")
	}
}

func (ep *EventProcessor) updateDatabaseOnLogout(sessionID uuid.UUID) {
	ctx := context.Background()
	if err := ep.dbRepo.UpdateStatus(ctx, sessionID, session.WhatsAppStatusDisconnected); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to update status to disconnected in database")
	}
	if err := ep.dbRepo.UpdateJID(ctx, sessionID, ""); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to clear JID in database")
	}
}

// Métodos auxiliares para webhooks
func (ep *EventProcessor) sendConnectedWebhook(sessionID uuid.UUID, jid string) {
	data := map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid,
		"timestamp": time.Now(),
	}
	if err := ep.webhookService.SendWebhook(sessionID, "connected", data); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to send connected webhook")
	}
}

func (ep *EventProcessor) sendDisconnectedWebhook(sessionID uuid.UUID) {
	data := map[string]interface{}{
		"sessionId": sessionID,
		"timestamp": time.Now(),
	}
	if err := ep.webhookService.SendWebhook(sessionID, "disconnected", data); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to send disconnected webhook")
	}
}

func (ep *EventProcessor) sendLoggedOutWebhook(sessionID uuid.UUID) {
	data := map[string]interface{}{
		"sessionId": sessionID,
		"timestamp": time.Now(),
	}
	if err := ep.webhookService.SendWebhook(sessionID, "logged_out", data); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to send logged out webhook")
	}
}

func (ep *EventProcessor) sendMessageWebhook(sessionID uuid.UUID, evt *events.Message) {
	data := map[string]interface{}{
		"sessionId":   sessionID,
		"messageId":   evt.Info.ID,
		"from":        evt.Info.Sender.String(),
		"timestamp":   evt.Info.Timestamp,
		"messageType": getMessageType(evt.Message),
		"message":     evt.Message,
	}
	if err := ep.webhookService.SendWebhook(sessionID, "message", data); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to send message webhook")
	}
}

func (ep *EventProcessor) sendPairSuccessWebhook(sessionID uuid.UUID, jid string) {
	data := map[string]interface{}{
		"sessionId": sessionID,
		"jid":       jid,
		"timestamp": time.Now(),
	}
	if err := ep.webhookService.SendWebhook(sessionID, "pair_success", data); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to send pair success webhook")
	}
}

func (ep *EventProcessor) sendPairErrorWebhook(sessionID uuid.UUID, errorMsg string) {
	data := map[string]interface{}{
		"sessionId": sessionID,
		"error":     errorMsg,
		"timestamp": time.Now(),
	}
	if err := ep.webhookService.SendWebhook(sessionID, "pair_error", data); err != nil {
		ep.logger.WithError(err).Error().Msg("Failed to send pair error webhook")
	}
}

// logRawEvent registra o evento bruto para debugging
func (ep *EventProcessor) logRawEvent(sessionID uuid.UUID, evt interface{}) {
	ep.logger.WithFields(map[string]interface{}{
		"sessionId":  sessionID,
		"eventType":  GetEventTypeName(evt),
		"rawPayload": evt,
	}).Trace().Msg("Raw WhatsApp Event Payload")
}

// getMessageType retorna o tipo da mensagem
func getMessageType(_ interface{}) string {
	// TODO: Implementar detecção de tipo de mensagem baseada no protocolo do WhatsApp
	return "text" // Placeholder por enquanto
}

// GetEventTypeName retorna o nome do tipo de evento
func GetEventTypeName(evt interface{}) string {
	// Mapeamento básico de tipos de evento
	switch evt.(type) {
	case *events.Connected:
		return "Connected"
	case *events.Disconnected:
		return "Disconnected"
	case *events.LoggedOut:
		return "LoggedOut"
	case *events.Message:
		return "Message"
	case *events.QR:
		return "QR"
	case *events.PairSuccess:
		return "PairSuccess"
	case *events.PairError:
		return "PairError"
	case *events.Receipt:
		return "Receipt"
	case *events.Presence:
		return "Presence"
	default:
		return "Unknown"
	}
}
