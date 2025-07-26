package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"zmeow/internal/domain/session"
	"zmeow/pkg/logger"
)

// NewDatabase cria uma nova conexão com o banco de dados PostgreSQL
func NewDatabase(dsn string, debug bool, log logger.Logger) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())

	// Habilitar logging de queries se necessário
	if debug {
		db.AddQueryHook(logger.NewBunQueryHook(log))
	}

	// Configurar pool de conexões
	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(25)

	// Testar conexão
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// RunMigrations executa as migrações do banco de dados
func RunMigrations(db *bun.DB) error {
	// Criar tabela de sessões se não existir
	_, err := db.NewCreateTable().
		Model((*session.Session)(nil)).
		IfNotExists().
		Exec(context.Background())

	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	return nil
}
