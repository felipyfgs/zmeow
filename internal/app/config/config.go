package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Env  string
		Port string
		Host string
	}

	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}

	WhatsApp struct {
		DebugLevel  string
		StorePrefix string
	}

	Logging struct {
		// Configurações básicas
		Level          string
		Output         string
		ConsoleFormat  string
		FileFormat     string
		FilePath       string
		FileMaxSize    int
		FileMaxBackups int
		FileMaxAge     int
		FileCompress   bool
		ConsoleColors  bool

		// Configurações contextuais
		AppName     string
		Environment string
		Version     string
		ServiceName string

		// Configurações avançadas
		EnableCaller     bool
		EnableStackTrace bool
		EnableSampling   bool
		SampleRate       int
		EnableMetrics    bool
	}

	RateLimit struct {
		Requests int
		Window   time.Duration
	}

	CORS struct {
		AllowedOrigins string
	}
}

func LoadConfig() (*Config, error) {
	// Carregar .env se existir
	_ = godotenv.Load()

	cfg := &Config{}

	// App
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Port = getEnv("APP_PORT", "8080")
	cfg.App.Host = getEnv("APP_HOST", "0.0.0.0")

	// Database
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "zmeow")
	cfg.Database.Password = getEnv("DB_PASSWORD", "zmeow123")
	cfg.Database.Name = getEnv("DB_NAME", "zmeow")
	cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")

	// WhatsApp
	cfg.WhatsApp.DebugLevel = getEnv("WA_DEBUG_LEVEL", "INFO")
	cfg.WhatsApp.StorePrefix = getEnv("WA_STORE_PREFIX", "zmeow")

	// Logging - Configurações básicas
	cfg.Logging.Level = getEnv("LOG_LEVEL", "info")
	cfg.Logging.Output = getEnv("LOG_OUTPUT", "dual")
	cfg.Logging.ConsoleFormat = getEnv("LOG_CONSOLE_FORMAT", "console")
	cfg.Logging.FileFormat = getEnv("LOG_FILE_FORMAT", "json")
	cfg.Logging.FilePath = getEnv("LOG_FILE_PATH", "logs/zmeow.log")
	cfg.Logging.FileMaxSize = getEnvAsInt("LOG_FILE_MAX_SIZE", 100)
	cfg.Logging.FileMaxBackups = getEnvAsInt("LOG_FILE_MAX_BACKUPS", 3)
	cfg.Logging.FileMaxAge = getEnvAsInt("LOG_FILE_MAX_AGE", 28)
	cfg.Logging.FileCompress = getEnvAsBool("LOG_FILE_COMPRESS", true)
	cfg.Logging.ConsoleColors = getEnvAsBool("LOG_CONSOLE_COLORS", true)

	// Logging - Configurações contextuais
	cfg.Logging.AppName = getEnv("APP_NAME", "zmeow")
	cfg.Logging.Environment = getEnv("APP_ENV", "development")
	cfg.Logging.Version = getEnv("APP_VERSION", "1.0.0")
	cfg.Logging.ServiceName = getEnv("SERVICE_NAME", "whatsapp-api")

	// Logging - Configurações avançadas
	cfg.Logging.EnableCaller = getEnvAsBool("LOG_ENABLE_CALLER", true)
	cfg.Logging.EnableStackTrace = getEnvAsBool("LOG_ENABLE_STACK_TRACE", false)
	cfg.Logging.EnableSampling = getEnvAsBool("LOG_ENABLE_SAMPLING", false)
	cfg.Logging.SampleRate = getEnvAsInt("LOG_SAMPLE_RATE", 10)
	cfg.Logging.EnableMetrics = getEnvAsBool("LOG_ENABLE_METRICS", false)

	// Rate Limit
	cfg.RateLimit.Requests = getEnvAsInt("RATE_LIMIT_REQUESTS", 100)
	windowStr := getEnv("RATE_LIMIT_WINDOW", "1m")
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		window = 1 * time.Minute
	}
	cfg.RateLimit.Window = window

	// CORS
	cfg.CORS.AllowedOrigins = getEnv("CORS_ALLOWED_ORIGINS", "*")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func (c *Config) GetDatabaseDSN() string {
	return "postgres://" + c.Database.User + ":" + c.Database.Password +
		"@" + c.Database.Host + ":" + c.Database.Port +
		"/" + c.Database.Name + "?sslmode=" + c.Database.SSLMode
}

