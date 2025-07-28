package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"zmeow/internal/domain/message"
	"zmeow/internal/http/responses"
	messageUseCases "zmeow/internal/usecases/message"
	"zmeow/pkg/logger"
)

// MessageHandler implementa os handlers para mensagens
type MessageHandler struct {
	sendTextUseCase      *messageUseCases.SendTextMessageUseCase
	sendMediaUseCase     *messageUseCases.SendMediaMessageUseCase
	sendLocationUseCase  *messageUseCases.SendLocationMessageUseCase
	sendContactUseCase   *messageUseCases.SendContactMessageUseCase
	sendStickerUseCase   *messageUseCases.SendStickerMessageUseCase
	sendButtonsUseCase   *messageUseCases.SendButtonsMessageUseCase
	sendListUseCase      *messageUseCases.SendListMessageUseCase
	sendPollUseCase      *messageUseCases.SendPollMessageUseCase
	editMessageUseCase   *messageUseCases.EditMessageUseCase
	deleteMessageUseCase *messageUseCases.DeleteMessageUseCase
	reactMessageUseCase  *messageUseCases.ReactMessageUseCase
	logger               logger.Logger
}

// NewMessageHandler cria uma nova instância do message handler
func NewMessageHandler(
	sendTextUseCase *messageUseCases.SendTextMessageUseCase,
	sendMediaUseCase *messageUseCases.SendMediaMessageUseCase,
	sendLocationUseCase *messageUseCases.SendLocationMessageUseCase,
	sendContactUseCase *messageUseCases.SendContactMessageUseCase,
	sendStickerUseCase *messageUseCases.SendStickerMessageUseCase,
	sendButtonsUseCase *messageUseCases.SendButtonsMessageUseCase,
	sendListUseCase *messageUseCases.SendListMessageUseCase,
	sendPollUseCase *messageUseCases.SendPollMessageUseCase,
	editMessageUseCase *messageUseCases.EditMessageUseCase,
	deleteMessageUseCase *messageUseCases.DeleteMessageUseCase,
	reactMessageUseCase *messageUseCases.ReactMessageUseCase,
	logger logger.Logger,
) *MessageHandler {
	return &MessageHandler{
		sendTextUseCase:      sendTextUseCase,
		sendMediaUseCase:     sendMediaUseCase,
		sendLocationUseCase:  sendLocationUseCase,
		sendContactUseCase:   sendContactUseCase,
		sendStickerUseCase:   sendStickerUseCase,
		sendButtonsUseCase:   sendButtonsUseCase,
		sendListUseCase:      sendListUseCase,
		sendPollUseCase:      sendPollUseCase,
		editMessageUseCase:   editMessageUseCase,
		deleteMessageUseCase: deleteMessageUseCase,
		reactMessageUseCase:  reactMessageUseCase,
		logger:               logger,
	}
}

// SendTextMessage envia uma mensagem de texto
// @Summary Enviar mensagem de texto
// @Description Envia uma mensagem de texto para um número específico através de uma sessão ativa
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da mensagem de texto"
// @Success 200 {object} responses.SuccessResponse "Mensagem enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/text [post]
func (h *MessageHandler) SendTextMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendTextMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send text message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.sendTextUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send text message")
		responses.InternalError(w, "Failed to send text message")
		return
	}

	responses.Success(w, "Mensagem de texto enviada com sucesso", response)
}

// SendMediaMessage envia mídia unificada (imagem, áudio, vídeo, documento)
// @Summary Enviar mídia (imagem, áudio, vídeo, documento)
// @Description Envia mídia para um número específico. Suporta três formatos: JSON com URL/Base64, form-data para upload direto
// @Tags Mensagens
// @Accept json,multipart/form-data
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da mídia (JSON)" SchemaExample({"number": "5511999999999", "mediaType": "image", "media": "https://example.com/image.jpg", "caption": "Minha imagem"})
// @Success 200 {object} responses.SuccessResponse "Mídia enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/media [post]
func (h *MessageHandler) SendMediaMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	// Detectar tipo de conteúdo e processar adequadamente
	contentType := r.Header.Get("Content-Type")

	var req message.SendMediaMessageRequest

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Processar form-data para upload direto de arquivos
		req, err = h.parseFormDataMedia(r)
		if err != nil {
			h.logger.WithError(err).Error().Msg("Failed to parse form-data media request")
			responses.BadRequest(w, "Invalid form-data request", err.Error())
			return
		}
	} else {
		// Processar JSON com URL ou Base64
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.WithError(err).Error().Msg("Failed to decode send media message request")
			responses.BadRequest(w, "Invalid request body", err.Error())
			return
		}
	}

	// Validar e processar a mídia
	if err := h.validateAndProcessMedia(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Media validation failed")
		responses.BadRequest(w, "Invalid media data", err.Error())
		return
	}

	response, err := h.sendMediaUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send media message")
		responses.InternalError(w, "Failed to send media message")
		return
	}

	responses.Success(w, "Mídia enviada com sucesso", response)
}

