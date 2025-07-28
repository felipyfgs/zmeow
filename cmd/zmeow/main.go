// Package main ZMeow API
// @title           ZMeow WhatsApp API
// @version         1.0
// @description     API completa para integração com WhatsApp usando Go + Chi router. Permite gerenciar sessões, enviar mensagens, gerenciar grupos e muito mais.
// @termsOfService  http://swagger.io/terms/

// @contact.name   ZMeow API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@zmeow.com

// @license.name  MIT
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"zmeow/internal/app"
	"zmeow/internal/app/config"
	"zmeow/internal/app/server"
	"zmeow/internal/http/router"
	"zmeow/internal/infra/database"
	"zmeow/internal/infra/whatsapp/core"
	"zmeow/pkg/logger"
)

func main() {
	// Carregar configuração
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Configurar logger usando as configurações do .env
	log := logger.Setup(cfg).WithComponent("main")

	log.WithFields(map[string]interface{}{
		"env":  cfg.App.Env,
		"port": cfg.App.Port,
	}).Info().Msg("Starting ZMeow API")

	// Conectar ao banco de dados
	dsn := cfg.GetDatabaseDSN()

	db, err := database.NewDatabase(dsn, cfg.App.Env == "development", log)
	if err != nil {
		log.WithError(err).Fatal().Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Connected to database successfully")

	// Executar migrações
	if err := database.RunMigrations(db); err != nil {
		log.WithError(err).Fatal().Msg("Failed to run migrations")
	}

	// Inicializar WhatsApp Manager
	whatsappManager, err := core.NewManager(db, cfg, log)
	if err != nil {
		log.WithError(err).Fatal().Msg("Failed to initialize WhatsApp manager")
	}
	defer whatsappManager.Close()

	log.Info().Msg("WhatsApp manager initialized successfully")

	// Restaurar sessões
	if err := whatsappManager.RestoreSessions(context.Background()); err != nil {
		log.WithError(err).Error().Msg("Failed to restore sessions")
	}

	// Reconectar sessões restauradas (com melhorias de segurança)
	whatsappManager.ConnectRestoredSessions(context.Background())

	// Inicializar container de dependências
	container, err := app.NewContainer(db, whatsappManager)
	if err != nil {
		log.WithError(err).Fatal().Msg("Failed to initialize container")
	}

	// Configurar router com handlers
	handler := router.NewRouter(container.SessionHandler, container.HealthHandler, container.MessageHandler, container.ChatHandler, container.GroupHandler)

	// Criar servidor
	srv := server.New(cfg, handler, log)

	// Canal para capturar sinais do sistema
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Iniciar servidor em goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.WithError(err).Fatal().Msg("Failed to start server")
		}
	}()

	log.Info().Msg("ZMeow API started successfully")

	// Aguardar sinal de parada
	<-stop

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.WithError(err).Error().Msg("Error during server shutdown")
	}

	log.Info().Msg("Application stopped")
}
