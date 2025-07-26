package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// WhatsAppSessionStatus representa o status de uma sessão WhatsApp
type WhatsAppSessionStatus string

const (
	WhatsAppStatusDisconnected WhatsAppSessionStatus = "disconnected"
	WhatsAppStatusConnecting   WhatsAppSessionStatus = "connecting"
	WhatsAppStatusConnected    WhatsAppSessionStatus = "connected"
)

// Session representa uma sessão do WhatsApp
type Session struct {
	bun.BaseModel `bun:"table:zapcore_sessions,alias:s"`

	ID        uuid.UUID             `bun:"id,pk,type:uuid" json:"id"`
	Name      string                `bun:"name,type:varchar(100),notnull,unique" json:"name"`
	Status    WhatsAppSessionStatus `bun:"status,type:varchar(20),notnull" json:"status"`
	WaJID     string                `bun:"waJid,type:varchar(100)" json:"waJid,omitempty"`
	QRCode    string                `bun:"-" json:"qrCode,omitempty"`   // Não persistir no banco
	ProxyURL  string                `bun:"-" json:"proxyUrl,omitempty"` // Não persistir no banco
	Webhook   string                `bun:"-" json:"webhook,omitempty"`  // Não persistir no banco
	IsActive  bool                  `bun:"isActive,type:boolean" json:"isActive"`
	LastSeen  *time.Time            `bun:"lastSeen,type:timestamptz" json:"lastSeen,omitempty"`
	CreatedAt time.Time             `bun:"createdAt,type:timestamptz,notnull" json:"createdAt"`
	UpdatedAt time.Time             `bun:"updatedAt,type:timestamptz,notnull" json:"updatedAt"`
	Metadata  map[string]any        `bun:"-" json:"metadata,omitempty"` // Não persistir no banco por enquanto
}

// TableName retorna o nome da tabela para o Bun ORM
func (*Session) TableName() string {
	return "zapcore_sessions"
}

// IsConnected verifica se a sessão está conectada
func (s *Session) IsConnected() bool {
	return s.Status == WhatsAppStatusConnected
}

// CanConnect verifica se a sessão pode ser conectada
func (s *Session) CanConnect() bool {
	return s.Status == WhatsAppStatusDisconnected && s.IsActive
}

// SetConnecting define o status como conectando
func (s *Session) SetConnecting() {
	s.Status = WhatsAppStatusConnecting
	s.UpdatedAt = time.Now()
}

// SetConnected define o status como conectado
func (s *Session) SetConnected(waJID string) {
	s.Status = WhatsAppStatusConnected
	s.WaJID = waJID
	s.LastSeen = &time.Time{}
	*s.LastSeen = time.Now()
	s.UpdatedAt = time.Now()
}

// SetWaJID define o WhatsApp JID da sessão
func (s *Session) SetWaJID(waJID string) {
	s.WaJID = waJID
	s.UpdatedAt = time.Now()
}

// HasWaJID verifica se a sessão tem um WhatsApp JID
func (s *Session) HasWaJID() bool {
	return s.WaJID != ""
}

// IsAuthenticated verifica se a sessão está autenticada (tem WhatsApp JID)
func (s *Session) IsAuthenticated() bool {
	return s.HasWaJID()
}

// SetDisconnected define o status como desconectado
func (s *Session) SetDisconnected() {
	s.Status = WhatsAppStatusDisconnected
	s.UpdatedAt = time.Now()
}

// Deactivate desativa a sessão
func (s *Session) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}