// SendImageMessage envia uma imagem
// @Summary Enviar imagem
// @Description Envia uma imagem para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da imagem"
// @Success 200 {object} responses.SuccessResponse "Imagem enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/image [post]
func (h *MessageHandler) SendImageMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendImageMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send image message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	// Converter para SendMediaMessageRequest
	mediaReq := message.SendMediaMessageRequest{
		Number:      req.Number,
		MediaType:   "image",
		Media:       req.Image,
		Caption:     req.Caption,
		MimeType:    req.MimeType,
		ContextInfo: req.ContextInfo,
		Metadata:    req.Metadata,
	}

	response, err := h.sendMediaUseCase.Execute(r.Context(), sessionID, mediaReq)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send image message")
		responses.InternalError(w, "Failed to send image message")
		return
	}

	responses.Success(w, "Imagem enviada com sucesso", response)
}

// SendAudioMessage envia um áudio
// @Summary Enviar áudio
// @Description Envia um arquivo de áudio para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados do áudio"
// @Success 200 {object} responses.SuccessResponse "Áudio enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/audio [post]
func (h *MessageHandler) SendAudioMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendAudioMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send audio message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	// Converter para SendMediaMessageRequest
	mediaReq := message.SendMediaMessageRequest{
		Number:      req.Number,
		MediaType:   "audio",
		Media:       req.Audio,
		Caption:     req.Caption,
		MimeType:    "audio/mpeg", // Default para áudio
		ContextInfo: req.ContextInfo,
		Metadata:    req.Metadata,
	}

	response, err := h.sendMediaUseCase.Execute(r.Context(), sessionID, mediaReq)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send audio message")
		responses.InternalError(w, "Failed to send audio message")
		return
	}

	responses.Success(w, "Áudio enviado com sucesso", response)
}

// SendVideoMessage envia um vídeo
// @Summary Enviar vídeo
// @Description Envia um arquivo de vídeo para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados do vídeo"
// @Success 200 {object} responses.SuccessResponse "Vídeo enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/video [post]
func (h *MessageHandler) SendVideoMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendVideoMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send video message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	// Converter para SendMediaMessageRequest
	mediaReq := message.SendMediaMessageRequest{
		Number:      req.Number,
		MediaType:   "video",
		Media:       req.Video,
		Caption:     req.Caption,
		MimeType:    req.MimeType,
		ContextInfo: req.ContextInfo,
		Metadata:    req.Metadata,
	}

	response, err := h.sendMediaUseCase.Execute(r.Context(), sessionID, mediaReq)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send video message")
		responses.InternalError(w, "Failed to send video message")
		return
	}

	responses.Success(w, "Vídeo enviado com sucesso", response)
}

// SendDocumentMessage envia um documento
// @Summary Enviar documento
// @Description Envia um arquivo de documento para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados do documento"
// @Success 200 {object} responses.SuccessResponse "Documento enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/document [post]
func (h *MessageHandler) SendDocumentMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendDocumentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send document message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	// Converter para SendMediaMessageRequest
	mediaReq := message.SendMediaMessageRequest{
		Number:      req.Number,
		MediaType:   "document",
		Media:       req.Document,
		Caption:     req.Caption,
		FileName:    req.FileName,
		MimeType:    req.MimeType,
		ContextInfo: req.ContextInfo,
		Metadata:    req.Metadata,
	}

	response, err := h.sendMediaUseCase.Execute(r.Context(), sessionID, mediaReq)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send document message")
		responses.InternalError(w, "Failed to send document message")
		return
	}

	responses.Success(w, "Documento enviado com sucesso", response)
}

// SendLocationMessage envia uma localização
// @Summary Enviar localização
// @Description Envia uma localização (latitude e longitude) para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da localização"
// @Success 200 {object} responses.SuccessResponse "Localização enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/location [post]
func (h *MessageHandler) SendLocationMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendLocationMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send location message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.sendLocationUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send location message")
		responses.InternalError(w, "Failed to send location message")
		return
	}

	responses.Success(w, "Localização enviada com sucesso", response)
}

