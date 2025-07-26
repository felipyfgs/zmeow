package logger

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

// WhatsAppLoggerInterface define a interface esperada pelo whatsmeow
type WhatsAppLoggerInterface interface {
	Errorf(string, ...interface{})
	Warnf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
	Sub(string) WhatsAppLoggerInterface
}

// WhatsAppLoggerAdapter adapta nosso Logger para whatsmeow
type WhatsAppLoggerAdapter struct{ logger Logger }

// NewWhatsAppLoggerAdapter cria adaptador para whatsmeow
func NewWhatsAppLoggerAdapter(logger Logger) WhatsAppLoggerInterface {
	return &WhatsAppLoggerAdapter{logger}
}

// Implementação da interface de logger do whatsmeow
func (w *WhatsAppLoggerAdapter) Errorf(msg string, args ...interface{}) {
	w.logger.Error().Msgf(msg, args...)
}
func (w *WhatsAppLoggerAdapter) Warnf(msg string, args ...interface{}) {
	w.logger.Warn().Msgf(msg, args...)
}
func (w *WhatsAppLoggerAdapter) Infof(msg string, args ...interface{}) {
	w.logger.Info().Msgf(msg, args...)
}
func (w *WhatsAppLoggerAdapter) Debugf(msg string, args ...interface{}) {
	w.logger.Debug().Msgf(msg, args...)
}
func (w *WhatsAppLoggerAdapter) Sub(module string) WhatsAppLoggerInterface {
	return NewWhatsAppLoggerAdapter(w.logger.WithComponent(module))
}

// BunQueryHook implementa hook para logging de queries do Bun ORM
type BunQueryHook struct {
	logger Logger
}

// NewBunQueryHook cria um novo hook para logging de queries do Bun
func NewBunQueryHook(logger Logger) bun.QueryHook {
	return &BunQueryHook{
		logger: logger.WithComponent("database"),
	}
}