// Implementação da interface ConfigProvider para integração com o logger
// Configurações básicas
func (c *Config) GetLogLevel() string         { return c.Logging.Level }
func (c *Config) GetLogOutput() string        { return c.Logging.Output }
func (c *Config) GetLogConsoleFormat() string { return c.Logging.ConsoleFormat }
func (c *Config) GetLogFileFormat() string    { return c.Logging.FileFormat }
func (c *Config) GetLogFilePath() string      { return c.Logging.FilePath }
func (c *Config) GetLogFileMaxSize() int      { return c.Logging.FileMaxSize }
func (c *Config) GetLogFileMaxBackups() int   { return c.Logging.FileMaxBackups }
func (c *Config) GetLogFileMaxAge() int       { return c.Logging.FileMaxAge }
func (c *Config) GetLogFileCompress() bool    { return c.Logging.FileCompress }
func (c *Config) GetLogConsoleColors() bool   { return c.Logging.ConsoleColors }

// Configurações contextuais
func (c *Config) GetLogAppName() string     { return c.Logging.AppName }
func (c *Config) GetLogEnvironment() string { return c.Logging.Environment }
func (c *Config) GetLogVersion() string     { return c.Logging.Version }
func (c *Config) GetLogServiceName() string { return c.Logging.ServiceName }

// Configurações avançadas
func (c *Config) GetLogEnableCaller() bool     { return c.Logging.EnableCaller }
func (c *Config) GetLogEnableStackTrace() bool { return c.Logging.EnableStackTrace }
func (c *Config) GetLogEnableSampling() bool   { return c.Logging.EnableSampling }
func (c *Config) GetLogSampleRate() int        { return c.Logging.SampleRate }
func (c *Config) GetLogEnableMetrics() bool    { return c.Logging.EnableMetrics }

// Configurações por ambiente

// ApplyDevelopmentLoggingConfig aplica configurações de logging para desenvolvimento
func (c *Config) ApplyDevelopmentLoggingConfig() {
	c.Logging.Level = "debug"
	c.Logging.Environment = "development"
	c.Logging.ConsoleColors = true
	c.Logging.EnableCaller = true
	c.Logging.EnableStackTrace = true
	c.Logging.EnableSampling = false
	c.Logging.SampleRate = 10
	c.Logging.EnableMetrics = false
}

// ApplyProductionLoggingConfig aplica configurações de logging para produção
func (c *Config) ApplyProductionLoggingConfig() {
	c.Logging.Level = "info"
	c.Logging.Environment = "production"
	c.Logging.ConsoleColors = false
	c.Logging.EnableCaller = false
	c.Logging.EnableStackTrace = false
	c.Logging.EnableSampling = true
	c.Logging.SampleRate = 100
	c.Logging.EnableMetrics = false
}

// ApplyTestingLoggingConfig aplica configurações de logging para testes
func (c *Config) ApplyTestingLoggingConfig() {
	c.Logging.Level = "warn"
	c.Logging.Environment = "testing"
	c.Logging.Output = "stdout"
	c.Logging.ConsoleColors = false
	c.Logging.EnableCaller = false
	c.Logging.EnableStackTrace = false
	c.Logging.EnableSampling = false
	c.Logging.EnableMetrics = false
}

// ApplyStagingLoggingConfig aplica configurações de logging para staging
func (c *Config) ApplyStagingLoggingConfig() {
	c.Logging.Level = "debug"
	c.Logging.Environment = "staging"
	c.Logging.ConsoleColors = true
	c.Logging.EnableCaller = true
	c.Logging.EnableStackTrace = true
	c.Logging.EnableSampling = false
	c.Logging.SampleRate = 10
	c.Logging.EnableMetrics = false
}

// Funções de conveniência para setup rápido

// LoadConfigForDevelopment carrega configuração otimizada para desenvolvimento
func LoadConfigForDevelopment() (*Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	cfg.ApplyDevelopmentLoggingConfig()
	return cfg, nil
}

// LoadConfigForProduction carrega configuração otimizada para produção
func LoadConfigForProduction() (*Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	cfg.ApplyProductionLoggingConfig()
	return cfg, nil
}

// LoadConfigForTesting carrega configuração otimizada para testes
func LoadConfigForTesting() (*Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	cfg.ApplyTestingLoggingConfig()
	return cfg, nil
}

// LoadConfigForStaging carrega configuração otimizada para staging
func LoadConfigForStaging() (*Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	cfg.ApplyStagingLoggingConfig()
	return cfg, nil
}
