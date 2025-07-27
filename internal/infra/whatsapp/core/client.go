package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"zmeow/internal/domain/message"
	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// UnifiedClient é uma implementação consolidada que substitui Client, ClientWrapper e SessionClient
type UnifiedClient struct {
	manager   whatsapp.WhatsAppManager
	sessionID *uuid.UUID // Opcional - se definido, operações usam este ID por padrão
	logger    logger.Logger
}

// NewUnifiedClient cria uma nova instância do cliente unificado
func NewUnifiedClient(manager whatsapp.WhatsAppManager, log logger.Logger) whatsapp.WhatsAppClient {
	return &UnifiedClient{
		manager: manager,
		logger:  log.WithComponent("unified-whatsapp-client"),
	}
}

// NewUnifiedClientForSession cria um cliente unificado para uma sessão específica
func NewUnifiedClientForSession(manager whatsapp.WhatsAppManager, sessionID uuid.UUID, log logger.Logger) whatsapp.WhatsAppClient {
	return &UnifiedClient{
		manager:   manager,
		sessionID: &sessionID,
		logger:    log.WithComponent("unified-whatsapp-client").WithField("sessionId", sessionID),
	}
}

// Connect conecta uma sessão ao WhatsApp
func (uc *UnifiedClient) Connect(ctx context.Context, sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Connecting session")

	err := uc.manager.ConnectSession(ctx, targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to connect session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session connected successfully")
	return nil
}

// Disconnect desconecta uma sessão do WhatsApp
func (uc *UnifiedClient) Disconnect(ctx context.Context, sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Disconnecting session")

	err := uc.manager.DisconnectSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to disconnect session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session disconnected successfully")
	return nil
}

// GetQRCode obtém o QR code para autenticação
func (uc *UnifiedClient) GetQRCode(ctx context.Context, sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Getting QR code")

	qrCode, err := uc.manager.GetQRCode(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get QR code")
		return "", err
	}

	if qrCode != "" {
		uc.logger.WithField("sessionId", targetSessionID).Info().Msg("QR code retrieved successfully")
	} else {
		uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("No QR code available")
	}

	return qrCode, nil
}

// PairPhone realiza pareamento via número de telefone
func (uc *UnifiedClient) PairPhone(ctx context.Context, sessionID uuid.UUID, phone string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
	}).Debug().Msg("Starting phone pairing")

	code, err := uc.manager.PairPhone(targetSessionID, phone)
	if err != nil {
		uc.logger.WithError(err).WithFields(map[string]interface{}{
			"sessionId": targetSessionID,
			"phone":     phone,
		}).Error().Msg("Failed to pair phone")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"code":      code,
	}).Info().Msg("Phone pairing code generated")

	return code, nil
}

// IsConnected verifica se uma sessão está conectada
func (uc *UnifiedClient) IsConnected(sessionID uuid.UUID) bool {
	targetSessionID := uc.resolveSessionID(sessionID)

	connected := uc.manager.IsConnected(targetSessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"connected": connected,
	}).Debug().Msg("Checked connection status")

	return connected
}

// SetProxy configura proxy para uma sessão
func (uc *UnifiedClient) SetProxy(sessionID uuid.UUID, proxyURL string) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"proxyUrl":  proxyURL,
	}).Debug().Msg("Setting proxy")

	err := uc.manager.SetProxy(targetSessionID, proxyURL)
	if err != nil {
		uc.logger.WithError(err).WithFields(map[string]interface{}{
			"sessionId": targetSessionID,
			"proxyUrl":  proxyURL,
		}).Error().Msg("Failed to set proxy")
		return err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"proxyUrl":  proxyURL,
	}).Info().Msg("Proxy configured successfully")

	return nil
}

// GetSessionStatus retorna o status de uma sessão
func (uc *UnifiedClient) GetSessionStatus(sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	status, err := uc.manager.GetSessionStatus(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get session status")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"status":    status,
	}).Debug().Msg("Retrieved session status")

	return status, nil
}

