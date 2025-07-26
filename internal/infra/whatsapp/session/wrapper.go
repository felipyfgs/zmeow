package session

import (
	"context"
	"time"
	"zmeow/internal/domain/session"
	"zmeow/internal/infra/database"
	"zmeow/pkg/logger"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// ManagerInterface define a interface necessária do Manager para evitar ciclo de imports
type ManagerInterface interface {
	GetLogger() logger.Logger
	GetDB() *bun.DB
	GetSessionState(sessionID uuid.UUID) (*WrapperSessionState, bool)
	UpdateSessionState(sessionID uuid.UUID, updater func(*WrapperSessionState))
	LockSessions()
	UnlockSessions()
}

// WrapperSessionState representa o estado de uma sessão (definição local para wrapper)
type WrapperSessionState struct {
	ID       uuid.UUID
	Status   string
	JID      *types.JID
	LastSeen *time.Time
	QRCode   string
	Webhook  string
}

// SessionWrapper wrapper para capturar eventos de uma sessão específica
type SessionWrapper struct {
	SessionID uuid.UUID
	Client    *whatsmeow.Client
	Manager   ManagerInterface
}

// handleEvent manipula eventos do WhatsApp para uma sessão específica
func (sw *SessionWrapper) handleEvent(evt interface{}) {
	// TODO: Implementar logging após resolver ciclo de imports
	// logger.LogRawEventPayload(sw.Manager.logger, sw.SessionID.String(), getEventType(evt), evt)

	switch v := evt.(type) {
	case *events.Connected:
		sw.handleConnected(v)
	case *events.Disconnected:
		sw.handleDisconnected(v)
	case *events.LoggedOut:
		sw.handleLoggedOut(v)
	case *events.Message:
		sw.handleMessage(v)
	case *events.QR:
		sw.handleQR(v)
	case *events.PairSuccess:
		sw.handlePairSuccess(v)
	case *events.PairError:
		sw.handlePairError(v)
	default:
		sw.Manager.GetLogger().WithFields(map[string]interface{}{
			"sessionId": sw.SessionID,
			"eventType": getEventType(evt),
		}).Debug().Msg("Unhandled WhatsApp event")
	}
}

// handleConnected manipula evento de conexão
func (sw *SessionWrapper) handleConnected(evt *events.Connected) {
	sw.Manager.LockSessions()
	defer sw.Manager.UnlockSessions()

	if state, exists := sw.Manager.GetSessionState(sw.SessionID); exists {
		sw.Manager.UpdateSessionState(sw.SessionID, func(s *WrapperSessionState) {
			s.Status = "connected"
			s.JID = sw.Client.Store.ID
			now := time.Now()
			s.LastSeen = &now
			s.QRCode = "" // Limpar QR code após conexão bem-sucedida
		})

		sw.Manager.GetLogger().WithFields(map[string]interface{}{
			"sessionId": sw.SessionID,
			"jid":       state.JID.String(),
		}).Info().Msg("Session connected")

		// Atualizar o JID e status no banco de dados
		go func() {
			repo := database.NewSessionRepository(sw.Manager.GetDB())
			if err := repo.UpdateJID(context.Background(), sw.SessionID, state.JID.String()); err != nil {
				sw.Manager.GetLogger().WithError(err).Error().Msg("Failed to update JID in database")
			}
			// Atualizar status para connected no banco
			if err := repo.UpdateStatus(context.Background(), sw.SessionID, session.WhatsAppStatusConnected); err != nil {
				sw.Manager.GetLogger().WithError(err).Error().Msg("Failed to update status to connected in database")
			}
		}()

		// TODO: Enviar webhook se configurado
		if state.Webhook != "" {
			go sw.sendWebhook("connected", map[string]interface{}{
				"sessionId": sw.SessionID,
				"jid":       state.JID.String(),
				"timestamp": time.Now(),
			})
		}
	}
}

// handleDisconnected manipula evento de desconexão
func (sw *SessionWrapper) handleDisconnected(evt *events.Disconnected) {
	sw.Manager.LockSessions()
	defer sw.Manager.UnlockSessions()

	if state, exists := sw.Manager.GetSessionState(sw.SessionID); exists {
		sw.Manager.UpdateSessionState(sw.SessionID, func(s *WrapperSessionState) {
			s.Status = "disconnected"
			now := time.Now()
			s.LastSeen = &now
		})

		sw.Manager.GetLogger().WithFields(map[string]interface{}{
			"sessionId": sw.SessionID,
		}).Info().Msg("Session disconnected")

		// Atualizar status para disconnected no banco
		go func() {
			repo := database.NewSessionRepository(sw.Manager.GetDB())
			if err := repo.UpdateStatus(context.Background(), sw.SessionID, session.WhatsAppStatusDisconnected); err != nil {
				sw.Manager.GetLogger().WithError(err).Error().Msg("Failed to update status to disconnected in database")
			}
		}()

		// TODO: Enviar webhook se configurado
		if state.Webhook != "" {
			go sw.sendWebhook("disconnected", map[string]interface{}{
				"sessionId": sw.SessionID,
				"timestamp": time.Now(),
			})
		}
	}
}

// handleLoggedOut manipula evento de logout
func (sw *SessionWrapper) handleLoggedOut(evt *events.LoggedOut) {
	sw.Manager.LockSessions()
	defer sw.Manager.UnlockSessions()

	if state, exists := sw.Manager.GetSessionState(sw.SessionID); exists {
		sw.Manager.UpdateSessionState(sw.SessionID, func(s *WrapperSessionState) {
			s.Status = "disconnected"
			s.JID = nil
			now := time.Now()
			s.LastSeen = &now
		})

		sw.Manager.GetLogger().WithFields(map[string]interface{}{
			"sessionId": sw.SessionID,
		}).Warn().Msg("Session logged out")

		// Atualizar status para disconnected no banco e limpar JID
		go func() {
			repo := database.NewSessionRepository(sw.Manager.GetDB())
			if err := repo.UpdateStatus(context.Background(), sw.SessionID, session.WhatsAppStatusDisconnected); err != nil {
				sw.Manager.GetLogger().WithError(err).Error().Msg("Failed to update status to disconnected in database")
			}
			// Limpar JID no banco quando fizer logout
			if err := repo.UpdateJID(context.Background(), sw.SessionID, ""); err != nil {
				sw.Manager.GetLogger().WithError(err).Error().Msg("Failed to clear JID in database")
			}
		}()

		// TODO: Enviar webhook se configurado
		if state.Webhook != "" {
			go sw.sendWebhook("logged_out", map[string]interface{}{
				"sessionId": sw.SessionID,
				"timestamp": time.Now(),
			})
		}
	}
}

// handleMessage manipula mensagens recebidas
func (sw *SessionWrapper) handleMessage(evt *events.Message) {
	sw.Manager.GetLogger().WithFields(map[string]interface{}{
		"sessionId":   sw.SessionID,
		"messageId":   evt.Info.ID,
		"from":        evt.Info.Sender.String(),
		"timestamp":   evt.Info.Timestamp,
		"messageType": getMessageType(evt.Message),
	}).Info().Msg("Message received")

	// Atualizar último visto
	if state, exists := sw.Manager.GetSessionState(sw.SessionID); exists {
		sw.Manager.UpdateSessionState(sw.SessionID, func(s *WrapperSessionState) {
			now := time.Now()
			s.LastSeen = &now
		})

		// TODO: Enviar webhook se configurado
		if state.Webhook != "" {
			go sw.sendWebhook("message", map[string]interface{}{
				"sessionId":   sw.SessionID,
				"messageId":   evt.Info.ID,
				"from":        evt.Info.Sender.String(),
				"timestamp":   evt.Info.Timestamp,
				"messageType": getMessageType(evt.Message),
				"message":     evt.Message,
			})
		}
	}
}

// handleQR manipula eventos de QR code
func (sw *SessionWrapper) handleQR(evt *events.QR) {
	sw.Manager.LockSessions()
	defer sw.Manager.UnlockSessions()

	if state, exists := sw.Manager.GetSessionState(sw.SessionID); exists {
		sw.Manager.UpdateSessionState(sw.SessionID, func(s *WrapperSessionState) {
			for _, code := range evt.Codes {
				s.QRCode = code
				break // Usar apenas o primeiro código
			}
		})

		sw.Manager.GetLogger().WithField("sessionId", sw.SessionID).Info().Msg("QR code updated")

		// TODO: Enviar webhook se configurado
		if state.Webhook != "" {
			go sw.sendWebhook("qr_code", map[string]interface{}{
				"sessionId": sw.SessionID,
				"qrCode":    state.QRCode,
				"timestamp": time.Now(),
			})
		}
	}
}

// handlePairSuccess manipula sucesso no pareamento
func (sw *SessionWrapper) handlePairSuccess(evt *events.PairSuccess) {
	sw.Manager.LockSessions()
	defer sw.Manager.UnlockSessions()

	if state, exists := sw.Manager.GetSessionState(sw.SessionID); exists {
		sw.Manager.UpdateSessionState(sw.SessionID, func(s *WrapperSessionState) {
			s.Status = "connected"
			s.JID = sw.Client.Store.ID
			now := time.Now()
			s.LastSeen = &now
			s.QRCode = ""
		})

		sw.Manager.GetLogger().WithFields(map[string]interface{}{
			"sessionId": sw.SessionID,
			"jid":       state.JID.String(),
		}).Info().Msg("Phone pairing successful")

		// Atualizar o JID e status no banco de dados
		go func() {
			repo := database.NewSessionRepository(sw.Manager.GetDB())
			if err := repo.UpdateJID(context.Background(), sw.SessionID, state.JID.String()); err != nil {
				sw.Manager.GetLogger().WithError(err).Error().Msg("Failed to update JID in database")
			}
			// Atualizar status para connected no banco
			if err := repo.UpdateStatus(context.Background(), sw.SessionID, session.WhatsAppStatusConnected); err != nil {
				sw.Manager.GetLogger().WithError(err).Error().Msg("Failed to update status to connected in database")
			}
		}()

		// TODO: Enviar webhook se configurado
		if state.Webhook != "" {
			go sw.sendWebhook("pair_success", map[string]interface{}{
				"sessionId": sw.SessionID,
				"jid":       state.JID.String(),
				"timestamp": time.Now(),
			})
		}
	}
}

// handlePairError manipula erro no pareamento
func (sw *SessionWrapper) handlePairError(evt *events.PairError) {
	sw.Manager.GetLogger().WithFields(map[string]interface{}{
		"sessionId": sw.SessionID,
		"error":     evt.Error.Error(),
	}).Error().Msg("Phone pairing failed")

	// TODO: Enviar webhook se configurado
	if state, exists := sw.Manager.GetSessionState(sw.SessionID); exists && state.Webhook != "" {
		go sw.sendWebhook("pair_error", map[string]interface{}{
			"sessionId": sw.SessionID,
			"error":     evt.Error.Error(),
			"timestamp": time.Now(),
		})
	}
}

// sendWebhook envia um webhook (implementação placeholder)
func (sw *SessionWrapper) sendWebhook(event string, data map[string]interface{}) {
	// TODO: Implementar envio de webhook HTTP
	sw.Manager.GetLogger().WithFields(map[string]interface{}{
		"sessionId": sw.SessionID,
		"event":     event,
		"data":      data,
	}).Info().Msg("Webhook would be sent")
}

// getEventType retorna o tipo do evento como string
func getEventType(evt interface{}) string {
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
	case *events.HistorySync:
		return "HistorySync"
	case *events.AppStateSyncComplete:
		return "AppStateSyncComplete"
	case *events.PushNameSetting:
		return "PushNameSetting"
	case *events.AppState:
		return "AppState"
	case *events.Archive:
		return "Archive"
	case *events.Blocklist:
		return "Blocklist"
	case *events.BusinessName:
		return "BusinessName"
	case *events.CallAccept:
		return "CallAccept"
	case *events.CallOffer:
		return "CallOffer"
	case *events.CallOfferNotice:
		return "CallOfferNotice"
	case *events.CallPreAccept:
		return "CallPreAccept"
	case *events.CallReject:
		return "CallReject"
	case *events.CallRelayLatency:
		return "CallRelayLatency"
	case *events.CallTerminate:
		return "CallTerminate"
	case *events.CallTransport:
		return "CallTransport"
	case *events.ChatPresence:
		return "ChatPresence"
	case *events.ClearChat:
		return "ClearChat"
	case *events.Contact:
		return "Contact"
	case *events.DeleteChat:
		return "DeleteChat"
	case *events.DeleteForMe:
		return "DeleteForMe"
	case *events.GroupInfo:
		return "GroupInfo"
	case *events.IdentityChange:
		return "IdentityChange"
	case *events.JoinedGroup:
		return "JoinedGroup"
	case *events.KeepAliveRestored:
		return "KeepAliveRestored"
	case *events.KeepAliveTimeout:
		return "KeepAliveTimeout"
	case *events.LabelAssociationChat:
		return "LabelAssociationChat"
	case *events.LabelAssociationMessage:
		return "LabelAssociationMessage"
	case *events.LabelEdit:
		return "LabelEdit"
	case *events.MarkChatAsRead:
		return "MarkChatAsRead"
	case *events.MediaRetry:
		return "MediaRetry"
	case *events.MediaRetryError:
		return "MediaRetryError"
	case *events.Mute:
		return "Mute"
	case *events.NewsletterJoin:
		return "NewsletterJoin"
	case *events.NewsletterLeave:
		return "NewsletterLeave"
	case *events.NewsletterLiveUpdate:
		return "NewsletterLiveUpdate"
	case *events.NewsletterMuteChange:
		return "NewsletterMuteChange"
	case *events.OfflineSyncCompleted:
		return "OfflineSyncCompleted"
	case *events.OfflineSyncPreview:
		return "OfflineSyncPreview"
	case *events.Picture:
		return "Picture"
	case *events.Pin:
		return "Pin"
	case *events.PrivacySettings:
		return "PrivacySettings"
	case *events.PushName:
		return "PushName"
	case *events.QRScannedWithoutMultidevice:
		return "QRScannedWithoutMultidevice"
	case *events.Star:
		return "Star"
	case *events.StreamError:
		return "StreamError"
	case *events.StreamReplaced:
		return "StreamReplaced"
	case *events.TemporaryBan:
		return "TemporaryBan"
	case *events.UnarchiveChatsSetting:
		return "UnarchiveChatsSetting"
	case *events.UndecryptableMessage:
		return "UndecryptableMessage"
	case *events.UnknownCallEvent:
		return "UnknownCallEvent"
	case *events.UserAbout:
		return "UserAbout"
	case *events.UserStatusMute:
		return "UserStatusMute"
	default:
		return "Unknown"
	}
}

// getMessageType retorna o tipo da mensagem
func getMessageType(msg interface{}) string {
	// TODO: Implementar detecção de tipo de mensagem baseada no protocolo do WhatsApp
	return "text" // Placeholder
}
