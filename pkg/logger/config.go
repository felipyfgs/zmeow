package logger

import (
	"os"
	"strconv"
	"strings"
)

// LogConfig configuração avançada para o logger
type LogConfig struct {
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

// DefaultConfig retorna configuração padrão
func DefaultConfig() *LogConfig {
	return &LogConfig{
		// Básicas
		Level:          "info",
		Output:         "dual",
		ConsoleFormat:  "console",
		FileFormat:     "json",
		FilePath:       "logs/zmeow.log",
		FileMaxSize:    100,
		FileMaxBackups: 3,
		FileMaxAge:     28,
		FileCompress:   true,
		ConsoleColors:  true,

		// Contextuais
		AppName:     "zmeow",
		Environment: "development",
		Version:     "1.0.0",
		ServiceName: "whatsapp-api",

		// Avançadas
		EnableCaller:     true,
		EnableStackTrace: false,
		EnableSampling:   false,
		SampleRate:       10,
		EnableMetrics:    false,
	}
}

// LoadFromEnv carrega configuração das variáveis de ambiente
func LoadFromEnv() *LogConfig {
	config := DefaultConfig()

	// Configurações básicas
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.Level = val
	}
	if val := os.Getenv("LOG_OUTPUT"); val != "" {
		config.Output = val
	}
	if val := os.Getenv("LOG_CONSOLE_FORMAT"); val != "" {
		config.ConsoleFormat = val
	}
	if val := os.Getenv("LOG_FILE_FORMAT"); val != "" {
		config.FileFormat = val
	}
	if val := os.Getenv("LOG_FILE_PATH"); val != "" {
		config.FilePath = val
	}
	if val := os.Getenv("LOG_FILE_MAX_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.FileMaxSize = size
		}
	}
	if val := os.Getenv("LOG_FILE_MAX_BACKUPS"); val != "" {
		if backups, err := strconv.Atoi(val); err == nil {
			config.FileMaxBackups = backups
		}
	}
	if val := os.Getenv("LOG_FILE_MAX_AGE"); val != "" {
		if age, err := strconv.Atoi(val); err == nil {
			config.FileMaxAge = age
		}
	}
	if val := os.Getenv("LOG_FILE_COMPRESS"); val != "" {
		config.FileCompress = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("LOG_CONSOLE_COLORS"); val != "" {
		config.ConsoleColors = strings.ToLower(val) == "true"
	}

	// Configurações contextuais
	if val := os.Getenv("APP_NAME"); val != "" {
		config.AppName = val
	}
	if val := os.Getenv("APP_ENV"); val != "" {
		config.Environment = val
	}
	if val := os.Getenv("APP_VERSION"); val != "" {
		config.Version = val
	}
	if val := os.Getenv("SERVICE_NAME"); val != "" {
		config.ServiceName = val
	}

	// Configurações avançadas
	if val := os.Getenv("LOG_ENABLE_CALLER"); val != "" {
		config.EnableCaller = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("LOG_ENABLE_STACK_TRACE"); val != "" {
		config.EnableStackTrace = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("LOG_ENABLE_SAMPLING"); val != "" {
		config.EnableSampling = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("LOG_SAMPLE_RATE"); val != "" {
		if rate, err := strconv.Atoi(val); err == nil {
			config.SampleRate = rate
		}
	}
	if val := os.Getenv("LOG_ENABLE_METRICS"); val != "" {
		config.EnableMetrics = strings.ToLower(val) == "true"
	}

	return config
}

// Implementação da interface ConfigProvider
func (c *LogConfig) GetLogLevel() string         { return c.Level }
func (c *LogConfig) GetLogOutput() string        { return c.Output }
func (c *LogConfig) GetLogConsoleFormat() string { return c.ConsoleFormat }
func (c *LogConfig) GetLogFileFormat() string    { return c.FileFormat }
func (c *LogConfig) GetLogFilePath() string      { return c.FilePath }
func (c *LogConfig) GetLogFileMaxSize() int      { return c.FileMaxSize }
func (c *LogConfig) GetLogFileMaxBackups() int   { return c.FileMaxBackups }
func (c *LogConfig) GetLogFileMaxAge() int       { return c.FileMaxAge }
func (c *LogConfig) GetLogFileCompress() bool    { return c.FileCompress }
func (c *LogConfig) GetLogConsoleColors() bool   { return c.ConsoleColors }

// SetupWithConfig configura logger com configuração customizada
func SetupWithConfig(config *LogConfig) Logger {
	logger := Setup(config)

	// Adicionar contexto global
	return logger.WithFields(map[string]interface{}{
		"app":     config.AppName,
		"env":     config.Environment,
		"version": config.Version,
		"service": config.ServiceName,
	})
}

// Configurações pré-definidas para diferentes ambientes

// DevelopmentConfig configuração para desenvolvimento
func DevelopmentConfig() *LogConfig {
	config := DefaultConfig()
	config.Level = "debug"
	config.Environment = "development"
	config.ConsoleColors = true
	config.EnableCaller = true
	config.EnableStackTrace = true
	return config
}

// ProductionConfig configuração para produção
func ProductionConfig() *LogConfig {
	config := DefaultConfig()
	config.Level = "info"
	config.Environment = "production"
	config.ConsoleColors = false
	config.EnableCaller = false
	config.EnableStackTrace = false
	config.EnableSampling = true
	config.SampleRate = 100
	return config
}

// TestingConfig configuração para testes
func TestingConfig() *LogConfig {
	config := DefaultConfig()
	config.Level = "warn"
	config.Environment = "testing"
	config.Output = "stdout"
	config.ConsoleColors = false
	config.EnableCaller = false
	return config
}

// StagingConfig configuração para staging
func StagingConfig() *LogConfig {
	config := DefaultConfig()
	config.Level = "debug"
	config.Environment = "staging"
	config.EnableCaller = true
	config.EnableStackTrace = true
	return config
}

// Funções de conveniência para setup rápido

// SetupForDev configura logger para desenvolvimento
func SetupForDev() Logger {
	return SetupWithConfig(DevelopmentConfig())
}

// SetupForProd configura logger para produção
func SetupForProd() Logger {
	return SetupWithConfig(ProductionConfig())
}

// SetupForTest configura logger para testes
func SetupForTest() Logger {
	return SetupWithConfig(TestingConfig())
}

// SetupForStaging configura logger para staging
func SetupForStaging() Logger {
	return SetupWithConfig(StagingConfig())
}

// SetupFromEnv configura logger a partir das variáveis de ambiente
func SetupFromEnv() Logger {
	return SetupWithConfig(LoadFromEnv())
}