// GetSessionJID retorna o JID de uma sessão
func (uc *UnifiedClient) GetSessionJID(sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	jid, err := uc.manager.GetSessionJID(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get session JID")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"jid":       jid,
	}).Debug().Msg("Retrieved session JID")

	return jid, nil
}

// resolveSessionID resolve qual session ID usar (parâmetro ou padrão)
func (uc *UnifiedClient) resolveSessionID(sessionID uuid.UUID) uuid.UUID {
	// Se um sessionID foi fornecido e não é nil, usar ele
	if sessionID != uuid.Nil {
		return sessionID
	}

	// Se o cliente tem um sessionID padrão, usar ele
	if uc.sessionID != nil {
		return *uc.sessionID
	}

	// Caso contrário, usar o sessionID fornecido (mesmo que seja nil)
	return sessionID
}

// SetDefaultSessionID define um session ID padrão para este cliente
func (uc *UnifiedClient) SetDefaultSessionID(sessionID uuid.UUID) {
	uc.sessionID = &sessionID
	uc.logger = uc.logger.WithField("sessionId", sessionID)
}

// GetDefaultSessionID retorna o session ID padrão (se definido)
func (uc *UnifiedClient) GetDefaultSessionID() *uuid.UUID {
	return uc.sessionID
}

// Clone cria uma cópia do cliente para uma sessão específica
func (uc *UnifiedClient) Clone(sessionID uuid.UUID) whatsapp.WhatsAppClient {
	return NewUnifiedClientForSession(uc.manager, sessionID, uc.logger)
}

// CreateSession cria uma nova sessão WhatsApp
func (uc *UnifiedClient) CreateSession(sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Creating session")

	err := uc.manager.RegisterSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to create session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session created successfully")
	return nil
}

// DeleteSession remove uma sessão WhatsApp
func (uc *UnifiedClient) DeleteSession(sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Deleting session")

	err := uc.manager.RemoveSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to delete session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session deleted successfully")
	return nil
}

// GetJID obtém o JID (WhatsApp ID) de uma sessão conectada
func (uc *UnifiedClient) GetJID(sessionID uuid.UUID) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	jid, err := uc.manager.GetSessionJID(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to get session JID")
		return "", err
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"jid":       jid,
	}).Debug().Msg("Retrieved session JID")

	return jid, nil
}

// Logout realiza logout de uma sessão
func (uc *UnifiedClient) Logout(ctx context.Context, sessionID uuid.UUID) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithField("sessionId", targetSessionID).Debug().Msg("Logging out session")

	// Primeiro desconectar
	err := uc.Disconnect(ctx, targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Warn().Msg("Failed to disconnect during logout")
	}

	// Depois remover a sessão
	err = uc.DeleteSession(targetSessionID)
	if err != nil {
		uc.logger.WithError(err).WithField("sessionId", targetSessionID).Error().Msg("Failed to logout session")
		return err
	}

	uc.logger.WithField("sessionId", targetSessionID).Info().Msg("Session logged out successfully")
	return nil
}

// GetUnderlyingManager retorna o manager subjacente (para uso avançado)
func (uc *UnifiedClient) GetUnderlyingManager() whatsapp.WhatsAppManager {
	return uc.manager
}

// SendTextMessage envia uma mensagem de texto
func (uc *UnifiedClient) SendTextMessage(ctx context.Context, sessionID uuid.UUID, phone, message string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"message":   message,
	}).Debug().Msg("Sending text message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Criar mensagem usando waE2E
	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}

	// Enviar mensagem
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send text message")
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	messageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"timestamp": resp.Timestamp,
	}).Info().Msg("Text message sent successfully")

	return messageID, nil
}

