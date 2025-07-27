package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ============================================================================
// INTERFACES
// ============================================================================

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
	WithFields(fields map[string]any) Logger
	WithField(key string, value any) Logger
	WithError(err error) Logger

	// Método para obter o zerolog.Logger subjacente
	GetZerolog() *zerolog.Logger
}

// ConfigProvider interface para configuração do logger
type ConfigProvider interface {
	// Configurações básicas
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

	// Configurações contextuais
	GetLogAppName() string
	GetLogEnvironment() string
	GetLogVersion() string
	GetLogServiceName() string

	// Configurações avançadas
	GetLogEnableCaller() bool
	GetLogEnableStackTrace() bool
	GetLogEnableSampling() bool
	GetLogSampleRate() int
	GetLogEnableMetrics() bool
}

// ============================================================================
// CORE IMPLEMENTATION
// ============================================================================

// ZerologLogger implementa a interface Logger usando zerolog
type ZerologLogger struct {
	logger *zerolog.Logger
}

// NewZerologLogger cria uma nova instância do ZerologLogger
func NewZerologLogger(zl *zerolog.Logger) Logger {
	return &ZerologLogger{logger: zl}
}

// Implementação dos métodos de logging por nível
func (l *ZerologLogger) Trace() *zerolog.Event { return l.logger.Trace() }
func (l *ZerologLogger) Debug() *zerolog.Event { return l.logger.Debug() }
func (l *ZerologLogger) Info() *zerolog.Event  { return l.logger.Info() }
func (l *ZerologLogger) Warn() *zerolog.Event  { return l.logger.Warn() }
func (l *ZerologLogger) Error() *zerolog.Event { return l.logger.Error() }
func (l *ZerologLogger) Fatal() *zerolog.Event { return l.logger.Fatal() }
func (l *ZerologLogger) Panic() *zerolog.Event { return l.logger.Panic() }

// Métodos para adicionar contexto (otimizados)
func (l *ZerologLogger) WithComponent(component string) Logger {
	if component == "" {
		return l
	}
	newLogger := l.logger.With().Str("component", component).Logger()
	return &ZerologLogger{logger: &newLogger}
}

func (l *ZerologLogger) WithFields(fields map[string]any) Logger {
	if len(fields) == 0 {
		return l
	}
	ctx := l.logger.With()
	for key, value := range fields {
		if value != nil {
			ctx = ctx.Interface(key, value)
		}
	}
	newLogger := ctx.Logger()
	return &ZerologLogger{logger: &newLogger}
}

func (l *ZerologLogger) WithField(key string, value any) Logger {
	if key == "" || value == nil {
		return l
	}
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &ZerologLogger{logger: &newLogger}
}

func (l *ZerologLogger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	newLogger := l.logger.With().Err(err).Logger()
	return &ZerologLogger{logger: &newLogger}
}

func (l *ZerologLogger) GetZerolog() *zerolog.Logger {
	return l.logger
}

// ============================================================================
// LOGGER SETUP & CONFIGURATION
// ============================================================================

// Setup configura o logger principal da aplicação de forma otimizada
func Setup(cfg ConfigProvider) Logger {
	// Configurar nível de log globalmente
	if level := parseLogLevel(cfg.GetLogLevel()); level != zerolog.NoLevel {
		zerolog.SetGlobalLevel(level)
	}

	// Configurar writers baseado na configuração
	writers := setupWriters(cfg)
	if len(writers) == 0 {
		writers = []io.Writer{os.Stdout} // fallback
	}

	// Criar logger com múltiplos writers (otimizado)
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = io.MultiWriter(writers...)
	}

	// Configurar contexto base de forma eficiente
	logger := zerolog.New(output).With().Timestamp()

	// Adicionar caller se habilitado
	if cfg.GetLogEnableCaller() {
		logger = logger.Caller()
	}

	// Adicionar contexto da aplicação (apenas campos não vazios)
	if appName := cfg.GetLogAppName(); appName != "" {
		logger = logger.Str("app", appName)
	}
	if env := cfg.GetLogEnvironment(); env != "" {
		logger = logger.Str("env", env)
	}
	if version := cfg.GetLogVersion(); version != "" {
		logger = logger.Str("version", version)
	}
	if service := cfg.GetLogServiceName(); service != "" {
		logger = logger.Str("service", service)
	}

	finalLogger := logger.Logger()
	return &ZerologLogger{logger: &finalLogger}
}

