package message

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"zmeow/internal/domain/message"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// SendMediaMessageUseCase implementa o caso de uso para envio de mídia
type SendMediaMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
	numberValidator *NumberValidator
}

// NewSendMediaMessageUseCase cria uma nova instância do caso de uso
func NewSendMediaMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendMediaMessageUseCase {
	return &SendMediaMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
		numberValidator: NewNumberValidator(),
	}
}

// Execute executa o caso de uso para enviar mídia
func (uc *SendMediaMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendMediaMessageRequest) (*message.SendMessageResponse, error) {
	// Obter destinatário
	destination := uc.numberValidator.GetDestination(req.Number, req.GroupJid)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"number":      req.Number,
		"groupJid":    req.GroupJid,
		"destination": destination,
		"mediaType":   req.MediaType,
		"caption":     req.Caption,
		"fileName":    req.FileName,
		"mimeType":    req.MimeType,
	}).Info().Msg("Sending media message")

	// Validar entrada
	if err := uc.validateRequest(req); err != nil {
		uc.logger.WithError(err).Error().Msg("Invalid request")
		return nil, err
	}

	// Verificar se a sessão existe
	_, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get session")
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Verificar se a sessão está conectada usando o WhatsApp Manager
	if !uc.whatsappManager.IsConnected(sessionID) {
		uc.logger.WithField("sessionId", sessionID).Warn().Msg("Session is not connected")
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	// Verificar se é URL ou dados Base64
	var mediaData []byte
	isURL := strings.HasPrefix(req.Media, "http://") || strings.HasPrefix(req.Media, "https://")

	if !isURL {
		// Decodificar dados da mídia (Base64 ou data URL)
		var err error
		mediaData, err = uc.decodeMediaData(req.Media)
		if err != nil {
			uc.logger.WithError(err).Error().Msg("Failed to decode media data")
			return nil, fmt.Errorf("failed to decode media data: %w", err)
		}
	}

	// Determinar MIME type se não fornecido
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = uc.detectMimeType(req.MediaType, req.FileName)
	}

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client")
		return nil, fmt.Errorf("failed to get WhatsApp client: %w", err)
	}

	// Enviar mídia
	var messageID string
	if isURL {
		// Para URLs, usar método específico que baixa e envia
		messageID, err = client.SendMediaFromURL(ctx, sessionID, destination, req.MediaType, req.Media, req.Caption, req.FileName, mimeType)
	} else {
		// Para dados Base64, usar método tradicional
		messageID, err = client.SendMediaMessage(ctx, sessionID, destination, req.MediaType, mediaData, req.Caption, req.FileName, mimeType)
	}

	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send media message")
		return nil, fmt.Errorf("failed to send media: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   sessionID,
		"destination": destination,
		"messageId":   messageID,
		"mediaType":   req.MediaType,
	}).Info().Msg("Media message sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID:     messageID,
		Status: "sent",
		Details: map[string]interface{}{
			"number":      req.Number,
			"groupJid":    req.GroupJid,
			"destination": destination,
			"sessionId":   sessionID,
			"type":        "media",
			"mediaType":   req.MediaType,
			"fileName":    req.FileName,
			"mimeType":    mimeType,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de mídia
func (uc *SendMediaMessageUseCase) validateRequest(req message.SendMediaMessageRequest) error {
	// Validar destinatário (number ou groupJid)
	if err := uc.numberValidator.ValidateDestination(req.Number, req.GroupJid); err != nil {
		return err
	}

	if req.Media == "" {
		return fmt.Errorf("media data is required")
	}

	if req.MediaType == "" {
		return fmt.Errorf("media type is required")
	}

	// Validar tipo de mídia
	validTypes := map[string]bool{
		"image":    true,
		"audio":    true,
		"video":    true,
		"document": true,
	}

	if !validTypes[req.MediaType] {
		return fmt.Errorf("invalid media type: %s (allowed: image, audio, video, document)", req.MediaType)
	}

	// Para documentos, nome do arquivo é obrigatório
	if req.MediaType == "document" && req.FileName == "" {
		return fmt.Errorf("file name is required for document type")
	}

	return nil
}

// decodeMediaData decodifica os dados da mídia (base64, data URL ou URL)
func (uc *SendMediaMessageUseCase) decodeMediaData(mediaData string) ([]byte, error) {
	// Se for uma URL HTTP/HTTPS, retornar nil (será tratado pelo cliente WhatsApp)
	if strings.HasPrefix(mediaData, "http://") || strings.HasPrefix(mediaData, "https://") {
		return nil, nil // URLs são tratadas diretamente pelo cliente
	}

	// Se for data URL (data:mime/type;base64,data), extrair apenas os dados
	if strings.HasPrefix(mediaData, "data:") {
		parts := strings.Split(mediaData, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		mediaData = parts[1]
	}

	// Decodificar base64
	data, err := base64.StdEncoding.DecodeString(mediaData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Validar tamanho (máximo 16MB)
	if len(data) > 16*1024*1024 {
		return nil, fmt.Errorf("media file too large (max 16MB)")
	}

	return data, nil
}

// detectMimeType detecta o MIME type baseado no tipo de mídia e nome do arquivo
func (uc *SendMediaMessageUseCase) detectMimeType(mediaType, fileName string) string {
	// Se há nome de arquivo, tentar detectar pela extensão
	if fileName != "" {
		if mimeType := mime.TypeByExtension(filepath.Ext(fileName)); mimeType != "" {
			return mimeType
		}
	}

	// Fallback para tipos padrão
	switch mediaType {
	case "image":
		return "image/jpeg"
	case "audio":
		return "audio/mpeg"
	case "video":
		return "video/mp4"
	case "document":
		return "application/octet-stream"
	default:
		return "application/octet-stream"
	}
}