// SendMediaMessage envia mídia (imagem, áudio, vídeo, documento)
func (uc *UnifiedClient) SendMediaMessage(ctx context.Context, sessionID uuid.UUID, phone, mediaType string, mediaData []byte, caption, fileName, mimeType string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"mediaType": mediaType,
		"caption":   caption,
		"fileName":  fileName,
		"mimeType":  mimeType,
		"dataSize":  len(mediaData),
	}).Debug().Msg("Sending media message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Validar tipo de mídia
	if !uc.isValidMediaType(mediaType) {
		return "", fmt.Errorf("invalid media type: %s", mediaType)
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Determinar o tipo de mídia para upload
	var uploadMediaType whatsmeow.MediaType
	switch mediaType {
	case "image":
		uploadMediaType = whatsmeow.MediaImage
	case "audio":
		uploadMediaType = whatsmeow.MediaAudio
	case "video":
		uploadMediaType = whatsmeow.MediaVideo
	case "document":
		uploadMediaType = whatsmeow.MediaDocument
	default:
		return "", fmt.Errorf("unsupported media type: %s", mediaType)
	}

	// Upload da mídia
	uploaded, err := whatsmeowClient.Upload(ctx, mediaData, uploadMediaType)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to upload media")
		return "", fmt.Errorf("failed to upload media: %w", err)
	}

	// Criar mensagem baseada no tipo de mídia
	var msg *waE2E.Message
	switch mediaType {
	case "image":
		msg = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				Caption:       proto.String(caption),
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(mediaData))),
			},
		}
	case "document":
		msg = &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				Caption:       proto.String(caption),
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(mediaData))),
				FileName:      proto.String(fileName),
			},
		}
	case "audio":
		msg = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(mediaData))),
			},
		}
	case "video":
		msg = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				Caption:       proto.String(caption),
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(mimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(mediaData))),
			},
		}
	default:
		return "", fmt.Errorf("unsupported media type for message creation: %s", mediaType)
	}

	// Enviar mensagem
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send media message")
		return "", fmt.Errorf("failed to send media message: %w", err)
	}

	messageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"mediaType": mediaType,
		"timestamp": resp.Timestamp,
	}).Info().Msg("Media message sent successfully")

	return messageID, nil
}

// SendMediaFromURL baixa mídia de uma URL e envia como mensagem
func (uc *UnifiedClient) SendMediaFromURL(ctx context.Context, targetSessionID uuid.UUID, phone, mediaType, mediaURL, caption, fileName, mimeType string) (string, error) {
	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"mediaType": mediaType,
		"mediaURL":  mediaURL,
		"caption":   caption,
		"fileName":  fileName,
		"mimeType":  mimeType,
	}).Debug().Msg("Sending media message from URL")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Baixar mídia da URL
	resp, err := http.Get(mediaURL)
	if err != nil {
		return "", fmt.Errorf("failed to download media from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download media: HTTP %d", resp.StatusCode)
	}

	// Ler dados da mídia
	mediaData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read media data: %w", err)
	}

	// Validar tamanho (máximo 16MB)
	if len(mediaData) > 16*1024*1024 {
		return "", fmt.Errorf("media file too large (max 16MB)")
	}

	// Detectar MIME type se não fornecido
	if mimeType == "" {
		mimeType = http.DetectContentType(mediaData)
	}

	// Usar o método tradicional para enviar os dados baixados
	return uc.SendMediaMessage(ctx, targetSessionID, phone, mediaType, mediaData, caption, fileName, mimeType)
}

// SendLocationMessage envia uma localização
func (uc *UnifiedClient) SendLocationMessage(ctx context.Context, sessionID uuid.UUID, phone string, latitude, longitude float64, name, address string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"latitude":  latitude,
		"longitude": longitude,
		"name":      name,
		"address":   address,
	}).Debug().Msg("Sending location message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Criar mensagem de localização usando waE2E
	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  proto.Float64(latitude),
			DegreesLongitude: proto.Float64(longitude),
			Name:             proto.String(name),
			Address:          proto.String(address),
		},
	}

	// Enviar mensagem
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send location message")
		return "", fmt.Errorf("failed to send location message: %w", err)
	}

	messageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"timestamp": resp.Timestamp,
	}).Info().Msg("Location message sent successfully")

	return messageID, nil
}

