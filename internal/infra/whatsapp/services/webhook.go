package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"zmeow/pkg/logger"
)

// WebhookPayload representa o payload de um webhook
type WebhookPayload struct {
	SessionID uuid.UUID              `json:"sessionId"`
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WebhookConfig representa a configuração de webhook para uma sessão
type WebhookConfig struct {
	SessionID uuid.UUID `json:"sessionId"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret,omitempty"`
	Enabled   bool      `json:"enabled"`
	Retries   int       `json:"retries"`
	Timeout   int       `json:"timeout"` // em segundos
}

// WebhookServiceImpl implementa o serviço de webhooks
type WebhookServiceImpl struct {
	configs    map[uuid.UUID]*WebhookConfig
	httpClient *http.Client
	mutex      sync.RWMutex
	logger     logger.Logger
}

// NewWebhookService cria uma nova instância do WebhookService
func NewWebhookService(log logger.Logger) *WebhookServiceImpl {
	return &WebhookServiceImpl{
		configs: make(map[uuid.UUID]*WebhookConfig),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: log.WithComponent("webhook-service"),
	}
}

// SetWebhookConfig configura webhook para uma sessão
func (ws *WebhookServiceImpl) SetWebhookConfig(config *WebhookConfig) error {
	if config.SessionID == uuid.Nil {
		return fmt.Errorf("session ID cannot be empty")
	}

	if config.URL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	if config.Retries <= 0 {
		config.Retries = 3 // Padrão: 3 tentativas
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 // Padrão: 30 segundos
	}

	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	ws.configs[config.SessionID] = config

	ws.logger.WithFields(map[string]interface{}{
		"sessionId": config.SessionID,
		"url":       config.URL,
		"enabled":   config.Enabled,
	}).Info().Msg("Webhook configuration updated")

	return nil
}

// GetWebhookConfig retorna a configuração de webhook de uma sessão
func (ws *WebhookServiceImpl) GetWebhookConfig(sessionID uuid.UUID) (*WebhookConfig, error) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	config, exists := ws.configs[sessionID]
	if !exists {
		return nil, fmt.Errorf("webhook config not found for session %s", sessionID)
	}

	// Retornar cópia
	configCopy := *config
	return &configCopy, nil
}

// RemoveWebhookConfig remove a configuração de webhook de uma sessão
func (ws *WebhookServiceImpl) RemoveWebhookConfig(sessionID uuid.UUID) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	delete(ws.configs, sessionID)

	ws.logger.WithField("sessionId", sessionID).Info().Msg("Webhook configuration removed")
	return nil
}

