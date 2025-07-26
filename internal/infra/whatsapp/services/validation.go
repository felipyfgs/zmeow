package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// ValidationServiceImpl implementa o serviço de validação
type ValidationServiceImpl struct {
	logger logger.Logger
}

// NewValidationService cria uma nova instância do ValidationService
func NewValidationService(log logger.Logger) whatsapp.ValidationService {
	return &ValidationServiceImpl{
		logger: log.WithComponent("validation-service"),
	}
}

// ValidatePhoneNumber valida um número de telefone
func (vs *ValidationServiceImpl) ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// Remover espaços e caracteres especiais
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "+", "")

	// Verificar se contém apenas dígitos
	if !regexp.MustCompile(`^\d+$`).MatchString(cleanPhone) {
		return fmt.Errorf("phone number must contain only digits")
	}

	// Verificar comprimento (entre 10 e 15 dígitos)
	if len(cleanPhone) < 10 || len(cleanPhone) > 15 {
		return fmt.Errorf("phone number must be between 10 and 15 digits")
	}

	vs.logger.WithField("phone", phone).Debug().Msg("Phone number validated successfully")
	return nil
}

// ValidateWebhookURL valida uma URL de webhook
func (vs *ValidationServiceImpl) ValidateWebhookURL(webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	// Parse da URL
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL format: %w", err)
	}

	// Verificar se é HTTP ou HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTP or HTTPS protocol")
	}

	// Verificar se tem host
	if parsedURL.Host == "" {
		return fmt.Errorf("webhook URL must have a valid host")
	}

	// Recomendar HTTPS para produção
	if parsedURL.Scheme == "http" {
		vs.logger.WithField("url", webhookURL).Warn().Msg("HTTP webhook URL detected, HTTPS recommended for production")
	}

	vs.logger.WithField("url", webhookURL).Debug().Msg("Webhook URL validated successfully")
	return nil
}

// ValidateSessionName valida um nome de sessão
func (vs *ValidationServiceImpl) ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	// Verificar comprimento
	if len(name) < 3 {
		return fmt.Errorf("session name must be at least 3 characters long")
	}

	if len(name) > 100 {
		return fmt.Errorf("session name must be at most 100 characters long")
	}

	// Verificar caracteres válidos (letras, números, hífen, underscore)
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return fmt.Errorf("session name can only contain letters, numbers, hyphens, and underscores")
	}

	// Não pode começar ou terminar com hífen ou underscore
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, "_") ||
		strings.HasSuffix(name, "-") || strings.HasSuffix(name, "_") {
		return fmt.Errorf("session name cannot start or end with hyphen or underscore")
	}

	vs.logger.WithField("name", name).Debug().Msg("Session name validated successfully")
	return nil
}

// ValidateJID valida um JID do WhatsApp
func (vs *ValidationServiceImpl) ValidateJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("JID cannot be empty")
	}

	// Formato básico de JID: número@s.whatsapp.net ou número:device@s.whatsapp.net
	jidPattern := regexp.MustCompile(`^\d+(:?\d+)?@s\.whatsapp\.net$`)
	if !jidPattern.MatchString(jid) {
		return fmt.Errorf("invalid JID format, expected format: number@s.whatsapp.net or number:device@s.whatsapp.net")
	}

	// Extrair número da parte antes do @
	parts := strings.Split(jid, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid JID format")
	}

	numberPart := parts[0]
	
	// Se tem device ID, separar
	if strings.Contains(numberPart, ":") {
		deviceParts := strings.Split(numberPart, ":")
		if len(deviceParts) != 2 {
			return fmt.Errorf("invalid JID device format")
		}
		numberPart = deviceParts[0]
	}

	// Validar se o número tem comprimento adequado
	if len(numberPart) < 10 || len(numberPart) > 15 {
		return fmt.Errorf("invalid phone number in JID")
	}

	vs.logger.WithField("jid", jid).Debug().Msg("JID validated successfully")
	return nil
}