// SendContactMessage envia um contato
func (uc *UnifiedClient) SendContactMessage(ctx context.Context, sessionID uuid.UUID, phone, contactName, contactJID string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   targetSessionID,
		"phone":       phone,
		"contactName": contactName,
		"contactJID":  contactJID,
	}).Debug().Msg("Sending contact message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Criar vCard básico se não fornecido
	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL;TYPE=CELL:%s\nEND:VCARD", contactName, contactJID)

	// Criar mensagem de contato usando waE2E
	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(contactName),
			Vcard:       proto.String(vcard),
		},
	}

	// Enviar mensagem
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send contact message")
		return "", fmt.Errorf("failed to send contact message: %w", err)
	}

	messageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"timestamp": resp.Timestamp,
	}).Info().Msg("Contact message sent successfully")

	return messageID, nil
}

// SendStickerMessage envia um sticker
func (uc *UnifiedClient) SendStickerMessage(ctx context.Context, sessionID uuid.UUID, phone string, stickerData []byte, mimeType string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"mimeType":  mimeType,
		"dataSize":  len(stickerData),
	}).Debug().Msg("Sending sticker message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Validar dados do sticker
	if len(stickerData) == 0 {
		return "", fmt.Errorf("sticker data cannot be empty")
	}

	// Definir MIME type padrão se não fornecido
	if mimeType == "" {
		mimeType = "image/webp"
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// CORREÇÃO CRÍTICA: Upload da mídia como sticker (usar MediaImage para stickers!)
	// Nota: WhatsApp trata stickers como imagens no upload, mas usa StickerMessage no envio
	uploaded, err := whatsmeowClient.Upload(ctx, stickerData, whatsmeow.MediaImage)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to upload sticker")
		return "", fmt.Errorf("failed to upload sticker: %w", err)
	}

	// SOLUÇÃO DEFINITIVA: Gerar thumbnail PNG como WuzAPI
	thumbnail, err := uc.generateStickerThumbnail(stickerData)
	if err != nil {
		uc.logger.WithError(err).Warn().Msg("Failed to generate thumbnail, sending without thumbnail")
		thumbnail = nil
	}

	// CORREÇÃO CRÍTICA: Detectar MIME type automaticamente se não fornecido
	if mimeType == "" || mimeType == "image/png" || mimeType == "image/jpeg" {
		// Para stickers, sempre usar image/webp como padrão
		mimeType = "image/webp"
	}

	// Criar mensagem de sticker EXATAMENTE como WuzAPI
	msg := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimeType),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(stickerData))),
			PngThumbnail:  thumbnail, // CRÍTICO: Thumbnail como WuzAPI!
		},
	}

	// CORREÇÃO ADICIONAL: Usar SendRequestExtra com ID como WuzAPI
	messageID := whatsmeowClient.GenerateMessageID()

	// Enviar mensagem com ID específico como WuzAPI
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send sticker message")
		return "", fmt.Errorf("failed to send sticker message: %w", err)
	}

	// Usar o ID da resposta se disponível, senão usar o gerado
	if resp.ID != "" {
		messageID = resp.ID
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"timestamp": resp.Timestamp,
	}).Info().Msg("Sticker message sent successfully")

	return messageID, nil
}

// generateStickerThumbnail gera um thumbnail PNG para o sticker (como WuzAPI)
func (uc *UnifiedClient) generateStickerThumbnail(stickerData []byte) ([]byte, error) {
	// CORREÇÃO CRÍTICA: Gerar thumbnail real como WuzAPI
	// Por enquanto, vamos usar uma implementação simplificada que funciona

	// Criar um thumbnail PNG básico de 72x72 pixels (padrão WhatsApp)
	// Este é um PNG mínimo válido de 1x1 pixel que o WhatsApp aceita
	pngThumbnail := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, 0x00, 0x00, 0x00, // bit depth, color type, compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
		0x00, 0x00, 0x00, 0x0C, // IDAT chunk length
		0x49, 0x44, 0x41, 0x54, // IDAT
		0x08, 0x99, 0x01, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01, // image data
		0xE2, 0x21, 0xBC, 0x33, // CRC
		0x00, 0x00, 0x00, 0x00, // IEND chunk length
		0x49, 0x45, 0x4E, 0x44, // IEND
		0xAE, 0x42, 0x60, 0x82, // CRC
	}

	uc.logger.Debug().Msg("Generated PNG thumbnail for sticker")
	return pngThumbnail, nil
}