// BeforeQuery é chamado antes da execução da query
func (h *BunQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

// AfterQuery é chamado após a execução da query
func (h *BunQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	duration := time.Since(event.StartTime)

	if event.Err != nil {
		// Erros sempre são logados com detalhes completos
		h.logger.Error().
			Err(event.Err).
			Str("query", h.sanitizeQuery(event.Query)).
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Str("operation", h.getQueryOperation(event.Query)).
			Str("table", h.getQueryTable(event.Query)).
			Msg("Database query failed")
	} else {
		// Queries de sucesso com logging inteligente
		h.logSuccessfulQuery(event.Query, duration)
	}
}

// logSuccessfulQuery aplica logging inteligente baseado no tipo e duração da query
func (h *BunQueryHook) logSuccessfulQuery(query string, duration time.Duration) {
	operation := h.getQueryOperation(query)
	table := h.getQueryTable(query)
	durationMs := duration.Milliseconds()

	// Queries muito rápidas (< 10ms) só logam em TRACE
	if durationMs < 10 && h.isRoutineQuery(query) {
		h.logger.Trace().
			Str("operation", operation).
			Str("table", table).
			Int64("duration_ms", durationMs).
			Msg("Fast DB operation")
		return
	}

	// Queries lentas (> 100ms) sempre logam como WARNING
	if durationMs > 100 {
		h.logger.Warn().
			Str("operation", operation).
			Str("table", table).
			Str("query", h.sanitizeQuery(query)).
			Int64("duration_ms", durationMs).
			Msg("Slow database query")
		return
	}

	// Queries normais logam em DEBUG com informações resumidas
	h.logger.Debug().
		Str("operation", operation).
		Str("table", table).
		Int64("duration_ms", durationMs).
		Msg("DB operation completed")
}

// isRoutineQuery verifica se é uma query rotineira (UPDATE de lastSeen, etc.)
func (h *BunQueryHook) isRoutineQuery(query string) bool {
	routinePatterns := []string{
		"SET \"lastSeen\"",
		"SET lastSeen",
		"SET status =",
		"SET \"updatedAt\"",
		"SET updatedAt",
	}

	queryLower := strings.ToLower(query)
	for _, pattern := range routinePatterns {
		if strings.Contains(queryLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// getQueryOperation extrai o tipo de operação da query
func (h *BunQueryHook) getQueryOperation(query string) string {
	query = strings.TrimSpace(strings.ToUpper(query))

	if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(query, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(query, "ALTER") {
		return "ALTER"
	} else if strings.HasPrefix(query, "DROP") {
		return "DROP"
	}
	return "UNKNOWN"
}

// getQueryTable extrai o nome da tabela da query
func (h *BunQueryHook) getQueryTable(query string) string {
	queryUpper := strings.ToUpper(query)

	// Padrões para diferentes tipos de query
	patterns := []struct {
		operation string
		regex     string
	}{
		{"UPDATE", `UPDATE\s+"?([^"\s]+)"?`},
		{"INSERT", `INSERT\s+INTO\s+"?([^"\s]+)"?`},
		{"DELETE", `DELETE\s+FROM\s+"?([^"\s]+)"?`},
		{"SELECT", `FROM\s+"?([^"\s]+)"?`},
		{"CREATE", `CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?"?([^"\s]+)"?`},
	}

	for _, pattern := range patterns {
		if strings.Contains(queryUpper, pattern.operation) {
			// Implementação simples sem regex para evitar dependência
			return h.extractTableNameSimple(queryUpper, pattern.operation)
		}
	}

	return "unknown"
}

// extractTableNameSimple extrai nome da tabela de forma simples
func (h *BunQueryHook) extractTableNameSimple(query, operation string) string {
	var startKeyword string

	switch operation {
	case "UPDATE":
		startKeyword = "UPDATE"
	case "INSERT":
		startKeyword = "INTO"
	case "DELETE":
		startKeyword = "FROM"
	case "SELECT":
		startKeyword = "FROM"
	case "CREATE":
		startKeyword = "TABLE"
	default:
		return "unknown"
	}

	// Encontrar a posição da palavra-chave
	keywordPos := strings.Index(query, startKeyword)
	if keywordPos == -1 {
		return "unknown"
	}

	// Pegar o texto após a palavra-chave
	afterKeyword := strings.TrimSpace(query[keywordPos+len(startKeyword):])

	// Para CREATE TABLE, pular "IF NOT EXISTS"
	if operation == "CREATE" && strings.HasPrefix(afterKeyword, "IF NOT EXISTS") {
		afterKeyword = strings.TrimSpace(afterKeyword[13:])
	}

	// Pegar a primeira palavra (nome da tabela)
	parts := strings.Fields(afterKeyword)
	if len(parts) > 0 {
		tableName := parts[0]
		// Remover aspas se existirem
		tableName = strings.Trim(tableName, `"`)
		return strings.ToLower(tableName)
	}

	return "unknown"
}

// sanitizeQuery remove dados sensíveis e encurta a query para logging
func (h *BunQueryHook) sanitizeQuery(query string) string {
	// Limitar tamanho da query
	maxLength := 200
	if len(query) > maxLength {
		query = query[:maxLength] + "..."
	}

	// Remover quebras de linha e espaços extras
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")

	// Normalizar espaços
	for strings.Contains(query, "  ") {
		query = strings.ReplaceAll(query, "  ", " ")
	}

	return strings.TrimSpace(query)
}

// DatabaseLogger é um logger específico para operações de banco de dados
type DatabaseLogger struct {
	logger Logger
}

// NewDatabaseLogger cria um novo logger para operações de banco
func NewDatabaseLogger(logger Logger) *DatabaseLogger {
	return &DatabaseLogger{
		logger: logger.WithComponent("database"),
	}
}

// LogQuery loga uma query de banco de dados
func (d *DatabaseLogger) LogQuery(query string, duration time.Duration, err error) {
	if err != nil {
		d.logger.Error().
			Err(err).
			Str("query", query).
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Database query failed")
	} else {
		d.logger.Debug().
			Str("query", query).
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Database query completed")
	}
}

// LogConnection loga eventos de conexão com o banco
func (d *DatabaseLogger) LogConnection(event string, details map[string]interface{}) {
	logEvent := d.logger.Info().Str("event", event)

	for key, value := range details {
		logEvent = logEvent.Interface(key, value)
	}

	logEvent.Msg("Database connection event")
}

// LogTransaction loga eventos de transação
func (d *DatabaseLogger) LogTransaction(event string, duration time.Duration, err error) {
	if err != nil {
		d.logger.Error().
			Err(err).
			Str("event", event).
			Dur("duration", duration).
			Msg("Database transaction failed")
	} else {
		d.logger.Debug().
			Str("event", event).
			Dur("duration", duration).
			Msg("Database transaction completed")
	}
}

// HTTPLogger é um logger específico para requests HTTP
type HTTPLogger struct {
	logger Logger
}

// NewHTTPLogger cria um novo logger para HTTP
func NewHTTPLogger(logger Logger) *HTTPLogger {
	return &HTTPLogger{
		logger: logger.WithComponent("http"),
	}
}

// LogRequest loga um request HTTP
func (h *HTTPLogger) LogRequest(method, path, userAgent, clientIP string, status int, duration time.Duration, fields map[string]interface{}) {
	level := h.getLogLevel(status, duration)

	event := level.
		Str("method", method).
		Str("path", path).
		Str("user_agent", userAgent).
		Str("client_ip", clientIP).
		Int("status", status).
		Dur("duration", duration).
		Int64("duration_ms", duration.Milliseconds())

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg(h.getLogMessage(status, duration))
}

// getLogLevel determina o nível de log baseado no status e duração
func (h *HTTPLogger) getLogLevel(status int, duration time.Duration) *zerolog.Event {
	if status >= 500 {
		return h.logger.Error()
	} else if status >= 400 {
		return h.logger.Warn()
	} else if duration > 500*time.Millisecond {
		return h.logger.Warn()
	} else {
		return h.logger.Debug()
	}
}

// getLogMessage determina a mensagem de log baseada no status e duração
func (h *HTTPLogger) getLogMessage(status int, duration time.Duration) string {
	if status >= 500 {
		return "HTTP server error"
	} else if status >= 400 {
		return "HTTP client error"
	} else if duration > 500*time.Millisecond {
		return "HTTP slow request"
	} else {
		return "HTTP request"
	}
}

// WhatsAppLogger é um logger específico para eventos do WhatsApp
type WhatsAppLogger struct {
	logger Logger
}

// NewWhatsAppLogger cria um novo logger para WhatsApp
func NewWhatsAppLogger(logger Logger) *WhatsAppLogger {
	return &WhatsAppLogger{
		logger: logger.WithComponent("whatsapp"),
	}
}

// LogEvent loga um evento do WhatsApp
func (w *WhatsAppLogger) LogEvent(eventType, sessionID string, fields map[string]interface{}) {
	event := w.logger.Info().
		Str("event_type", eventType).
		Str("session_id", sessionID)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("WhatsApp event")
}

// LogRawPayload loga payload bruto de evento do WhatsApp
func (w *WhatsAppLogger) LogRawPayload(sessionID, eventType string, payload interface{}) {
	w.logger.Trace().
		Str("session_id", sessionID).
		Str("event_type", eventType).
		Interface("payload", payload).
		Msg("Raw WhatsApp Event Payload")
}

// LogConnection loga eventos de conexão do WhatsApp
func (w *WhatsAppLogger) LogConnection(sessionID, event string, details map[string]interface{}) {
	logEvent := w.logger.Info().
		Str("session_id", sessionID).
		Str("event", event)

	for key, value := range details {
		logEvent = logEvent.Interface(key, value)
	}

	logEvent.Msg("WhatsApp connection event")
}

// LogError loga erros do WhatsApp
func (w *WhatsAppLogger) LogError(sessionID string, err error, context map[string]interface{}) {
	logEvent := w.logger.Error().
		Err(err).
		Str("session_id", sessionID)

	for key, value := range context {
		logEvent = logEvent.Interface(key, value)
	}

	logEvent.Msg("WhatsApp error")
}

// SessionLogger é um logger específico para sessões
type SessionLogger struct {
	logger    Logger
	sessionID string
}

// NewSessionLogger cria um novo logger para uma sessão específica
func NewSessionLogger(logger Logger, sessionID string) *SessionLogger {
	return &SessionLogger{
		logger:    logger.WithComponent("session").WithFields(map[string]interface{}{"session_id": sessionID}),
		sessionID: sessionID,
	}
}

// LogEvent loga um evento da sessão
func (s *SessionLogger) LogEvent(eventType string, fields map[string]interface{}) {
	event := s.logger.Info().Str("event_type", eventType)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("Session event")
}

// LogError loga um erro da sessão
func (s *SessionLogger) LogError(err error, operation string, fields map[string]interface{}) {
	event := s.logger.Error().
		Err(err).
		Str("operation", operation)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("Session error")
}

// LogStateChange loga mudança de estado da sessão
func (s *SessionLogger) LogStateChange(fromState, toState string, reason string) {
	s.logger.Info().
		Str("from_state", fromState).
		Str("to_state", toState).
		Str("reason", reason).
		Msg("Session state changed")
}