// setupWriters configura os writers baseado na configuração (otimizado)
func setupWriters(cfg ConfigProvider) []io.Writer {
	output := cfg.GetLogOutput()

	switch output {
	case "console":
		return []io.Writer{setupConsoleWriter(cfg)}
	case "file":
		return []io.Writer{setupFileWriter(cfg)}
	case "dual":
		return []io.Writer{
			setupConsoleWriter(cfg),
			setupFileWriter(cfg),
		}
	case "stdout":
		return []io.Writer{os.Stdout}
	case "stderr":
		return []io.Writer{os.Stderr}
	default:
		// Default para dual
		return []io.Writer{
			setupConsoleWriter(cfg),
			setupFileWriter(cfg),
		}
	}
}

// setupConsoleWriter configura o writer para console
func setupConsoleWriter(cfg ConfigProvider) io.Writer {
	consoleFormat := cfg.GetLogConsoleFormat()
	useColors := cfg.GetLogConsoleColors()

	if consoleFormat == "json" {
		return &PrettyJSONWriter{Writer: os.Stdout}
	}

	// Console formatado com processamento de JSON
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    !useColors,
	}

	return &PrettyConsoleWriter{Writer: consoleWriter}
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

	fileWriter := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.GetLogFileMaxSize(),
		MaxBackups: cfg.GetLogFileMaxBackups(),
		MaxAge:     cfg.GetLogFileMaxAge(),
		Compress:   cfg.GetLogFileCompress(),
	}

	// Se o formato do arquivo for JSON, sempre usar pretty print
	if cfg.GetLogFileFormat() == "json" {
		return &PrettyJSONWriter{Writer: fileWriter}
	}

	return fileWriter
}

// ============================================================================
// JSON PROCESSORS & WRITERS
// ============================================================================

// JSONProcessor centraliza o processamento de JSON para melhor performance
type JSONProcessor struct{}

// Singleton para reutilização
var jsonProcessor = &JSONProcessor{}

// processJSONForFile processa JSON para arquivo com indentação
func (jp *JSONProcessor) processJSONForFile(p []byte) []byte {
	// Tentar fazer parse do JSON
	var jsonObj map[string]any
	if err := json.Unmarshal(p, &jsonObj); err != nil {
		// Se não for JSON válido, retornar como está
		return p
	}

	// Processar campos que podem conter JSON aninhado
	jp.processNestedJSON(jsonObj)

	// Formatar com indentação
	prettyJSON, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		// Se falhar na formatação, retornar como está
		return p
	}

	// Adicionar quebra de linha no final
	return append(prettyJSON, '\n')
}

// processJSONForConsole processa JSON para console
func (jp *JSONProcessor) processJSONForConsole(p []byte) []byte {
	// Tentar fazer parse do JSON
	var jsonObj map[string]any
	if err := json.Unmarshal(p, &jsonObj); err != nil {
		// Se não for JSON válido, retornar como está
		return p
	}

	// Processar campos que podem conter JSON aninhado
	jp.processNestedJSON(jsonObj)

	// Reformatar o JSON processado
	processedJSON, err := json.Marshal(jsonObj)
	if err != nil {
		// Se falhar na formatação, retornar como está
		return p
	}

	return processedJSON
}

// processNestedJSON processa campos que podem conter JSON aninhado (método centralizado)
func (jp *JSONProcessor) processNestedJSON(obj map[string]any) {
	for key, value := range obj {
		switch v := value.(type) {
		case string:
			// Primeiro verificar se contém "raw=" (prioritário para mensagens do WhatsApp)
			if key == "message" && strings.Contains(v, "raw=") {
				// Processar mensagens que contêm "raw={...}" e extrair para campo separado
				if rawData := extractRawFromMessage(v); rawData != nil {
					// Criar campo separado para o JSON extraído
					obj["raw_data"] = rawData
					// Manter apenas a parte da mensagem antes do "raw="
					if prefix := getMessagePrefix(v); prefix != "" {
						obj[key] = prefix
					}
				}
			} else if isJSONString(v) {
				// Verificar se o campo contém JSON aninhado
				if parsed := parseJSONString(v); parsed != nil {
					obj[key] = parsed
				}
			}
		case map[string]any:
			// Recursivamente processar objetos aninhados
			jp.processNestedJSON(v)
		case []any:
			// Processar arrays
			for _, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					jp.processNestedJSON(itemMap)
				}
			}
		}
	}
}