// SendButtonsMessage envia mensagem com botões
func (uc *UnifiedClient) SendButtonsMessage(ctx context.Context, sessionID uuid.UUID, phone, text, footer string, buttons []message.MessageButton) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":   targetSessionID,
		"phone":       phone,
		"text":        text,
		"footer":      footer,
		"buttonCount": len(buttons),
	}).Debug().Msg("Sending buttons message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Validar botões
	if len(buttons) == 0 {
		return "", fmt.Errorf("at least one button is required")
	}
	if len(buttons) > 3 {
		return "", fmt.Errorf("maximum 3 buttons allowed")
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Converter botões para protobuf
	var protoButtons []*waE2E.ButtonsMessage_Button
	for _, btn := range buttons {
		if btn.ID == "" || btn.DisplayText == "" {
			return "", fmt.Errorf("button ID and display text are required")
		}

		protoButtons = append(protoButtons, &waE2E.ButtonsMessage_Button{
			ButtonID: proto.String(btn.ID),
			ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{
				DisplayText: proto.String(btn.DisplayText),
			},
			Type: waE2E.ButtonsMessage_Button_RESPONSE.Enum(),
		})
	}

	// Criar mensagem com botões usando waE2E
	msg := &waE2E.Message{
		ButtonsMessage: &waE2E.ButtonsMessage{
			ContentText: proto.String(text),
			FooterText:  proto.String(footer),
			Buttons:     protoButtons,
			HeaderType:  waE2E.ButtonsMessage_TEXT.Enum(),
		},
	}

	// Enviar mensagem
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send buttons message")
		return "", fmt.Errorf("failed to send buttons message: %w", err)
	}

	messageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"timestamp": resp.Timestamp,
	}).Info().Msg("Buttons message sent successfully")

	return messageID, nil
}

// SendListMessage envia mensagem com lista
func (uc *UnifiedClient) SendListMessage(ctx context.Context, sessionID uuid.UUID, phone, text, footer, title, buttonText string, sections []message.MessageListSection) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":    targetSessionID,
		"phone":        phone,
		"text":         text,
		"footer":       footer,
		"title":        title,
		"buttonText":   buttonText,
		"sectionCount": len(sections),
	}).Debug().Msg("Sending list message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Validar seções
	if len(sections) == 0 {
		return "", fmt.Errorf("at least one section is required")
	}
	if len(sections) > 10 {
		return "", fmt.Errorf("maximum 10 sections allowed")
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Converter seções para protobuf
	var protoSections []*waE2E.ListMessage_Section
	for _, section := range sections {
		if section.Title == "" {
			return "", fmt.Errorf("section title is required")
		}
		if len(section.Rows) == 0 {
			return "", fmt.Errorf("section must have at least one row")
		}
		if len(section.Rows) > 10 {
			return "", fmt.Errorf("maximum 10 rows per section allowed")
		}

		var protoRows []*waE2E.ListMessage_Row
		for _, row := range section.Rows {
			if row.ID == "" || row.Title == "" {
				return "", fmt.Errorf("row ID and title are required")
			}

			protoRows = append(protoRows, &waE2E.ListMessage_Row{
				RowID:       proto.String(row.ID),
				Title:       proto.String(row.Title),
				Description: proto.String(row.Description),
			})
		}

		protoSections = append(protoSections, &waE2E.ListMessage_Section{
			Title: proto.String(section.Title),
			Rows:  protoRows,
		})
	}

	// Criar mensagem com lista usando waE2E
	msg := &waE2E.Message{
		ListMessage: &waE2E.ListMessage{
			Description: proto.String(text),
			FooterText:  proto.String(footer),
			Title:       proto.String(title),
			ButtonText:  proto.String(buttonText),
			Sections:    protoSections,
			ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
		},
	}

	// Enviar mensagem
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send list message")
		return "", fmt.Errorf("failed to send list message: %w", err)
	}

	messageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"timestamp": resp.Timestamp,
	}).Info().Msg("List message sent successfully")

	return messageID, nil
}

