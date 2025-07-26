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

	// Logging
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
