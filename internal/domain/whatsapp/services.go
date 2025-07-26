package whatsapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventProcessor define operações para processamento de eventos WhatsApp
type EventProcessor interface {
	// ProcessEvent processa um evento do WhatsApp
	ProcessEvent(sessionID uuid.UUID, evt interface{})
}

// QRCodeManager define operações para gerenciamento de QR codes
type QRCodeManager interface {
	// GetQRCode retorna o QR code atual de uma sessão
	GetQRCode(sessionID uuid.UUID) (string, error)

	// ClearQRCode remove o QR code de uma sessão
	ClearQRCode(sessionID uuid.UUID)

	// IsQRCodeValid verifica se há um QR code válido para a sessão
	IsQRCodeValid(sessionID uuid.UUID) bool
}

// ConnectionManager define operações para gerenciamento de conexões
type ConnectionManager interface {
	// Connect conecta uma sessão ao WhatsApp
	Connect(ctx context.Context, sessionID uuid.UUID) error

	// Disconnect desconecta uma sessão
	Disconnect(sessionID uuid.UUID) error

	// Reconnect reconecta uma sessão
	Reconnect(ctx context.Context, sessionID uuid.UUID) error

	// IsConnected verifica se uma sessão está conectada
	IsConnected(sessionID uuid.UUID) bool

	// PairPhone realiza pareamento via telefone
	PairPhone(ctx context.Context, sessionID uuid.UUID, phoneNumber string) (string, error)
}

// WebhookService define operações para webhooks
type WebhookService interface {
	// SendWebhook envia um webhook para uma sessão
	SendWebhook(sessionID uuid.UUID, event string, data map[string]interface{}) error

	// SetWebhookConfig configura webhook para uma sessão
	SetWebhookConfig(config *WebhookConfig) error

	// GetWebhookConfig retorna a configuração de webhook de uma sessão
	GetWebhookConfig(sessionID uuid.UUID) (*WebhookConfig, error)

	// RemoveWebhookConfig remove a configuração de webhook de uma sessão
	RemoveWebhookConfig(sessionID uuid.UUID) error

	// EnableWebhook habilita webhook para uma sessão
	EnableWebhook(sessionID uuid.UUID) error

	// DisableWebhook desabilita webhook para uma sessão
	DisableWebhook(sessionID uuid.UUID) error
}

// WebhookConfig representa a configuração de webhook
type WebhookConfig struct {
	SessionID uuid.UUID `json:"sessionId"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret,omitempty"`
	Enabled   bool      `json:"enabled"`
	Retries   int       `json:"retries"`
	Timeout   int       `json:"timeout"` // em segundos
}

// SessionManager define operações para gerenciamento de sessões WhatsApp
type SessionManager interface {
	// CreateSession cria uma nova sessão
	CreateSession(sessionID uuid.UUID) error

	// GetSession retorna uma sessão pelo ID
	GetSession(sessionID uuid.UUID) (*SessionInfo, error)

	// RemoveSession remove uma sessão
	RemoveSession(sessionID uuid.UUID) error

	// UpdateSessionStatus atualiza o status de uma sessão
	UpdateSessionStatus(sessionID uuid.UUID, status string) error

	// IsConnected verifica se uma sessão está conectada
	IsConnected(sessionID uuid.UUID) bool

	// GetAllSessions retorna todas as sessões
	GetAllSessions() map[uuid.UUID]*SessionInfo
}

// SessionInfo representa informações de uma sessão
type SessionInfo struct {
	ID        uuid.UUID  `json:"id"`
	Status    string     `json:"status"`
	JID       string     `json:"jid,omitempty"`
	LastSeen  *time.Time `json:"lastSeen,omitempty"`
	QRCode    string     `json:"qrCode,omitempty"`
	ProxyURL  string     `json:"proxyUrl,omitempty"`
	Webhook   string     `json:"webhook,omitempty"`
	IsActive  bool       `json:"isActive"`
}

// ConfigService define operações para configurações
type ConfigService interface {
	// Get obtém uma configuração por chave
	Get(key string) (string, error)

	// Set define uma configuração
	Set(key, value string) error

	// GetDatabaseDSN retorna a DSN do banco de dados
	GetDatabaseDSN() string

	// GetWhatsAppConfig retorna configurações do WhatsApp
	GetWhatsAppConfig() *WhatsAppConfig
}

// WhatsAppConfig representa configurações do WhatsApp
type WhatsAppConfig struct {
	DebugLevel  string `json:"debugLevel"`
	StorePrefix string `json:"storePrefix"`
	DatabaseDSN string `json:"databaseDSN"`
}

// MetricsService define operações para métricas
type MetricsService interface {
	// RecordEvent registra um evento para métricas
	RecordEvent(sessionID uuid.UUID, eventType string, data map[string]interface{}) error

	// GetSessionStats obtém estatísticas de uma sessão
	GetSessionStats(sessionID uuid.UUID) (*SessionStats, error)

	// GetGlobalStats obtém estatísticas globais
	GetGlobalStats() (*GlobalStats, error)
}

// SessionStats representa estatísticas de uma sessão
type SessionStats struct {
	SessionID        uuid.UUID `json:"sessionId"`
	MessagesReceived int64     `json:"messagesReceived"`
	MessagesSent     int64     `json:"messagesSent"`
	ConnectionTime   int64     `json:"connectionTime"` // em segundos
	LastActivity     time.Time `json:"lastActivity"`
	ErrorCount       int64     `json:"errorCount"`
}

// GlobalStats representa estatísticas globais
type GlobalStats struct {
	TotalSessions     int64 `json:"totalSessions"`
	ActiveSessions    int64 `json:"activeSessions"`
	ConnectedSessions int64 `json:"connectedSessions"`
	TotalMessages     int64 `json:"totalMessages"`
	TotalErrors       int64 `json:"totalErrors"`
	UptimeSeconds     int64 `json:"uptimeSeconds"`
}

// CacheService define operações de cache
type CacheService interface {
	// Get obtém um valor do cache
	Get(key string) (string, error)

	// Set define um valor no cache com TTL
	Set(key, value string, ttl time.Duration) error

	// Delete remove um valor do cache
	Delete(key string) error

	// Exists verifica se uma chave existe no cache
	Exists(key string) (bool, error)

	// Clear limpa todo o cache
	Clear() error
}

// ValidationService define operações de validação
type ValidationService interface {
	// ValidatePhoneNumber valida um número de telefone
	ValidatePhoneNumber(phone string) error

	// ValidateWebhookURL valida uma URL de webhook
	ValidateWebhookURL(url string) error

	// ValidateSessionName valida um nome de sessão
	ValidateSessionName(name string) error

	// ValidateJID valida um JID do WhatsApp
	ValidateJID(jid string) error
}

// SecurityService define operações de segurança
type SecurityService interface {
	// GenerateSignature gera uma assinatura para webhook
	GenerateSignature(data []byte, secret string) string

	// ValidateSignature valida uma assinatura de webhook
	ValidateSignature(data []byte, signature, secret string) bool

	// EncryptData criptografa dados sensíveis
	EncryptData(data string) (string, error)

	// DecryptData descriptografa dados sensíveis
	DecryptData(encryptedData string) (string, error)
}