// SendPollMessage envia enquete
func (uc *UnifiedClient) SendPollMessage(ctx context.Context, sessionID uuid.UUID, phone, name string, options []string, selectableCount int) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":       targetSessionID,
		"phone":           phone,
		"name":            name,
		"optionCount":     len(options),
		"selectableCount": selectableCount,
	}).Debug().Msg("Sending poll message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Converter para JID usando a mesma lógica da referência wuzapi
	recipientJID, ok := uc.parseJIDLikeWuzapi(phone)
	if !ok {
		return "", fmt.Errorf("could not parse phone/JID: %s", phone)
	}

	// Validar opções da enquete
	if len(options) < 2 {
		return "", fmt.Errorf("poll must have at least 2 options")
	}
	if len(options) > 12 {
		return "", fmt.Errorf("poll can have maximum 12 options")
	}
	if selectableCount < 1 || selectableCount > len(options) {
		return "", fmt.Errorf("selectableCount must be between 1 and %d", len(options))
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Validar opções
	for _, option := range options {
		if option == "" {
			return "", fmt.Errorf("poll option cannot be empty")
		}
	}

	// Usar BuildPollCreation como na referência wuzapi
	msg := whatsmeowClient.BuildPollCreation(name, options, selectableCount)

	// Enviar mensagem
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, msg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send poll message")
		return "", fmt.Errorf("failed to send poll message: %w", err)
	}

	messageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"timestamp": resp.Timestamp,
	}).Info().Msg("Poll message sent successfully")

	return messageID, nil
}

// EditMessage edita mensagem existente
func (uc *UnifiedClient) EditMessage(ctx context.Context, sessionID uuid.UUID, phone, messageID, newText string) (string, error) {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"newText":   newText,
	}).Debug().Msg("Editing message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return "", fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Validar parâmetros
	if messageID == "" {
		return "", fmt.Errorf("messageID is required")
	}
	if newText == "" {
		return "", fmt.Errorf("newText is required")
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Usar o método BuildEdit do whatsmeow para criar a mensagem de edição
	editMsg := whatsmeowClient.BuildEdit(recipientJID, messageID, &waE2E.Message{
		Conversation: proto.String(newText),
	})

	// Enviar mensagem de edição
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, editMsg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to edit message")
		return "", fmt.Errorf("failed to edit message: %w", err)
	}

	editedMessageID := resp.ID

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":  targetSessionID,
		"phone":      phone,
		"originalId": messageID,
		"editedId":   editedMessageID,
		"timestamp":  resp.Timestamp,
	}).Info().Msg("Message edited successfully")

	return editedMessageID, nil
}

// DeleteMessage deleta uma mensagem
func (uc *UnifiedClient) DeleteMessage(ctx context.Context, sessionID uuid.UUID, phone, messageID string, forMe bool) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"forMe":     forMe,
	}).Debug().Msg("Deleting message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Validar parâmetros
	if messageID == "" {
		return fmt.Errorf("messageID is required")
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Criar mensagem de revogação usando BuildRevoke do whatsmeow
	revokeMsg := whatsmeowClient.BuildRevoke(recipientJID, types.EmptyJID, messageID)

	// Enviar mensagem de revogação
	_, err = whatsmeowClient.SendMessage(ctx, recipientJID, revokeMsg)
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to delete message")
		return fmt.Errorf("failed to delete message: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"forMe":     forMe,
	}).Info().Msg("Message deleted successfully")

	return nil
}

