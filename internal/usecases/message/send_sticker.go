package message

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/vincent-petithory/dataurl"

	"zmeow/internal/domain/message"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// SendStickerMessageUseCase implementa o caso de uso para envio de sticker
type SendStickerMessageUseCase struct {
	sessionRepo     session.SessionRepository
	whatsappManager whatsapp.WhatsAppManager
	logger          logger.Logger
}

// NewSendStickerMessageUseCase cria uma nova instância do caso de uso
func NewSendStickerMessageUseCase(
	sessionRepo session.SessionRepository,
	whatsappManager whatsapp.WhatsAppManager,
	logger logger.Logger,
) *SendStickerMessageUseCase {
	return &SendStickerMessageUseCase{
		sessionRepo:     sessionRepo,
		whatsappManager: whatsappManager,
		logger:          logger,
	}
}

// Execute executa o caso de uso para enviar sticker
func (uc *SendStickerMessageUseCase) Execute(ctx context.Context, sessionID uuid.UUID, req message.SendStickerMessageRequest) (*message.SendMessageResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     req.Number,
		"mimeType":  req.MimeType,
	}).Info().Msg("Sending sticker message")

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

	// Normalizar número de telefone
	normalizedPhone := uc.normalizePhoneNumber(req.Number)

	// Obter dados do sticker (URL ou base64)
	stickerData, err := uc.getStickerData(ctx, req.Sticker)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get sticker data")
		return nil, fmt.Errorf("failed to get sticker data: %w", err)
	}

	// Determinar MIME type se não fornecido
	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "image/webp"
	}

	// Obter cliente WhatsApp
	client, err := uc.whatsappManager.GetClient(sessionID)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to get WhatsApp client")
		return nil, fmt.Errorf("failed to get WhatsApp client: %w", err)
	}

	// Enviar sticker
	messageID, err := client.SendStickerMessage(ctx, sessionID, normalizedPhone, stickerData, mimeType)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send sticker message")
		return nil, fmt.Errorf("failed to send sticker: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": sessionID,
		"phone":     normalizedPhone,
		"messageId": messageID,
	}).Info().Msg("Sticker message sent successfully")

	// Criar resposta
	response := &message.SendMessageResponse{
		ID: messageID,
		Status:    "sent",
		Details: map[string]interface{}{
			"phone":     normalizedPhone,
			"sessionId": sessionID,
			"type":      "sticker",
			"mimeType":  mimeType,
		},
	}

	return response, nil
}

// validateRequest valida a requisição de envio de sticker
func (uc *SendStickerMessageUseCase) validateRequest(req message.SendStickerMessageRequest) error {
	if req.Number == "" {
		return fmt.Errorf("phone number is required")
	}

	if req.Sticker == "" {
		return fmt.Errorf("sticker data is required")
	}

	// Validar formato do telefone
	if !uc.isValidPhoneNumber(req.Number) {
		return fmt.Errorf("invalid phone number format")
	}

	// Validar formato do sticker (base64 ou URL)
	if !uc.isValidStickerData(req.Sticker) {
		return fmt.Errorf("invalid sticker data format (must be base64, data URL, or HTTP URL)")
	}

	// Validar MIME type se fornecido
	if req.MimeType != "" && !uc.isValidStickerMimeType(req.MimeType) {
		return fmt.Errorf("invalid MIME type for sticker (allowed: image/webp, image/png, image/jpeg)")
	}

	return nil
}

// isValidPhoneNumber valida o formato do número de telefone
func (uc *SendStickerMessageUseCase) isValidPhoneNumber(phone string) bool {
	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Verificar se tem pelo menos 10 dígitos e no máximo 15
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}

	// Verificar se começa com código de país válido
	if len(cleaned) >= 11 && (strings.HasPrefix(cleaned, "55") || strings.HasPrefix(cleaned, "1")) {
		return true
	}

	// Aceitar números com 10-11 dígitos (formato nacional)
	if len(cleaned) >= 10 && len(cleaned) <= 11 {
		return true
	}

	return false
}

// normalizePhoneNumber normaliza o número de telefone para o formato WhatsApp
func (uc *SendStickerMessageUseCase) normalizePhoneNumber(phone string) string {
	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Se não tem código de país, assumir Brasil (55)
	if len(cleaned) == 10 || len(cleaned) == 11 {
		if !strings.HasPrefix(cleaned, "55") {
			cleaned = "55" + cleaned
		}
	}

	// Adicionar sufixo @s.whatsapp.net se não estiver presente
	if !strings.Contains(phone, "@") {
		return cleaned + "@s.whatsapp.net"
	}

	return phone
}

