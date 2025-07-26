package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger interface define os métodos disponíveis para logging
type Logger interface {
	// Métodos de logging por nível
	Trace() *zerolog.Event
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event

	// Métodos para adicionar contexto
	WithComponent(component string) Logger
	WithFields(fields map[string]interface{}) Logger
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger

	// Método para obter o zerolog.Logger subjacente
	GetZerolog() *zerolog.Logger
}

// ConfigProvider interface para configuração do logger
type ConfigProvider interface {
	GetLogLevel() string
	GetLogOutput() string
	GetLogConsoleFormat() string
	GetLogFileFormat() string
	GetLogFilePath() string
	GetLogFileMaxSize() int
	GetLogFileMaxBackups() int
	GetLogFileMaxAge() int
	GetLogFileCompress() bool
	GetLogConsoleColors() bool
}

// ZerologLogger implementa a interface Logger usando zerolog
type ZerologLogger struct {
	logger *zerolog.Logger
}

// NewZerologLogger cria uma nova instância do ZerologLogger
func NewZerologLogger(zl *zerolog.Logger) Logger {
	return &ZerologLogger{logger: zl}
}

// Implementação dos métodos de logging
func (l *ZerologLogger) Trace() *zerolog.Event {
	return l.logger.Trace()
}

func (l *ZerologLogger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

func (l *ZerologLogger) Info() *zerolog.Event {
	return l.logger.Info()
}

func (l *ZerologLogger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

func (l *ZerologLogger) Error() *zerolog.Event {
	return l.logger.Error()
}

func (l *ZerologLogger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

func (l *ZerologLogger) Panic() *zerolog.Event {
	return l.logger.Panic()
}

// Métodos para adicionar contexto
func (l *ZerologLogger) WithComponent(component string) Logger {
	newLogger := l.logger.With().Str("component", component).Logger()
	return NewZerologLogger(&newLogger)
}

func (l *ZerologLogger) WithFields(fields map[string]interface{}) Logger {
	ctx := l.logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	newLogger := ctx.Logger()
	return NewZerologLogger(&newLogger)
}

func (l *ZerologLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return NewZerologLogger(&newLogger)
}

func (l *ZerologLogger) WithError(err error) Logger {
	newLogger := l.logger.With().Err(err).Logger()
	return NewZerologLogger(&newLogger)
}

func (l *ZerologLogger) GetZerolog() *zerolog.Logger {
	return l.logger
}

// Setup configura o logger principal da aplicação
func Setup(cfg ConfigProvider) Logger {
	// Configurar nível de log
	level := parseLogLevel(cfg.GetLogLevel())
	zerolog.SetGlobalLevel(level)

	// Configurar writers baseado na configuração
	writers := setupWriters(cfg)

	// Criar logger com múltiplos writers
	var logger zerolog.Logger
	if len(writers) == 1 {
		logger = zerolog.New(writers[0])
	} else {
		logger = zerolog.New(io.MultiWriter(writers...))
	}

	// Adicionar timestamp e caller
	logger = logger.With().
		Timestamp().
		Caller().
		Logger()

	return NewZerologLogger(&logger)
}

// setupWriters configura os writers baseado na configuração
func setupWriters(cfg ConfigProvider) []io.Writer {
	var writers []io.Writer

	output := cfg.GetLogOutput()

	switch output {
	case "console":
		writers = append(writers, setupConsoleWriter(cfg))
	case "file":
		writers = append(writers, setupFileWriter(cfg))
	case "dual":
		writers = append(writers, setupConsoleWriter(cfg))
		writers = append(writers, setupFileWriter(cfg))
	case "stdout":
		writers = append(writers, os.Stdout)
	case "stderr":
		writers = append(writers, os.Stderr)
	default:
		// Default para dual
		writers = append(writers, setupConsoleWriter(cfg))
		writers = append(writers, setupFileWriter(cfg))
	}

	return writers
}

// setupConsoleWriter configura o writer para console
func setupConsoleWriter(cfg ConfigProvider) io.Writer {
	consoleFormat := cfg.GetLogConsoleFormat()
	useColors := cfg.GetLogConsoleColors()

	if consoleFormat == "json" {
		return os.Stdout
	}

	// Console formatado
	return zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    !useColors,
	}
}

// setupFileWriter configura o writer para arquivo com rotação
func setupFileWriter(cfg ConfigProvider) io.Writer {
	filePath := cfg.GetLogFilePath()

	// Criar diretório se não existir
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return os.Stdout
	}

	return &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.GetLogFileMaxSize(),
		MaxBackups: cfg.GetLogFileMaxBackups(),
		MaxAge:     cfg.GetLogFileMaxAge(),
		Compress:   cfg.GetLogFileCompress(),
	}
}

// parseLogLevel converte string para zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// Configurações pré-definidas