// SendContactMessage envia um contato
// @Summary Enviar contato
// @Description Envia informações de contato para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados do contato"
// @Success 200 {object} responses.SuccessResponse "Contato enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/contact [post]
func (h *MessageHandler) SendContactMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendContactMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send contact message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.sendContactUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send contact message")
		responses.InternalError(w, "Failed to send contact message")
		return
	}

	responses.Success(w, "Contato enviado com sucesso", response)
}

// SendStickerMessage envia um sticker
// @Summary Enviar sticker
// @Description Envia um sticker para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados do sticker"
// @Success 200 {object} responses.SuccessResponse "Sticker enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/sticker [post]
func (h *MessageHandler) SendStickerMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendStickerMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send sticker message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.sendStickerUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send sticker message")
		responses.InternalError(w, "Failed to send sticker message")
		return
	}

	responses.Success(w, "Sticker enviado com sucesso", response)
}

// SendButtonsMessage envia mensagem com botões
// @Summary Enviar mensagem com botões
// @Description Envia uma mensagem interativa com botões para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da mensagem com botões"
// @Success 200 {object} responses.SuccessResponse "Mensagem com botões enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/buttons [post]
func (h *MessageHandler) SendButtonsMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendButtonsMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send buttons message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.sendButtonsUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send buttons message")
		responses.InternalError(w, "Failed to send buttons message")
		return
	}

	responses.Success(w, "Mensagem com botões enviada com sucesso", response)
}

// SendListMessage envia mensagem com lista
// @Summary Enviar mensagem com lista
// @Description Envia uma mensagem interativa com lista de opções para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da mensagem com lista"
// @Success 200 {object} responses.SuccessResponse "Mensagem com lista enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/list [post]
func (h *MessageHandler) SendListMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendListMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send list message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.sendListUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send list message")
		responses.InternalError(w, "Failed to send list message")
		return
	}

	responses.Success(w, "Mensagem com lista enviada com sucesso", response)
}

// SendPollMessage envia enquete
// @Summary Enviar enquete
// @Description Envia uma enquete com opções de resposta para um número específico
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da enquete"
// @Success 200 {object} responses.SuccessResponse "Enquete enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/poll [post]
func (h *MessageHandler) SendPollMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.SendPollMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode send poll message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.sendPollUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to send poll message")
		responses.InternalError(w, "Failed to send poll message")
		return
	}

	responses.Success(w, "Enquete enviada com sucesso", response)
}

// EditMessage edita mensagem existente
// @Summary Editar mensagem
// @Description Edita o conteúdo de uma mensagem já enviada
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados para edição da mensagem"
// @Success 200 {object} responses.SuccessResponse "Mensagem editada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/send/edit [post]
func (h *MessageHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode edit message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.editMessageUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to edit message")
		responses.InternalError(w, "Failed to edit message")
		return
	}

	responses.Success(w, "Mensagem editada com sucesso", response)
}

// DeleteMessage deleta uma mensagem
// @Summary Deletar mensagem
// @Description Deleta uma mensagem específica do chat
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados para deletar a mensagem"
// @Success 200 {object} responses.SuccessResponse "Mensagem deletada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/delete [post]
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.DeleteMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode delete message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.deleteMessageUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to delete message")
		responses.InternalError(w, "Failed to delete message")
		return
	}

	responses.Success(w, "Mensagem deletada com sucesso", response)
}

// ReactMessage reage a uma mensagem
// @Summary Reagir a mensagem
// @Description Adiciona uma reação (emoji) a uma mensagem específica
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da reação"
// @Success 200 {object} responses.SuccessResponse "Reação enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /messages/{sessionID}/react [post]
func (h *MessageHandler) ReactMessage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req message.ReactMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode react message request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	response, err := h.reactMessageUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to react to message")
		responses.InternalError(w, "Failed to react to message")
		return
	}

	responses.Success(w, "Reação enviada com sucesso", response)
}

// parseFormDataMedia processa form-data para upload direto de arquivos
func (h *MessageHandler) parseFormDataMedia(r *http.Request) (message.SendMediaMessageRequest, error) {
	var req message.SendMediaMessageRequest

	// Parse multipart form (32MB max)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return req, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Campos obrigatórios
	req.Number = r.FormValue("number")
	req.MediaType = r.FormValue("mediaType")

	// Campos opcionais
	req.Caption = r.FormValue("caption")
	req.FileName = r.FormValue("fileName")
	req.MimeType = r.FormValue("mimeType")

	// Processar arquivo de mídia
	file, header, err := r.FormFile("media")
	if err != nil {
		return req, fmt.Errorf("failed to get media file: %w", err)
	}
	defer file.Close()

	// Ler conteúdo do arquivo
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return req, fmt.Errorf("failed to read file content: %w", err)
	}

	// Converter para Base64
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	req.Media = fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(fileBytes))

	// Se fileName não foi fornecido, usar o nome do arquivo enviado
	if req.FileName == "" {
		req.FileName = header.Filename
	}

	// Se mimeType não foi fornecido, usar o detectado
	if req.MimeType == "" {
		req.MimeType = mimeType
	}

	return req, nil
}