// isValidStickerData valida se os dados são URL válida ou data URL válida
func (uc *SendStickerMessageUseCase) isValidStickerData(data string) bool {
	// Verificar se é URL HTTP/HTTPS
	if uc.isValidURL(data) {
		return true
	}

	// Verificar se é data URL válida (como WuzAPI)
	if strings.HasPrefix(data, "data:image/") {
		_, err := dataurl.DecodeString(data)
		return err == nil
	}

	return false
}

// isValidURL valida se a string é uma URL HTTP/HTTPS válida
func (uc *SendStickerMessageUseCase) isValidURL(data string) bool {
	if strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://") {
		// Validação básica de URL
		return len(data) > 10 && strings.Contains(data, ".")
	}
	return false
}

// isValidStickerMimeType valida o MIME type para stickers
func (uc *SendStickerMessageUseCase) isValidStickerMimeType(mimeType string) bool {
	validTypes := map[string]bool{
		"image/webp": true,
		"image/png":  true,
		"image/jpeg": true,
	}
	return validTypes[mimeType]
}

// getStickerData obtém os dados do sticker (baixa URL ou decodifica base64) e converte para WEBP
func (uc *SendStickerMessageUseCase) getStickerData(ctx context.Context, data string) ([]byte, error) {
	var rawData []byte
	var err error

	// Se for URL, baixar
	if uc.isValidURL(data) {
		rawData, err = uc.downloadStickerFromURL(ctx, data)
	} else {
		// Se for base64, decodificar
		rawData, err = uc.decodeStickerData(data)
	}

	if err != nil {
		return nil, err
	}

	// Converter para WEBP 512x512 (requisito WhatsApp)
	webpData, err := uc.convertToStickerFormat(rawData)
	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to convert to WEBP format, using original data")
		return rawData, nil
	}

	return webpData, nil
}

// downloadStickerFromURL baixa o sticker de uma URL
func (uc *SendStickerMessageUseCase) downloadStickerFromURL(ctx context.Context, url string) ([]byte, error) {
	// Criar requisição HTTP
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Fazer requisição
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download sticker: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download sticker: HTTP %d", resp.StatusCode)
	}

	// Ler dados
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sticker data: %w", err)
	}

	// Validar tamanho
	if len(data) < 1024 {
		return nil, fmt.Errorf("downloaded sticker too small (minimum 1KB for WhatsApp compatibility)")
	}

	if len(data) > 5*1024*1024 {
		return nil, fmt.Errorf("downloaded sticker too large (maximum 5MB)")
	}

	return data, nil
}

// decodeStickerData decodifica os dados do sticker usando dataurl
func (uc *SendStickerMessageUseCase) decodeStickerData(data string) ([]byte, error) {
	// Verificar se é data URL válida
	if !strings.HasPrefix(data, "data:image/") {
		return nil, fmt.Errorf("sticker data should start with 'data:image/' (supported formats: webp, png, jpeg)")
	}

	// Decodificar data URL
	dataURL, err := dataurl.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("could not decode base64 encoded data from payload: %w", err)
	}

	filedata := dataURL.Data

	// Validar tamanho
	if len(filedata) < 1024 {
		return nil, fmt.Errorf("sticker data too small (minimum 1KB for WhatsApp compatibility)")
	}

	if len(filedata) > 5*1024*1024 {
		return nil, fmt.Errorf("sticker data too large (maximum 5MB)")
	}

	return filedata, nil
}

// convertToStickerFormat converte imagem para formato WEBP 512x512 (requisito WhatsApp)
func (uc *SendStickerMessageUseCase) convertToStickerFormat(imageData []byte) ([]byte, error) {
	// Detectar tipo da imagem
	contentType := http.DetectContentType(imageData)

	// Se já for WEBP, retornar como está
	if contentType == "image/webp" {
		return imageData, nil
	}

	// Decodificar imagem
	var img image.Image
	var err error

	reader := bytes.NewReader(imageData)

	switch contentType {
	case "image/png":
		img, err = png.Decode(reader)
	case "image/jpeg":
		img, err = jpeg.Decode(reader)
	default:
		// Tentar decodificar como PNG primeiro, depois JPEG
		img, err = png.Decode(reader)
		if err != nil {
			reader.Seek(0, 0)
			img, err = jpeg.Decode(reader)
		}
	}

	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to decode image, using original data")
		return imageData, nil
	}

	// Redimensionar para 512x512 (requisito WhatsApp)
	resizedImg := resize.Resize(512, 512, img, resize.Lanczos3)

	// Converter para WEBP
	var buf bytes.Buffer
	err = webp.Encode(&buf, resizedImg, &webp.Options{
		Lossless: false,
		Quality:  80,
	})
	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to encode WEBP, using original data")
		return imageData, nil
	}

	return buf.Bytes(), nil
}
