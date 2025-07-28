package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "zmeow/docs" // Swagger docs
	"zmeow/internal/app/config"
	"zmeow/internal/http/handlers"
	appMiddleware "zmeow/internal/http/middleware"
	"zmeow/pkg/logger"
)

// Router representa o roteador principal da aplicação
type Router struct {
	*chi.Mux
	config         *config.Config
	logger         logger.Logger
	sessionHandler *handlers.SessionHandler
	healthHandler  *handlers.HealthHandler
	messageHandler *handlers.MessageHandler
	chatHandler    *handlers.ChatHandler
	groupHandler   *handlers.GroupHandler
}

// NewRouter cria uma nova instância do router sem config (para compatibilidade)
func NewRouter(sessionHandler *handlers.SessionHandler, healthHandler *handlers.HealthHandler, messageHandler *handlers.MessageHandler, chatHandler *handlers.ChatHandler, groupHandler *handlers.GroupHandler) *Router {
	log := logger.WithComponent("router")

	r := &Router{
		Mux:            chi.NewRouter(),
		logger:         log,
		sessionHandler: sessionHandler,
		healthHandler:  healthHandler,
		messageHandler: messageHandler,
		chatHandler:    chatHandler,
		groupHandler:   groupHandler,
	}

	r.setupMiddlewares()
	r.setupRoutes()

	return r
}

// New cria uma nova instância do router
func New(
	cfg *config.Config,
	log logger.Logger,
	sessionHandler *handlers.SessionHandler,
	healthHandler *handlers.HealthHandler,
	messageHandler *handlers.MessageHandler,
	chatHandler *handlers.ChatHandler,
	groupHandler *handlers.GroupHandler,
) *Router {
	r := &Router{
		Mux:            chi.NewRouter(),
		config:         cfg,
		logger:         log.WithComponent("router"),
		sessionHandler: sessionHandler,
		healthHandler:  healthHandler,
		messageHandler: messageHandler,
		chatHandler:    chatHandler,
		groupHandler:   groupHandler,
	}

	r.setupMiddlewares()
	r.setupRoutes()

	return r
}

// setupMiddlewares configura os middlewares globais
func (r *Router) setupMiddlewares() {
	// Middleware básicos do Chi
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Timeout global
	r.Use(middleware.Timeout(60 * time.Second))

	// Middlewares customizados
	r.Use(appMiddleware.NewCORS())
	r.Use(appMiddleware.NewLoggingMiddleware(r.logger))
	r.Use(appMiddleware.NewRecoveryMiddleware(r.logger))
	r.Use(appMiddleware.NewRateLimit(100))
}

