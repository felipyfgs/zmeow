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
// @Description Envia uma mensagem de texto para um número específico através de uma sessão ativa do WhatsApp
// @Description
// @Description **Exemplo de uso:**
// @Description ```json
// @Description {
// @Description   "to": "559981769536",
// @Description   "text": "Olá! Como você está?",
// @Description   "contextInfo": {
// @Description     "mentionedJids": ["559987654321@s.whatsapp.net"]
// @Description   }
// @Description }
// @Description ```
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendTextMessageRequest true "Dados da mensagem de texto"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Mensagem enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos ou campos obrigatórios ausentes"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Envia mídia para um número específico. Suporta JSON com URL/Base64 ou form-data para upload direto
// @Description
// @Description **Tipos de mídia suportados:**
// @Description - `image`: Imagens (JPEG, PNG, WebP)
// @Description - `audio`: Áudios (MP3, OGG, WAV)
// @Description - `video`: Vídeos (MP4, AVI, MOV)
// @Description - `document`: Documentos (PDF, DOC, TXT, etc.)
// @Description
// @Description **Duas formas de envio:**
// @Description 1. JSON com Base64 ou URL
// @Description 2. Form-data com upload de arquivo
// @Tags Mensagens
// @Accept json,multipart/form-data
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendMediaMessageRequest true "Dados da mídia (para JSON)"
// @Param to formData string false "Número do destinatário ou JID do grupo (obrigatório para form-data)" example("559981769536")
// @Param mediaType formData string false "Tipo de mídia (obrigatório para form-data)" Enums(image, audio, video, document) example("image")
// @Param media formData file false "Arquivo de mídia (obrigatório para form-data)"
// @Param caption formData string false "Legenda da mídia (opcional para form-data)" example("Minha foto")
// @Param fileName formData string false "Nome do arquivo (obrigatório para documentos)" example("documento.pdf")
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Mídia enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, tipo de mídia não suportado ou arquivo muito grande"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Envia uma imagem para um número específico. Aceita URL pública ou dados Base64 no formato data:image/type;base64,data
// @Description
// @Description **Formatos suportados:** JPEG, PNG, WebP, GIF
// @Description **Tamanho máximo:** 16MB
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendImageMessageRequest true "Dados da imagem"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Imagem enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, formato não suportado ou imagem muito grande"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
		GroupJid:    req.GroupJid,
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
// @Description Envia um arquivo de áudio para um número específico. Aceita URL pública ou dados Base64
// @Description
// @Description **Formatos suportados:** MP3, OGG, WAV, M4A
// @Description **PTT (Push to Talk):** true = mensagem de voz, false = áudio normal
// @Description **Tamanho máximo:** 16MB
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendAudioMessageRequest true "Dados do áudio"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Áudio enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, formato não suportado ou áudio muito grande"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
		GroupJid:    req.GroupJid,
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
// @Description Envia um arquivo de vídeo para um número específico. Aceita URL pública ou dados Base64 no formato data:video/type;base64,data
// @Description
// @Description **Formatos suportados:** MP4, AVI, MOV, MKV
// @Description **Tamanho máximo:** 64MB
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendVideoMessageRequest true "Dados do vídeo"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Vídeo enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, formato não suportado ou vídeo muito grande"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
		GroupJid:    req.GroupJid,
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
// @Description Envia um arquivo de documento para um número específico. Aceita URL pública ou dados Base64. O campo fileName é obrigatório
// @Description
// @Description **Formatos suportados:** PDF, DOC, DOCX, XLS, XLSX, TXT, etc.
// @Description **Tamanho máximo:** 100MB
// @Description **Campos obrigatórios:** number, document, fileName
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendDocumentMessageRequest true "Dados do documento"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Documento enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, fileName ausente ou documento muito grande"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
		GroupJid:    req.GroupJid,
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
// @Description Envia coordenadas de localização (latitude e longitude) para um número específico
// @Description
// @Description **Campos obrigatórios:** number, latitude, longitude
// @Description **Campos opcionais:** name (nome do local), address (endereço)
// @Description **Formato das coordenadas:** Decimais (ex: -23.550520, -46.633309)
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendLocationMessageRequest true "Dados da localização"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Localização enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos ou coordenadas fora do intervalo válido"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Envia informações de contato para um número específico. Permite compartilhar dados de contato do WhatsApp
// @Description
// @Description **Campos obrigatórios:** number, contactName, contactJID
// @Description **Formato do contactJID:** número@s.whatsapp.net (ex: 559987654321@s.whatsapp.net)
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendContactMessageRequest true "Dados do contato"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Contato enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos ou formato de JID incorreto"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Envia um sticker para um número específico. Aceita URL pública ou dados Base64 preferencialmente no formato WebP
// @Description
// @Description **Formato recomendado:** WebP (data:image/webp;base64,data)
// @Description **Outros formatos aceitos:** PNG, JPEG
// @Description **Tamanho máximo:** 1MB
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendStickerMessageRequest true "Dados do sticker"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Sticker enviado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, formato não suportado ou sticker muito grande"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Envia uma mensagem interativa com botões para um número específico. Ideal para opções rápidas
// @Description
// @Description **Limites:** Mínimo 1, máximo 3 botões por mensagem
// @Description **Campos obrigatórios:** number, text, buttons
// @Description **Cada botão precisa:** id (único), displayText
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendButtonsMessageRequest true "Dados da mensagem com botões"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Mensagem com botões enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, muitos botões ou IDs duplicados"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Envia uma mensagem interativa com lista de opções organizadas em seções. Ideal para muitas opções
// @Description
// @Description **Estrutura:** Seções -> Linhas (itens)
// @Description **Campos obrigatórios:** number, text, title, buttonText, sections
// @Description **Cada seção precisa:** title, rows (mínimo 1)
// @Description **Cada linha precisa:** id (único), title
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendListMessageRequest true "Dados da mensagem com lista"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Mensagem com lista enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, seções vazias ou IDs duplicados"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Envia uma enquete (poll) com opções de resposta para um número específico
// @Description
// @Description **Limites:** Mínimo 2, máximo 12 opções por enquete
// @Description **Campos obrigatórios:** number, name (pergunta), options, selectableOptionsCount
// @Description **selectableOptionsCount:** quantas opções o usuário pode escolher (1 = múltipla escolha, >1 = seleção múltipla)
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.SendPollMessageRequest true "Dados da enquete"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Enquete enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, poucas/muitas opções ou selectableOptionsCount inválido"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou não conectada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha no envio"
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
// @Description Edita o conteúdo de uma mensagem já enviada. Funciona apenas para mensagens de texto
// @Description
// @Description **Limitações:** Apenas mensagens de texto podem ser editadas
// @Description **Tempo limite:** Mensagens podem ser editadas dentro de 15 minutos após o envio
// @Description **Campos obrigatórios:** number, id (da mensagem), newText
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.EditMessageRequest true "Dados para edição da mensagem"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Mensagem editada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos, mensagem não editável ou tempo limite excedido"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou mensagem não encontrada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha na edição"
// @Router /messages/{sessionID}/edit [post]
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
// @Description Deleta uma mensagem específica do chat com opção de deletar para todos ou apenas para você
// @Description
// @Description **Tipos de deleção:**
// @Description - forMe=true: Deleta apenas para você
// @Description - forMe=false: Deleta para todos (padrão)
// @Description **Tempo limite:** Mensagens podem ser deletadas para todos dentro de 7 minutos
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.DeleteMessageRequest true "Dados para deletar a mensagem"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Mensagem deletada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos ou tempo limite excedido para deletar para todos"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou mensagem não encontrada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha na deleção"
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
// @Description Adiciona ou remove uma reação (emoji) a uma mensagem específica
// @Description
// @Description **Adição de reação:** Informe o emoji no campo reaction
// @Description **Remoção de reação:** Use string vazia ("") no campo reaction
// @Description **Emojis suportados:** 👍, ❤️, 😂, 😮, 😢, 😡, etc.
// @Description **Prefixo \"me:\":** Use para reagir às suas próprias mensagens
// @Tags Mensagens
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão WhatsApp (UUID)" format(uuid) example("9a3a24d2-2b2c-4214-8797-7c6571837f53")
// @Param request body message.ReactMessageRequest true "Dados da reação"
// @Success 200 {object} responses.SuccessResponse{data=message.SendMessageResponse} "Reação enviada com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Parâmetros inválidos ou emoji não suportado"
// @Failure 404 {object} responses.ErrorResponse "Sessão não encontrada ou mensagem não encontrada"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor ou falha na reação"
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
	req.GroupJid = r.FormValue("groupJid")
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
	// Validar campos obrigatórios - pelo menos um deve estar preenchido
	if req.Number == "" && req.GroupJid == "" {
		return fmt.Errorf("either number or groupJid must be provided")
	}

	if req.Number != "" && req.GroupJid != "" {
		return fmt.Errorf("only one of number or groupJid should be provided")
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
