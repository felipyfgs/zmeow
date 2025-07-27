package whatsapp

import (
	"context"
	"zmeow/internal/domain/message"

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

	// SendTextMessage envia uma mensagem de texto
	SendTextMessage(ctx context.Context, sessionID uuid.UUID, phone, message string) (string, error)

	// SendMediaMessage envia mídia (imagem, áudio, vídeo, documento)
	SendMediaMessage(ctx context.Context, sessionID uuid.UUID, phone, mediaType string, mediaData []byte, caption, fileName, mimeType string) (string, error)

	// SendMediaFromURL baixa mídia de uma URL e envia como mensagem
	SendMediaFromURL(ctx context.Context, sessionID uuid.UUID, phone, mediaType, mediaURL, caption, fileName, mimeType string) (string, error)

	// SendLocationMessage envia uma localização
	SendLocationMessage(ctx context.Context, sessionID uuid.UUID, phone string, latitude, longitude float64, name, address string) (string, error)

	// SendContactMessage envia um contato
	SendContactMessage(ctx context.Context, sessionID uuid.UUID, phone, contactName, contactJID string) (string, error)

	// SendStickerMessage envia um sticker
	SendStickerMessage(ctx context.Context, sessionID uuid.UUID, phone string, stickerData []byte, mimeType string) (string, error)

	// SendButtonsMessage envia mensagem com botões
	SendButtonsMessage(ctx context.Context, sessionID uuid.UUID, phone, text, footer string, buttons []message.MessageButton) (string, error)

	// SendListMessage envia mensagem com lista
	SendListMessage(ctx context.Context, sessionID uuid.UUID, phone, text, footer, title, buttonText string, sections []message.MessageListSection) (string, error)

	// SendPollMessage envia enquete
	SendPollMessage(ctx context.Context, sessionID uuid.UUID, phone, name string, options []string, selectableCount int) (string, error)

	// EditMessage edita mensagem existente
	EditMessage(ctx context.Context, sessionID uuid.UUID, phone, messageID, newText string) (string, error)

	// DeleteMessage deleta uma mensagem
	DeleteMessage(ctx context.Context, sessionID uuid.UUID, phone, messageID string, forMe bool) error

	// ReactMessage reage a uma mensagem
	ReactMessage(ctx context.Context, sessionID uuid.UUID, phone, messageID, emoji string) error
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
