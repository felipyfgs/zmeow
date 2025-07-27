package session

import (
	"context"

	"github.com/google/uuid"
)

// SessionRepository define as operações de persistência para sessões
type SessionRepository interface {
	// Create cria uma nova sessão no banco de dados
	Create(ctx context.Context, session *Session) error

	// GetByID busca uma sessão pelo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)

	// GetByName busca uma sessão pelo nome
	GetByName(ctx context.Context, name string) (*Session, error)

	// List retorna todas as sessões ativas
	List(ctx context.Context) ([]*Session, error)

	// Update atualiza uma sessão existente
	Update(ctx context.Context, session *Session) error

	// Delete remove uma sessão do banco de dados
	Delete(ctx context.Context, id uuid.UUID) error

	// ListActive retorna todas as sessões ativas
	ListActive(ctx context.Context) ([]*Session, error)

	// ExistsByName verifica se uma sessão com o nome especificado já existe
	ExistsByName(ctx context.Context, name string) (bool, error)

	// UpdateJID atualiza o JID de uma sessão
	UpdateJID(ctx context.Context, id uuid.UUID, jid string) error

	// UpdateWhatsAppJID atualiza o WhatsApp JID de uma sessão
	UpdateWhatsAppJID(ctx context.Context, id uuid.UUID, whatsappJID string) error

	// GetSessionsWithWhatsAppJID retorna todas as sessões que têm WhatsApp JID
	GetSessionsWithWhatsAppJID(ctx context.Context) ([]*Session, error)

	// UpdateStatus atualiza apenas o status de uma sessão
	UpdateStatus(ctx context.Context, id uuid.UUID, status WhatsAppSessionStatus) error

	// UpdateLastSeen atualiza o último visto de uma sessão
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
}
