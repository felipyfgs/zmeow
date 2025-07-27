package logger

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

// ============================================================================
// WHATSAPP ADAPTER
// ============================================================================

// WhatsAppLoggerInterface define a interface esperada pelo whatsmeow
type WhatsAppLoggerInterface interface {
	Errorf(string, ...any)
	Warnf(string, ...any)
	Infof(string, ...any)
	Debugf(string, ...any)
	Sub(string) WhatsAppLoggerInterface
}

// WhatsAppLoggerAdapter adapta nosso Logger para whatsmeow (otimizado)
type WhatsAppLoggerAdapter struct {
	logger Logger
}

// NewWhatsAppLoggerAdapter cria adaptador para whatsmeow
func NewWhatsAppLoggerAdapter(logger Logger) WhatsAppLoggerInterface {
	return &WhatsAppLoggerAdapter{logger: logger}
}

// Implementação da interface de logger do whatsmeow (otimizada)
func (w *WhatsAppLoggerAdapter) Errorf(msg string, args ...any) {
	if len(args) == 0 {
		w.logger.Error().Msg(msg)
	} else {
		w.logger.Error().Msgf(msg, args...)
	}
}

func (w *WhatsAppLoggerAdapter) Warnf(msg string, args ...any) {
	if len(args) == 0 {
		w.logger.Warn().Msg(msg)
	} else {
		w.logger.Warn().Msgf(msg, args...)
	}
}

func (w *WhatsAppLoggerAdapter) Infof(msg string, args ...any) {
	if len(args) == 0 {
		w.logger.Info().Msg(msg)
	} else {
		w.logger.Info().Msgf(msg, args...)
	}
}

func (w *WhatsAppLoggerAdapter) Debugf(msg string, args ...any) {
	if len(args) == 0 {
		w.logger.Debug().Msg(msg)
	} else {
		w.logger.Debug().Msgf(msg, args...)
	}
}

func (w *WhatsAppLoggerAdapter) Sub(module string) WhatsAppLoggerInterface {
	if module == "" {
		return w
	}
	return &WhatsAppLoggerAdapter{logger: w.logger.WithComponent(module)}
}

// ============================================================================
// BUN ORM ADAPTER
// ============================================================================

// BunQueryHook implementa hook para logging de queries do Bun ORM (otimizado)
type BunQueryHook struct {
	logger Logger
}

// NewBunQueryHook cria um novo hook para logging de queries do Bun
func NewBunQueryHook(logger Logger) bun.QueryHook {
	return &BunQueryHook{
		logger: logger.WithComponent("database"),
	}
}

// BeforeQuery é chamado antes da execução da query (otimizado)
func (h *BunQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	// Não fazer nada aqui para melhor performance
	return ctx
}

// AfterQuery é chamado após a execução da query (otimizado)
func (h *BunQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	duration := time.Since(event.StartTime)
	durationMs := duration.Milliseconds()

	if event.Err != nil {
		// Erros sempre são logados com detalhes completos
		h.logger.Error().
			Err(event.Err).
			Str("query", h.sanitizeQuery(event.Query)).
			Dur("duration", duration).
			Int64("duration_ms", durationMs).
			Str("operation", h.getQueryOperation(event.Query)).
			Str("table", h.getQueryTable(event.Query)).
			Msg("Database query failed")
		return
	}

	// Queries de sucesso com logging inteligente (otimizado)
	h.logSuccessfulQuery(event.Query, duration, durationMs)
}

