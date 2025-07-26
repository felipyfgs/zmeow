package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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
}

// NewRouter cria uma nova instância do router sem config (para compatibilidade)
func NewRouter(sessionHandler *handlers.SessionHandler, healthHandler *handlers.HealthHandler) *Router {
	log := logger.WithComponent("router")
	
	r := &Router{
		Mux:            chi.NewRouter(),
		logger:         log,
		sessionHandler: sessionHandler,
		healthHandler:  healthHandler,
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
) *Router {
	r := &Router{
		Mux:            chi.NewRouter(),
		config:         cfg,
		logger:         log.WithComponent("router"),
		sessionHandler: sessionHandler,
		healthHandler:  healthHandler,
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