// setupRoutes configura as rotas da aplicação
func (r *Router) setupRoutes() {
	// Swagger documentation
	r.Get("/swagger/doc.json", r.swaggerDocHandler)
	r.Get("/swagger/*", httpSwagger.Handler())

	// Health check
	r.Get("/health", r.healthHandler.Health)

	// Rotas de sessões (sem prefixo api/v1)
	r.Route("/sessions", func(rt chi.Router) {
		rt.Post("/add", r.sessionHandler.AddSession)
		rt.Get("/list", r.sessionHandler.ListSessions)

		// Rotas que requerem sessionID
		rt.Route("/{sessionID}", func(rt chi.Router) {
			rt.Get("/", r.sessionHandler.GetSession)
			rt.Delete("/", r.sessionHandler.DeleteSession)
			rt.Post("/connect", r.sessionHandler.ConnectSession)
			rt.Post("/logout", r.sessionHandler.LogoutSession)
			rt.Get("/status", r.sessionHandler.GetSessionStatus)
			rt.Get("/qr", r.sessionHandler.GetQRCode)
			rt.Post("/pairphone", r.sessionHandler.PairPhone)
			rt.Post("/proxy/set", r.sessionHandler.SetProxy)
		})
	})

	// Rotas de mensagens
	r.Route("/messages", func(rt chi.Router) {
		// Rotas que requerem sessionID
		rt.Route("/{sessionID}", func(rt chi.Router) {
			// Rotas de envio
			rt.Route("/send", func(rt chi.Router) {
				rt.Post("/text", r.messageHandler.SendTextMessage)
				rt.Post("/media", r.messageHandler.SendMediaMessage)
				rt.Post("/image", r.messageHandler.SendImageMessage)
				rt.Post("/audio", r.messageHandler.SendAudioMessage)
				rt.Post("/video", r.messageHandler.SendVideoMessage)
				rt.Post("/document", r.messageHandler.SendDocumentMessage)
				rt.Post("/location", r.messageHandler.SendLocationMessage)
				rt.Post("/contact", r.messageHandler.SendContactMessage)
				rt.Post("/sticker", r.messageHandler.SendStickerMessage)
				rt.Post("/buttons", r.messageHandler.SendButtonsMessage)
				rt.Post("/list", r.messageHandler.SendListMessage)
				rt.Post("/poll", r.messageHandler.SendPollMessage)
				rt.Post("/edit", r.messageHandler.EditMessage)
			})

			// Outras operações de mensagem
			rt.Post("/delete", r.messageHandler.DeleteMessage)
			rt.Post("/react", r.messageHandler.ReactMessage)

			// TODO: Implementar histórico de mensagens
			// rt.Get("/", r.messageHandler.GetMessageHistory)
		})
	})

	// Rotas de chat (funcionalidades específicas de gerenciamento de chat)
	r.Route("/chat", func(rt chi.Router) {
		// Rotas que requerem sessionID
		rt.Route("/{sessionID}", func(rt chi.Router) {
			// Operações específicas de chat (não duplicadas)
			rt.Post("/presence", r.chatHandler.SendChatPresence)
			rt.Post("/markread", r.chatHandler.MarkAsRead)

			// Downloads de mídia
			rt.Post("/downloadimage", r.chatHandler.DownloadImage)
			rt.Post("/downloadvideo", r.chatHandler.DownloadVideo)
			rt.Post("/downloadaudio", r.chatHandler.DownloadAudio)
			rt.Post("/downloaddocument", r.chatHandler.DownloadDocument)
		})
	})

	// Rotas de grupos
	r.Route("/groups", func(rt chi.Router) {
		// Rotas que requerem sessionID
		rt.Route("/{sessionID}", func(rt chi.Router) {
			// Operações básicas de grupos
			rt.Post("/create", r.groupHandler.CreateGroup)
			rt.Get("/list", r.groupHandler.ListGroups)
			rt.Get("/info", r.groupHandler.GetGroupInfo)

			// Gerenciamento de participantes
			rt.Post("/leave", r.groupHandler.LeaveGroup)
			rt.Post("/participants/update", r.groupHandler.UpdateParticipants)

			// Configurações do grupo
			rt.Post("/settings/name", r.groupHandler.SetGroupName)
			rt.Post("/settings/topic", r.groupHandler.SetGroupTopic)
			rt.Post("/settings/photo", r.groupHandler.SetGroupPhoto)
			rt.Delete("/settings/photo", r.groupHandler.RemoveGroupPhoto)
			rt.Post("/settings/announce", r.groupHandler.SetGroupAnnounce)
			rt.Post("/settings/locked", r.groupHandler.SetGroupLocked)
			rt.Post("/settings/disappearing", r.groupHandler.SetDisappearingTimer)

			// Convites de grupo
			rt.Get("/invite/link", r.groupHandler.GetGroupInviteLink)
			rt.Post("/invite/join", r.groupHandler.JoinGroupWithLink)
			rt.Post("/invite/info", r.groupHandler.GetGroupInviteInfo)
		})
	})

	// Rota catch-all para 404
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte(`{
			"success": false,
			"message": "Endpoint não encontrado",
			"error": {
				"code": "NOT_FOUND",
				"details": "O endpoint solicitado não existe"
			}
		}`))
	})
}

// swaggerDocHandler serve o JSON do Swagger
func (r *Router) swaggerDocHandler(w http.ResponseWriter, req *http.Request) {
	// Serve o arquivo swagger.json diretamente
	http.ServeFile(w, req, "docs/swagger.json")
}