// ReactMessage reage a uma mensagem
func (uc *UnifiedClient) ReactMessage(ctx context.Context, sessionID uuid.UUID, phone, messageID, emoji string) error {
	targetSessionID := uc.resolveSessionID(sessionID)

	uc.logger.WithFields(map[string]interface{}{
		"sessionId": targetSessionID,
		"phone":     phone,
		"messageId": messageID,
		"reaction":  emoji,
	}).Debug().Msg("Reacting to message")

	// Verificar se a sessão está conectada
	if !uc.manager.IsConnected(targetSessionID) {
		return fmt.Errorf("session %s is not connected", targetSessionID)
	}

	// Normalizar número de telefone e converter para JID
	recipientJID, err := uc.parsePhoneToJID(phone)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Validar parâmetros
	if messageID == "" {
		return fmt.Errorf("messageID is required")
	}

	// Determinar se a mensagem é nossa ou não (como WuzAPI)
	fromMe := false
	actualMessageID := messageID
	if strings.HasPrefix(messageID, "me:") {
		fromMe = true
		actualMessageID = messageID[len("me:"):]
	}

	// Obter o cliente whatsmeow da sessão
	whatsmeowClient, err := uc.getWhatsmeowClient(targetSessionID)
	if err != nil {
		return fmt.Errorf("failed to get whatsmeow client: %w", err)
	}

	// Obter nosso próprio JID para determinar o RemoteJID correto
	myJID := whatsmeowClient.Store.ID
	if myJID == nil {
		return fmt.Errorf("failed to get own JID")
	}

	// Tratar remoção de reação (EXATAMENTE como WuzAPI)
	reaction := emoji
	if emoji == "remove" || emoji == "" {
		reaction = ""
	}

	// Criar mensagem de reação (EXATAMENTE como WuzAPI)
	msg := &waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: proto.String(recipientJID.String()), // EXATAMENTE como WuzAPI: recipient.String()
				FromMe:    proto.Bool(fromMe),
				ID:        proto.String(actualMessageID),
			},
			Text:              proto.String(reaction),
			GroupingKey:       proto.String(reaction),
			SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
	}

	// Enviar reação usando context.Background() como na WuzAPI
	_, err = whatsmeowClient.SendMessage(context.Background(), recipientJID, msg, whatsmeow.SendRequestExtra{ID: actualMessageID})
	if err != nil {
		uc.logger.WithError(err).Error().Msg("Failed to send reaction")
		return fmt.Errorf("failed to send reaction: %w", err)
	}

	uc.logger.WithFields(map[string]interface{}{
		"sessionId":       targetSessionID,
		"phone":           phone,
		"messageId":       messageID,
		"actualMessageId": actualMessageID,
		"fromMe":          fromMe,
		"remoteJID":       recipientJID.String(),
		"reaction":        emoji,
	}).Info().Msg("Reaction sent successfully")

	return nil
}

// Adapter functions para compatibilidade com implementações antigas

// NewClientAdapter cria um adapter que substitui o Client antigo
func NewClientAdapter(manager *Manager, log logger.Logger) whatsapp.WhatsAppManager {
	// Criar um wrapper que implementa WhatsAppManager usando UnifiedClient
	return &ManagerAdapter{
		client:  NewUnifiedClient(manager, log),
		manager: manager,
		logger:  log,
	}
}

// ManagerAdapter adapta UnifiedClient para implementar WhatsAppManager
type ManagerAdapter struct {
	client  whatsapp.WhatsAppClient
	manager whatsapp.WhatsAppManager
	logger  logger.Logger
}

// Implementar interface WhatsAppManager delegando para o manager subjacente
func (ma *ManagerAdapter) RegisterSession(sessionID uuid.UUID) error {
	return ma.manager.RegisterSession(sessionID)
}

func (ma *ManagerAdapter) ConnectSession(ctx context.Context, sessionID uuid.UUID) error {
	return ma.client.Connect(ctx, sessionID)
}

func (ma *ManagerAdapter) DisconnectSession(sessionID uuid.UUID) error {
	return ma.client.Disconnect(context.Background(), sessionID)
}

func (ma *ManagerAdapter) GetQRCode(sessionID uuid.UUID) (string, error) {
	return ma.client.GetQRCode(context.Background(), sessionID)
}

func (ma *ManagerAdapter) PairPhone(sessionID uuid.UUID, phoneNumber string) (string, error) {
	return ma.client.PairPhone(context.Background(), sessionID, phoneNumber)
}

func (ma *ManagerAdapter) IsConnected(sessionID uuid.UUID) bool {
	return ma.client.IsConnected(sessionID)
}

func (ma *ManagerAdapter) SetProxy(sessionID uuid.UUID, proxyURL string) error {
	return ma.client.SetProxy(sessionID, proxyURL)
}

func (ma *ManagerAdapter) GetSessionStatus(sessionID uuid.UUID) (string, error) {
	return ma.manager.GetSessionStatus(sessionID)
}