// PrettyJSONWriter é um wrapper que formata JSON de forma legível (otimizado)
type PrettyJSONWriter struct {
	Writer io.Writer
}

// PrettyConsoleWriter é um wrapper que processa JSON antes de enviar para o console
type PrettyConsoleWriter struct {
	Writer io.Writer
}

// Write implementa io.Writer formatando JSON com indentação (otimizado)
func (w *PrettyJSONWriter) Write(p []byte) (n int, err error) {
	// Usar o processor centralizado para melhor performance
	processed := jsonProcessor.processJSONForFile(p)

	// Escrever o resultado processado
	_, err = w.Writer.Write(processed)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Write implementa io.Writer para PrettyConsoleWriter (otimizado)
func (w *PrettyConsoleWriter) Write(p []byte) (n int, err error) {
	// Usar o processor centralizado para melhor performance
	processed := jsonProcessor.processJSONForConsole(p)

	// Escrever o resultado processado
	_, err = w.Writer.Write(processed)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// ============================================================================
// UTILITY FUNCTIONS (CONSOLIDATED)
// ============================================================================

// isJSONString verifica se uma string contém JSON válido
func isJSONString(s string) bool {
	// Verificar se começa e termina com { } ou [ ]
	trimmed := strings.TrimSpace(s)
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

// parseJSONString tenta fazer parse de uma string JSON
func parseJSONString(s string) interface{} {
	// Primeiro tentar como objeto
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(s), &obj); err == nil {
		processNestedJSONGlobal(obj) // Processar recursivamente
		return obj
	}

	// Depois tentar como array
	var arr []interface{}
	if err := json.Unmarshal([]byte(s), &arr); err == nil {
		return arr
	}

	// Se não conseguir fazer parse, retornar nil
	return nil
}

// processNestedJSONGlobal processa campos que podem conter JSON aninhado (função global)
func processNestedJSONGlobal(obj map[string]interface{}) {
	for key, value := range obj {
		switch v := value.(type) {
		case string:
			// Verificar se o campo contém JSON aninhado
			if isJSONString(v) {
				if parsed := parseJSONString(v); parsed != nil {
					obj[key] = parsed
				}
			} else if strings.Contains(v, "raw=") {
				// Processar mensagens que contêm "raw={...}" e extrair para campo separado
				if rawData := extractRawFromMessage(v); rawData != nil {
					// Criar campo separado para o JSON extraído
					obj["raw_data"] = rawData
					// Manter apenas a parte da mensagem antes do "raw="
					if prefix := getMessagePrefix(v); prefix != "" {
						obj[key] = prefix
					}
				}
			}
		case map[string]interface{}:
			// Recursivamente processar objetos aninhados
			processNestedJSONGlobal(v)
		case []interface{}:
			// Processar arrays
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					processNestedJSONGlobal(itemMap)
				}
			}
		}
	}
}

// extractBalanced extrai JSON balanceado de uma string
func extractBalanced(s string, open, close rune) string {
	if len(s) == 0 || rune(s[0]) != open {
		return ""
	}

	balance := 0
	inString := false
	escaped := false

	for i, r := range s {
		if escaped {
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if r == '"' {
			inString = !inString
			continue
		}

		if !inString {
			switch r {
			case open:
				balance++
			case close:
				balance--
				if balance == 0 {
					return s[:i+1]
				}
			}
		}
	}

	return ""
}

// extractRawFromMessage extrai apenas o JSON do campo "raw=" e retorna como objeto (otimizado)
func extractRawFromMessage(message string) interface{} {
	// Procurar por "raw=" na mensagem
	rawIndex := strings.Index(message, "raw=")
	if rawIndex == -1 {
		return nil
	}

	// Extrair a parte após "raw="
	jsonPart := strings.TrimSpace(message[rawIndex+4:])

	if len(jsonPart) == 0 {
		return nil
	}

	// Detectar se é um objeto ou array JSON
	var jsonStr string
	switch jsonPart[0] {
	case '{':
		jsonStr = extractBalanced(jsonPart, '{', '}')
	case '[':
		jsonStr = extractBalanced(jsonPart, '[', ']')
	default:
		return nil
	}

	if jsonStr == "" {
		return nil
	}

	// Tentar fazer parse do JSON
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		// Se falhar, tentar como string simples
		return jsonStr
	}

	// Se for um objeto, processar recursivamente
	if objMap, ok := obj.(map[string]interface{}); ok {
		processNestedJSONGlobal(objMap)
		return objMap
	}

	return obj
}