// logSuccessfulQuery aplica logging inteligente baseado no tipo e duração da query (otimizado)
func (h *BunQueryHook) logSuccessfulQuery(query string, duration time.Duration, durationMs int64) {
	operation := h.getQueryOperation(query)
	table := h.getQueryTable(query)

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

// sanitizeQuery remove dados sensíveis e encurta a query para logging (otimizado)
func (h *BunQueryHook) sanitizeQuery(query string) string {
	if query == "" {
		return ""
	}

	// Limitar tamanho da query primeiro para melhor performance
	const maxLength = 200
	if len(query) > maxLength {
		query = query[:maxLength] + "..."
	}

	// Usar strings.Builder para melhor performance
	var builder strings.Builder
	builder.Grow(len(query)) // Pre-allocate capacity

	// Processar caractere por caractere para normalizar espaços
	var lastWasSpace bool
	for _, r := range query {
		switch r {
		case '\n', '\t', '\r':
			if !lastWasSpace {
				builder.WriteByte(' ')
				lastWasSpace = true
			}
		case ' ':
			if !lastWasSpace {
				builder.WriteByte(' ')
				lastWasSpace = true
			}
		default:
			builder.WriteRune(r)
			lastWasSpace = false
		}
	}

	return strings.TrimSpace(builder.String())
}

// ============================================================================
// DATABASE LOGGER (DEPRECATED - Use BunQueryHook instead)
// ============================================================================

// DatabaseLogger é um logger específico para operações de banco de dados
// DEPRECATED: Use BunQueryHook para melhor integração com Bun ORM
type DatabaseLogger struct {
	logger Logger
}

// NewDatabaseLogger cria um novo logger para operações de banco
// DEPRECATED: Use NewBunQueryHook para melhor integração com Bun ORM
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

// ============================================================================
// WHATSAPP DOMAIN LOGGER (DEPRECATED - Use WhatsAppLoggerAdapter instead)
// ============================================================================

// WhatsAppLogger é um logger específico para eventos do WhatsApp
// DEPRECATED: Use WhatsAppLoggerAdapter para melhor integração com whatsmeow
type WhatsAppLogger struct {
	logger Logger
}

// NewWhatsAppLogger cria um novo logger para WhatsApp
// DEPRECATED: Use NewWhatsAppLoggerAdapter para melhor integração com whatsmeow
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

// ============================================================================
// SESSION LOGGER (DEPRECATED - Use Logger.WithFields instead)
// ============================================================================

// SessionLogger é um logger específico para sessões
// DEPRECATED: Use logger.WithFields(map[string]any{"session_id": sessionID}) instead
type SessionLogger struct {
	logger    Logger
	sessionID string
}

// NewSessionLogger cria um novo logger para uma sessão específica
// DEPRECATED: Use logger.WithFields(map[string]any{"session_id": sessionID}) instead
func NewSessionLogger(logger Logger, sessionID string) *SessionLogger {
	return &SessionLogger{
		logger:    logger.WithComponent("session").WithFields(map[string]any{"session_id": sessionID}),
		sessionID: sessionID,
	}
}

// LogEvent loga um evento da sessão (otimizado)
func (s *SessionLogger) LogEvent(eventType string, fields map[string]any) {
	if eventType == "" {
		return
	}

	event := s.logger.Info().Str("event_type", eventType)

	// Otimização: verificar se há campos antes de iterar
	if len(fields) > 0 {
		for key, value := range fields {
			if key != "" && value != nil {
				event = event.Interface(key, value)
			}
		}
	}

	event.Msg("Session event")
}

// LogError loga um erro da sessão (otimizado)
func (s *SessionLogger) LogError(err error, operation string, fields map[string]any) {
	if err == nil {
		return
	}

	event := s.logger.Error().Err(err)

	if operation != "" {
		event = event.Str("operation", operation)
	}

	// Otimização: verificar se há campos antes de iterar
	if len(fields) > 0 {
		for key, value := range fields {
			if key != "" && value != nil {
				event = event.Interface(key, value)
			}
		}
	}

	event.Msg("Session error")
}

// LogStateChange loga mudança de estado da sessão (otimizado)
func (s *SessionLogger) LogStateChange(fromState, toState string, reason string) {
	if fromState == "" && toState == "" {
		return
	}

	event := s.logger.Info()

	if fromState != "" {
		event = event.Str("from_state", fromState)
	}
	if toState != "" {
		event = event.Str("to_state", toState)
	}
	if reason != "" {
		event = event.Str("reason", reason)
	}

	event.Msg("Session state changed")
}