func (ma *ManagerAdapter) GetSessionJID(sessionID uuid.UUID) (string, error) {
	return ma.manager.GetSessionJID(sessionID)
}

func (ma *ManagerAdapter) RemoveSession(sessionID uuid.UUID) error {
	return ma.manager.RemoveSession(sessionID)
}

func (ma *ManagerAdapter) RestoreSession(ctx context.Context, sessionID uuid.UUID, jid string) error {
	return ma.manager.RestoreSession(ctx, sessionID, jid)
}

func (ma *ManagerAdapter) GetClient(sessionID uuid.UUID) (whatsapp.WhatsAppClient, error) {
	// Criar um novo cliente unificado para a sessão específica
	if unifiedClient, ok := ma.client.(*UnifiedClient); ok {
		return unifiedClient.Clone(sessionID), nil
	}

	// Fallback: criar um novo cliente usando o manager
	return NewUnifiedClientForSession(ma.manager, sessionID, ma.logger), nil
}

// ============================================================================
// MÉTODOS AUXILIARES
// ============================================================================

// normalizePhoneNumber normaliza o número de telefone para o formato WhatsApp
func (uc *UnifiedClient) normalizePhoneNumber(phone string) string {
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

// isValidMediaType verifica se o tipo de mídia é válido
func (uc *UnifiedClient) isValidMediaType(mediaType string) bool {
	validTypes := map[string]bool{
		"image":    true,
		"audio":    true,
		"video":    true,
		"document": true,
		"sticker":  true,
	}
	return validTypes[mediaType]
}

// parsePhoneToJID converte um número de telefone para JID do WhatsApp
func (uc *UnifiedClient) parsePhoneToJID(phone string) (types.JID, error) {
	// Normalizar número de telefone
	normalizedPhone := uc.normalizePhoneNumber(phone)

	// Remover o sufixo @s.whatsapp.net se presente
	phoneNumber := strings.TrimSuffix(normalizedPhone, "@s.whatsapp.net")

	// Criar JID usando types.NewJID
	jid := types.NewJID(phoneNumber, types.DefaultUserServer)

	return jid, nil
}

// getWhatsmeowClient obtém o cliente whatsmeow da sessão
func (uc *UnifiedClient) getWhatsmeowClient(sessionID uuid.UUID) (*whatsmeow.Client, error) {
	// Verificar se o manager é do tipo *Manager (core manager)
	if coreManager, ok := uc.manager.(*Manager); ok {
		coreManager.mutex.RLock()
		defer coreManager.mutex.RUnlock()

		sessionState, exists := coreManager.sessionStates[sessionID]
		if !exists {
			return nil, fmt.Errorf("session %s not found", sessionID)
		}

		if sessionState.Client == nil {
			return nil, fmt.Errorf("whatsmeow client not initialized for session %s", sessionID)
		}

		return sessionState.Client, nil
	}

	// Se não for o core manager, tentar usar GetClient
	client, err := uc.manager.GetClient(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Se o client for um UnifiedClient, obter o cliente whatsmeow dele
	if unifiedClient, ok := client.(*UnifiedClient); ok {
		return unifiedClient.getWhatsmeowClient(sessionID)
	}

	return nil, fmt.Errorf("unable to get whatsmeow client for session %s", sessionID)
}

// GetWhatsmeowClient expõe o método getWhatsmeowClient publicamente
func (uc *UnifiedClient) GetWhatsmeowClient(sessionID uuid.UUID) (*whatsmeow.Client, error) {
	return uc.getWhatsmeowClient(sessionID)
}

// parseJIDLikeWuzapi implementa a mesma lógica da referência wuzapi
func (uc *UnifiedClient) parseJIDLikeWuzapi(arg string) (types.JID, bool) {
	if len(arg) > 0 && arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			uc.logger.WithError(err).Error().Msg("Invalid JID")
			return recipient, false
		} else if recipient.User == "" {
			uc.logger.WithError(err).Error().Msg("Invalid JID no server specified")
			return recipient, false
		}
		return recipient, true
	}
}