// getMessagePrefix retorna a parte da mensagem antes do "raw=" (otimizado)
func getMessagePrefix(message string) string {
	rawIndex := strings.Index(message, "raw=")
	if rawIndex == -1 {
		return message
	}

	// Extrair e limpar a parte antes do "raw="
	prefix := strings.TrimSpace(message[:rawIndex])
	if prefix == "" {
		return ""
	}

	// Remover espaços extras e normalizar
	prefix = strings.ReplaceAll(prefix, "\n", " ")
	prefix = strings.ReplaceAll(prefix, "\t", " ")

	// Normalizar espaços múltiplos
	for strings.Contains(prefix, "  ") {
		prefix = strings.ReplaceAll(prefix, "  ", " ")
	}

	return strings.TrimSpace(prefix)
}

// fallbackConfig implementa ConfigProvider para compatibilidade com funções deprecated
type fallbackConfig struct {
	level            string
	output           string
	consoleFormat    string
	fileFormat       string
	filePath         string
	fileMaxSize      int
	fileMaxBackups   int
	fileMaxAge       int
	fileCompress     bool
	consoleColors    bool
	appName          string
	environment      string
	version          string
	serviceName      string
	enableCaller     bool
	enableStackTrace bool
	enableSampling   bool
	sampleRate       int
	enableMetrics    bool
}

// Implementação da interface ConfigProvider para fallbackConfig
func (f *fallbackConfig) GetLogLevel() string          { return f.level }
func (f *fallbackConfig) GetLogOutput() string         { return f.output }
func (f *fallbackConfig) GetLogConsoleFormat() string  { return f.consoleFormat }
func (f *fallbackConfig) GetLogFileFormat() string     { return f.fileFormat }
func (f *fallbackConfig) GetLogFilePath() string       { return f.filePath }
func (f *fallbackConfig) GetLogFileMaxSize() int       { return f.fileMaxSize }
func (f *fallbackConfig) GetLogFileMaxBackups() int    { return f.fileMaxBackups }
func (f *fallbackConfig) GetLogFileMaxAge() int        { return f.fileMaxAge }
func (f *fallbackConfig) GetLogFileCompress() bool     { return f.fileCompress }
func (f *fallbackConfig) GetLogConsoleColors() bool    { return f.consoleColors }
func (f *fallbackConfig) GetLogAppName() string        { return f.appName }
func (f *fallbackConfig) GetLogEnvironment() string    { return f.environment }
func (f *fallbackConfig) GetLogVersion() string        { return f.version }
func (f *fallbackConfig) GetLogServiceName() string    { return f.serviceName }
func (f *fallbackConfig) GetLogEnableCaller() bool     { return f.enableCaller }
func (f *fallbackConfig) GetLogEnableStackTrace() bool { return f.enableStackTrace }
func (f *fallbackConfig) GetLogEnableSampling() bool   { return f.enableSampling }
func (f *fallbackConfig) GetLogSampleRate() int        { return f.sampleRate }
func (f *fallbackConfig) GetLogEnableMetrics() bool    { return f.enableMetrics }

