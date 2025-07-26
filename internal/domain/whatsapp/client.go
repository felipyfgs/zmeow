package whatsapp

import (
	"context"

	"github.com/google/uuid"
)

// WhatsAppClient define as operações do cliente WhatsApp
type WhatsAppClient interface {
	// Connect conecta uma sessão ao WhatsApp
	Connect(ctx context.Context, sessionID uuid.UUID) error

	// Disconnect desconecta uma sessão do WhatsApp
	Disconnect(ctx context.Context, sessionID uuid.UUID) error

	// GetQRCode obtém o QR code para autenticação
	GetQRCode(ctx context.Context, sessionID uuid.UUID) (string, error)

	// PairPhone realiza pareamento via número de telefone
	PairPhone(ctx context.Context, sessionID uuid.UUID, phone string) (string, error)

	// IsConnected verifica se uma sessão está conectada
	IsConnected(sessionID uuid.UUID) bool

	// SetProxy configura proxy para uma sessão
	SetProxy(sessionID uuid.UUID, proxyURL string) error

	// GetJID obtém o JID (WhatsApp ID) de uma sessão conectada
	GetJID(sessionID uuid.UUID) (string, error)

	// Logout realiza logout de uma sessão
	Logout(ctx context.Context, sessionID uuid.UUID) error

	// CreateSession cria uma nova sessão WhatsApp
	CreateSession(sessionID uuid.UUID) error

	// DeleteSession remove uma sessão WhatsApp
	DeleteSession(sessionID uuid.UUID) error
}

// WhatsAppManager gerencia múltiplas sessões WhatsApp
type WhatsAppManager interface {
	// RegisterSession registra uma nova sessão
	RegisterSession(sessionID uuid.UUID) error

	// ConnectSession conecta uma sessão específica
	ConnectSession(ctx context.Context, sessionID uuid.UUID) error

	// DisconnectSession desconecta uma sessão específica
	DisconnectSession(sessionID uuid.UUID) error

	// GetQRCode retorna o QR code de uma sessão
	GetQRCode(sessionID uuid.UUID) (string, error)

	// PairPhone realiza pareamento por telefone
	PairPhone(sessionID uuid.UUID, phoneNumber string) (string, error)

	// IsConnected verifica se uma sessão está conectada
	IsConnected(sessionID uuid.UUID) bool

	// SetProxy configura proxy para uma sessão
	SetProxy(sessionID uuid.UUID, proxyURL string) error

	// GetSessionStatus retorna o status de uma sessão
	GetSessionStatus(sessionID uuid.UUID) (string, error)

	// GetSessionJID retorna o JID de uma sessão
	GetSessionJID(sessionID uuid.UUID) (string, error)

	// RemoveSession remove uma sessão
	RemoveSession(sessionID uuid.UUID) error

	// RestoreSession restaura uma sessão a partir do JID
	RestoreSession(ctx context.Context, sessionID uuid.UUID, jid string) error
	// GetClient retorna o cliente WhatsApp de uma sessão específica
	GetClient(sessionID uuid.UUID) (WhatsAppClient, error)
}