// SendWebhook envia um webhook para uma sessão
func (ws *WebhookServiceImpl) SendWebhook(sessionID uuid.UUID, event string, data map[string]interface{}) error {
	ws.mutex.RLock()
	config, exists := ws.configs[sessionID]
	ws.mutex.RUnlock()

	if !exists {
		ws.logger.WithField("sessionId", sessionID).Debug().Msg("No webhook configured for session")
		return nil // Não é um erro se não há webhook configurado
	}

	if !config.Enabled {
		ws.logger.WithField("sessionId", sessionID).Debug().Msg("Webhook disabled for session")
		return nil
	}

	// Criar payload
	payload := &WebhookPayload{
		SessionID: sessionID,
		Event:     event,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Enviar webhook de forma assíncrona
	go ws.sendWebhookAsync(config, payload)

	return nil
}

// sendWebhookAsync envia webhook de forma assíncrona com retry
func (ws *WebhookServiceImpl) sendWebhookAsync(config *WebhookConfig, payload *WebhookPayload) {
	var lastErr error

	for attempt := 1; attempt <= config.Retries; attempt++ {
		err := ws.sendWebhookHTTP(config, payload)
		if err == nil {
			ws.logger.WithFields(map[string]interface{}{
				"sessionId": config.SessionID,
				"event":     payload.Event,
				"attempt":   attempt,
			}).Debug().Msg("Webhook sent successfully")
			return
		}

		lastErr = err
		ws.logger.WithError(err).WithFields(map[string]interface{}{
			"sessionId":  config.SessionID,
			"event":      payload.Event,
			"attempt":    attempt,
			"maxRetries": config.Retries,
		}).Warn().Msg("Webhook send failed, retrying")

		// Aguardar antes da próxima tentativa (backoff exponencial)
		if attempt < config.Retries {
			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	// Todas as tentativas falharam
	ws.logger.WithError(lastErr).WithFields(map[string]interface{}{
		"sessionId": config.SessionID,
		"event":     payload.Event,
		"attempts":  config.Retries,
	}).Error().Msg("Webhook send failed after all retries")
}

// sendWebhookHTTP envia o webhook via HTTP
func (ws *WebhookServiceImpl) sendWebhookHTTP(config *WebhookConfig, payload *WebhookPayload) error {
	// Serializar payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Criar request
	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Configurar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ZMeow-Webhook/1.0")

	// Adicionar assinatura se secret estiver configurado
	if config.Secret != "" {
		signature := ws.generateSignature(jsonData, config.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Configurar timeout específico
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Enviar request
	resp, err := ws.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// generateSignature gera assinatura HMAC para o webhook
func (ws *WebhookServiceImpl) generateSignature(_ []byte, _ string) string {
	// TODO: Implementar assinatura HMAC-SHA256
	// Por enquanto, retorna um placeholder
	return "sha256=placeholder"
}

// SendTestWebhook envia um webhook de teste
func (ws *WebhookServiceImpl) SendTestWebhook(sessionID uuid.UUID) error {
	testData := map[string]interface{}{
		"message": "This is a test webhook",
		"test":    true,
	}

	return ws.SendWebhook(sessionID, "test", testData)
}

// GetWebhookStats retorna estatísticas de webhooks
func (ws *WebhookServiceImpl) GetWebhookStats() map[string]interface{} {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	stats := map[string]interface{}{
		"totalConfigs":   len(ws.configs),
		"enabledConfigs": 0,
	}

	for _, config := range ws.configs {
		if config.Enabled {
			stats["enabledConfigs"] = stats["enabledConfigs"].(int) + 1
		}
	}

	return stats
}

// ValidateWebhookURL valida se uma URL de webhook é válida
func (ws *WebhookServiceImpl) ValidateWebhookURL(url string) error {
	if url == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	// Criar request de teste
	req, err := http.NewRequest("POST", url, bytes.NewBufferString("{}"))
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ZMeow-Webhook-Validator/1.0")

	// Configurar timeout curto para validação
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Tentar conectar (não enviar dados reais)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook URL validation failed: %w", err)
	}
	defer resp.Body.Close()

	// Aceitar qualquer status code para validação
	ws.logger.WithFields(map[string]interface{}{
		"url":    url,
		"status": resp.StatusCode,
	}).Debug().Msg("Webhook URL validation completed")

	return nil
}

// EnableWebhook habilita webhook para uma sessão
func (ws *WebhookServiceImpl) EnableWebhook(sessionID uuid.UUID) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	config, exists := ws.configs[sessionID]
	if !exists {
		return fmt.Errorf("webhook config not found for session %s", sessionID)
	}

	config.Enabled = true

	ws.logger.WithField("sessionId", sessionID).Info().Msg("Webhook enabled")
	return nil
}

// DisableWebhook desabilita webhook para uma sessão
func (ws *WebhookServiceImpl) DisableWebhook(sessionID uuid.UUID) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	config, exists := ws.configs[sessionID]
	if !exists {
		return fmt.Errorf("webhook config not found for session %s", sessionID)
	}

	config.Enabled = false

	ws.logger.WithField("sessionId", sessionID).Info().Msg("Webhook disabled")
	return nil
}

// Close encerra o WebhookService
func (ws *WebhookServiceImpl) Close() {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	count := len(ws.configs)
	ws.configs = make(map[uuid.UUID]*WebhookConfig)

	ws.logger.WithField("clearedCount", count).Info().Msg("WebhookService closed")
}