// parseLogLevel converte string para zerolog.Level (otimizado)
func parseLogLevel(level string) zerolog.Level {
	if level == "" {
		return zerolog.InfoLevel
	}

	// Usar switch sem conversão para melhor performance
	switch level {
	case "trace", "TRACE":
		return zerolog.TraceLevel
	case "debug", "DEBUG":
		return zerolog.DebugLevel
	case "info", "INFO":
		return zerolog.InfoLevel
	case "warn", "WARN", "warning", "WARNING":
		return zerolog.WarnLevel
	case "error", "ERROR":
		return zerolog.ErrorLevel
	case "fatal", "FATAL":
		return zerolog.FatalLevel
	case "panic", "PANIC":
		return zerolog.PanicLevel
	case "disabled", "DISABLED":
		return zerolog.Disabled
	default:
		// Fallback com conversão apenas se necessário
		switch strings.ToLower(level) {
		case "trace":
			return zerolog.TraceLevel
		case "debug":
			return zerolog.DebugLevel
		case "warn", "warning":
			return zerolog.WarnLevel
		case "error":
			return zerolog.ErrorLevel
		case "fatal":
			return zerolog.FatalLevel
		case "panic":
			return zerolog.PanicLevel
		case "disabled":
			return zerolog.Disabled
		default:
			return zerolog.InfoLevel
		}
	}
}

// Configurações pré-definidas (DEPRECATED - Use config.LoadConfigFor* functions)

// SetupForDevelopment configura logger para desenvolvimento
// DEPRECATED: Use config.LoadConfigForDevelopment() + Setup(cfg)
func SetupForDevelopment() Logger {
	// Fallback para compatibilidade
	cfg := &fallbackConfig{
		level:            "debug",
		output:           "dual",
		consoleFormat:    "console",
		fileFormat:       "json",
		filePath:         "logs/zmeow.log",
		fileMaxSize:      100,
		fileMaxBackups:   3,
		fileMaxAge:       28,
		fileCompress:     true,
		consoleColors:    true,
		appName:          "zmeow",
		environment:      "development",
		version:          "1.0.0",
		serviceName:      "whatsapp-api",
		enableCaller:     true,
		enableStackTrace: true,
		enableSampling:   false,
		sampleRate:       10,
		enableMetrics:    false,
	}
	return Setup(cfg)
}

// SetupForProduction configura logger para produção
// DEPRECATED: Use config.LoadConfigForProduction() + Setup(cfg)
func SetupForProduction() Logger {
	// Fallback para compatibilidade
	cfg := &fallbackConfig{
		level:            "info",
		output:           "dual",
		consoleFormat:    "console",
		fileFormat:       "json",
		filePath:         "logs/zmeow.log",
		fileMaxSize:      100,
		fileMaxBackups:   3,
		fileMaxAge:       28,
		fileCompress:     true,
		consoleColors:    false,
		appName:          "zmeow",
		environment:      "production",
		version:          "1.0.0",
		serviceName:      "whatsapp-api",
		enableCaller:     false,
		enableStackTrace: false,
		enableSampling:   true,
		sampleRate:       100,
		enableMetrics:    false,
	}
	return Setup(cfg)
}

// SetupForTesting configura logger para testes
// DEPRECATED: Use config.LoadConfigForTesting() + Setup(cfg)
func SetupForTesting() Logger {
	// Fallback para compatibilidade
	cfg := &fallbackConfig{
		level:            "warn",
		output:           "stdout",
		consoleFormat:    "console",
		fileFormat:       "json",
		filePath:         "logs/test.log",
		fileMaxSize:      10,
		fileMaxBackups:   1,
		fileMaxAge:       1,
		fileCompress:     false,
		consoleColors:    false,
		appName:          "zmeow",
		environment:      "testing",
		version:          "1.0.0",
		serviceName:      "whatsapp-api",
		enableCaller:     false,
		enableStackTrace: false,
		enableSampling:   false,
		sampleRate:       10,
		enableMetrics:    false,
	}
	return Setup(cfg)
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
// DEPRECATED: Use config.LoadConfig() + Setup(cfg) + WithFields()
func NewContextualLogger(app, env, module string) Logger {
	logger := SetupForDevelopment()
	return logger.WithFields(map[string]interface{}{
		"module": module,
	})
}

// NewWhatsAppSessionLogger cria um logger específico para sessões WhatsApp
// DEPRECATED: Use config.LoadConfig() + Setup(cfg) + WithFields()
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
