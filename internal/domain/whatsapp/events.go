package whatsapp

import (
	"time"

	"github.com/google/uuid"
)

// EventType define os tipos de eventos WhatsApp
type EventType string

const (
	EventConnected    EventType = "connected"
	EventDisconnected EventType = "disconnected"
	EventConnecting   EventType = "connecting"
	EventQRCode       EventType = "qr_code"
	EventPairSuccess  EventType = "pair_success"
	EventPairCode     EventType = "pair_code"
	EventError        EventType = "error"
)

// Event representa um evento do WhatsApp
type Event struct {
	Type      EventType   `json:"type"`
	SessionID uuid.UUID   `json:"sessionId"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// ConnectedEvent dados do evento de conexão
type ConnectedEvent struct {
	JID string `json:"jid"`
}

// DisconnectedEvent dados do evento de desconexão
type DisconnectedEvent struct {
	Reason string `json:"reason"`
}

// QRCodeEvent dados do evento de QR code
type QRCodeEvent struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// PairCodeEvent dados do evento de código de pareamento
type PairCodeEvent struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ErrorEvent dados do evento de erro
type ErrorEvent struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// EventHandler define um manipulador de eventos
type EventHandler interface {
	HandleEvent(event Event)
}

// EventHandlerFunc é um adaptador para usar funções como EventHandler
type EventHandlerFunc func(event Event)

func (f EventHandlerFunc) HandleEvent(event Event) {
	f(event)
}

// EventBus gerencia a distribuição de eventos
type EventBus interface {
	// Subscribe registra um handler para eventos
	Subscribe(handler EventHandler)

	// Unsubscribe remove um handler
	Unsubscribe(handler EventHandler)

	// Publish publica um evento
	Publish(event Event)

	// PublishAsync publica um evento de forma assíncrona
	PublishAsync(event Event)
}