// validateAndProcessMedia valida e processa os dados de mídia
func (h *MessageHandler) validateAndProcessMedia(req *message.SendMediaMessageRequest) error {
	// Validar campos obrigatórios
	if req.Number == "" {
		return fmt.Errorf("number is required")
	}

	if req.MediaType == "" {
		return fmt.Errorf("mediaType is required")
	}

	if req.Media == "" {
		return fmt.Errorf("media is required")
	}

	// Validar mediaType
	validMediaTypes := map[string]bool{
		"image":    true,
		"audio":    true,
		"video":    true,
		"document": true,
	}

	if !validMediaTypes[req.MediaType] {
		return fmt.Errorf("invalid mediaType: must be one of image, audio, video, document")
	}

	// Para documentos, fileName é obrigatório
	if req.MediaType == "document" && req.FileName == "" {
		return fmt.Errorf("fileName is required for document mediaType")
	}

	// Detectar e validar mimeType se não fornecido
	if req.MimeType == "" {
		detectedMimeType, err := h.detectMimeType(req.Media, req.FileName)
		if err != nil {
			return fmt.Errorf("failed to detect mime type: %w", err)
		}
		req.MimeType = detectedMimeType
	}

	// Validar se o mimeType corresponde ao mediaType
	if err := h.validateMimeTypeMatch(req.MediaType, req.MimeType); err != nil {
		return err
	}

	return nil
}

// detectMimeType detecta o tipo MIME baseado no conteúdo ou extensão do arquivo
func (h *MessageHandler) detectMimeType(media, fileName string) (string, error) {
	// Se é uma URL, tentar detectar pela extensão da URL
	if strings.HasPrefix(media, "http://") || strings.HasPrefix(media, "https://") {
		// Extrair extensão da URL
		ext := filepath.Ext(media)
		if ext != "" {
			mimeType := mime.TypeByExtension(ext)
			if mimeType != "" {
				return mimeType, nil
			}
		}

		// Se fileName foi fornecido, usar sua extensão
		if fileName != "" {
			ext := filepath.Ext(fileName)
			mimeType := mime.TypeByExtension(ext)
			if mimeType != "" {
				return mimeType, nil
			}
		}

		// Para URLs sem extensão clara, retornar um tipo genérico
		// que será validado posteriormente
		return "application/octet-stream", nil
	}

	// Se é Base64, tentar extrair o mimeType do data URL
	if strings.HasPrefix(media, "data:") {
		parts := strings.Split(media, ";")
		if len(parts) > 0 {
			mimeType := strings.TrimPrefix(parts[0], "data:")
			if mimeType != "" {
				return mimeType, nil
			}
		}
	}

	// Fallback: detectar pela extensão do fileName
	if fileName != "" {
		ext := filepath.Ext(fileName)
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			return mimeType, nil
		}
	}

	return "application/octet-stream", nil
}

// validateMimeTypeMatch valida se o mimeType corresponde ao mediaType
func (h *MessageHandler) validateMimeTypeMatch(mediaType, mimeType string) error {
	// Se o mimeType é application/octet-stream (detectado automaticamente para URLs),
	// não validamos rigorosamente - deixamos o WhatsApp decidir
	if mimeType == "application/octet-stream" {
		return nil
	}

	switch mediaType {
	case "image":
		if !strings.HasPrefix(mimeType, "image/") {
			return fmt.Errorf("mimeType %s does not match mediaType %s", mimeType, mediaType)
		}
	case "audio":
		if !strings.HasPrefix(mimeType, "audio/") {
			return fmt.Errorf("mimeType %s does not match mediaType %s", mimeType, mediaType)
		}
	case "video":
		if !strings.HasPrefix(mimeType, "video/") {
			return fmt.Errorf("mimeType %s does not match mediaType %s", mimeType, mediaType)
		}
	case "document":
		// Documentos podem ter vários tipos MIME, então não validamos especificamente
		// Apenas verificamos que não é image, audio ou video
		if strings.HasPrefix(mimeType, "image/") ||
			strings.HasPrefix(mimeType, "audio/") ||
			strings.HasPrefix(mimeType, "video/") {
			return fmt.Errorf("mimeType %s is not valid for document mediaType", mimeType)
		}
	}

	return nil
}