// SetupForDevelopment configura logger para desenvolvimento
func SetupForDevelopment() Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	fileWriter := &lumberjack.Logger{
		Filename:   "logs/dev.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
	}

	// Criar diretório se não existir
	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}

	logger := zerolog.New(io.MultiWriter(consoleWriter, fileWriter)).
		With().
		Timestamp().
		Caller().
		Logger()

	return NewZerologLogger(&logger)
}

// SetupForProduction configura logger para produção
func SetupForProduction() Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	fileWriter := &lumberjack.Logger{
		Filename:   "logs/zmeow.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	// Criar diretório se não existir
	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}

	logger := zerolog.New(io.MultiWriter(os.Stdout, fileWriter)).
		With().
		Timestamp().
		Caller().
		Logger()

	return NewZerologLogger(&logger)
}

// SetupForTesting configura logger para testes
func SetupForTesting() Logger {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	return NewZerologLogger(&logger)
}

// Context helpers

type contextKey string

const loggerKey contextKey = "logger"

// WithContext adiciona logger ao contexto
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extrai logger do contexto
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	// Retorna logger padrão se não encontrar no contexto
	return SetupForDevelopment()
}

// Helper functions para logging estruturado

// LogSessionEvent loga eventos de sessão com contexto estruturado
func LogSessionEvent(logger Logger, eventType, sessionID string, fields map[string]interface{}) {
	event := logger.Info().
		Str("event_type", eventType).
		Str("session_id", sessionID)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("Session event")
}

// LogWhatsAppEvent loga eventos do WhatsApp com contexto estruturado
func LogWhatsAppEvent(logger Logger, eventType, sessionID, jid string, fields map[string]interface{}) {
	event := logger.Info().
		Str("event_type", eventType).
		Str("session_id", sessionID).
		Str("jid", jid)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("WhatsApp event")
}

// LogRawEventPayload loga payload JSON bruto de eventos do WhatsApp no nível DEBUG
func LogRawEventPayload(logger Logger, sessionID, eventType string, payload interface{}) {
	logger.Debug().
		Str("session_id", sessionID).
		Str("event_type", eventType).
		Interface("payload", payload).
		Msgf("📱 WhatsApp Event [%s] - Raw Payload", eventType)
}

// LogFormattedEventPayload loga payload formatado com informações extraídas
func LogFormattedEventPayload(logger Logger, sessionID, eventType string, payload interface{}, extractedInfo map[string]interface{}) {
	event := logger.Info().
		Str("session_id", sessionID).
		Str("event_type", eventType).
		Interface("payload", payload)

	// Adicionar informações extraídas
	for key, value := range extractedInfo {
		event = event.Interface(key, value)
	}

	event.Msgf("📱 WhatsApp Event [%s] - Formatted", eventType)
}

// LogHTTPRequest loga requests HTTP com contexto estruturado
func LogHTTPRequest(logger Logger, method, path, userAgent, clientIP string, status, durationMs int, fields map[string]interface{}) {
	event := logger.Info().
		Str("method", method).
		Str("path", path).
		Str("user_agent", userAgent).
		Str("client_ip", clientIP).
		Int("status", status).
		Int("duration_ms", durationMs)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("HTTP request")
}

// LogError loga erros com contexto estruturado
func LogError(logger Logger, err error, operation string, fields map[string]interface{}) {
	event := logger.Error().
		Err(err).
		Str("operation", operation)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("Operation error")
}

// Funções globais de conveniência

// WithComponent cria um logger com componente específico
func WithComponent(component string) Logger {
	return SetupForDevelopment().WithComponent(component)
}

// Loggers contextuais pré-configurados

// NewContextualLogger cria um logger com contexto rico
func NewContextualLogger(app, env, module string) Logger {
	logger := SetupForDevelopment()
	return logger.WithFields(map[string]interface{}{
		"app":    app,
		"env":    env,
		"module": module,
	})
}

// NewWhatsAppSessionLogger cria um logger específico para sessões WhatsApp
func NewWhatsAppSessionLogger(sessionID, jid string) Logger {
	return NewContextualLogger("zmeow", "development", "whatsapp_session").
		WithFields(map[string]interface{}{
			"session_id": sessionID,
			"jid":        jid,
		})
}

// NewRequestLogger cria um logger específico para requests HTTP
func NewRequestLogger(method, route, userAgent, clientIP string) Logger {
	return NewContextualLogger("zmeow", "development", "http_router").
		WithFields(map[string]interface{}{
			"method":     method,
			"route":      route,
			"user_agent": userAgent,
			"client_ip":  clientIP,
		})
}

// NewQueryLogger cria um logger específico para operações de banco
func NewQueryLogger(operation, table string) Logger {
	return NewContextualLogger("zmeow", "development", "database").
		WithFields(map[string]interface{}{
			"operation": operation,
			"table":     table,
		})
}
