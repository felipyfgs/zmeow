package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"zmeow/internal/domain/session"
)

// sessionRepository implementa a interface SessionRepository
type sessionRepository struct {
	db *bun.DB
}

// NewSessionRepository cria uma nova instância do repositório de sessões
func NewSessionRepository(db *bun.DB) session.SessionRepository {
	return &sessionRepository{db: db}
}

// Create cria uma nova sessão no banco de dados
func (r *sessionRepository) Create(ctx context.Context, sess *session.Session) error {
	sess.ID = uuid.New()
	sess.CreatedAt = time.Now()
	sess.UpdatedAt = time.Now()
	sess.Status = session.WhatsAppStatusDisconnected
	sess.IsActive = true

	_, err := r.db.NewInsert().Model(sess).Exec(ctx)
	return err
}

// GetByID busca uma sessão pelo ID
func (r *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	sess := new(session.Session)
	err := r.db.NewSelect().Model(sess).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, session.ErrSessionNotFound
		}
		return nil, err
	}
	return sess, nil
}

// GetByName busca uma sessão pelo nome
func (r *sessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	sess := new(session.Session)
	err := r.db.NewSelect().Model(sess).Where("name = ?", name).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, session.ErrSessionNotFound
		}
		return nil, err
	}
	return sess, nil
}

// List retorna todas as sessões
func (r *sessionRepository) List(ctx context.Context) ([]*session.Session, error) {
	var sessions []*session.Session
	err := r.db.NewSelect().Model(&sessions).Order("createdAt DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// Update atualiza uma sessão existente
func (r *sessionRepository) Update(ctx context.Context, sess *session.Session) error {
	sess.UpdatedAt = time.Now()

	_, err := r.db.NewUpdate().
		Model(sess).
		Where("id = ?", sess.ID).
		Exec(ctx)

	return err
}

// Delete remove uma sessão do banco de dados
func (r *sessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*session.Session)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// ListActive retorna todas as sessões ativas
func (r *sessionRepository) ListActive(ctx context.Context) ([]*session.Session, error) {
	var sessions []*session.Session
	err := r.db.NewSelect().
		Model(&sessions).
		Where("\"isActive\" = ?", true).
		Order("createdAt DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// ExistsByName verifica se uma sessão com o nome especificado já existe
func (r *sessionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*session.Session)(nil)).
		Where("name = ?", name).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpdateStatus atualiza apenas o status de uma sessão
func (r *sessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status session.WhatsAppSessionStatus) error {
	_, err := r.db.NewUpdate().
		Model((*session.Session)(nil)).
		Set("status = ?", status).
		Set("\"updatedAt\" = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// UpdateJID atualiza o WaJID de uma sessão (compatibilidade)
func (r *sessionRepository) UpdateJID(ctx context.Context, id uuid.UUID, jid string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*session.Session)(nil)).
		Set("\"waJid\" = ?", jid).
		Set("\"lastSeen\" = ?", now).
		Set("\"updatedAt\" = ?", now).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// UpdateWhatsAppJID atualiza o WhatsApp JID de uma sessão
func (r *sessionRepository) UpdateWhatsAppJID(ctx context.Context, id uuid.UUID, waJID string) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*session.Session)(nil)).
		Set("\"waJid\" = ?", waJID).
		Set("\"updatedAt\" = ?", now).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// GetSessionsWithWhatsAppJID retorna todas as sessões que têm WhatsApp JID
func (r *sessionRepository) GetSessionsWithWhatsAppJID(ctx context.Context) ([]*session.Session, error) {
	var sessions []*session.Session
	err := r.db.NewSelect().
		Model(&sessions).
		Where("\"waJid\" IS NOT NULL AND \"waJid\" != ''").
		Where("\"isActive\" = ?", true).
		Order("createdAt DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// UpdateLastSeen atualiza o último visto de uma sessão
func (r *sessionRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*session.Session)(nil)).
		Set("\"lastSeen\" = ?", now).
		Set("\"updatedAt\" = ?", now).
		Where("id = ?", id).
		Exec(ctx)

	return err
}
