package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.mau.fi/whatsmeow/types"

	"zmeow/internal/http/responses"
	messageUsecases "zmeow/internal/usecases/message"
	"zmeow/pkg/logger"
)

// ChatHandler gerencia operações específicas de chat
type ChatHandler struct {
	logger               logger.Logger
	sendTextUseCase      *messageUsecases.SendTextMessageUseCase
	sendMediaUseCase     *messageUsecases.SendMediaMessageUseCase
	sendLocationUseCase  *messageUsecases.SendLocationMessageUseCase
	sendContactUseCase   *messageUsecases.SendContactMessageUseCase
	sendStickerUseCase   *messageUsecases.SendStickerMessageUseCase
	sendButtonsUseCase   *messageUsecases.SendButtonsMessageUseCase
	sendListUseCase      *messageUsecases.SendListMessageUseCase
	sendPollUseCase      *messageUsecases.SendPollMessageUseCase
	editMessageUseCase   *messageUsecases.EditMessageUseCase
	deleteMessageUseCase *messageUsecases.DeleteMessageUseCase
	reactMessageUseCase  *messageUsecases.ReactMessageUseCase
}

// NewChatHandler cria uma nova instância do ChatHandler
func NewChatHandler(
	sendTextUseCase *messageUsecases.SendTextMessageUseCase,
	sendMediaUseCase *messageUsecases.SendMediaMessageUseCase,
	sendLocationUseCase *messageUsecases.SendLocationMessageUseCase,
	sendContactUseCase *messageUsecases.SendContactMessageUseCase,
	sendStickerUseCase *messageUsecases.SendStickerMessageUseCase,
	sendButtonsUseCase *messageUsecases.SendButtonsMessageUseCase,
	sendListUseCase *messageUsecases.SendListMessageUseCase,
	sendPollUseCase *messageUsecases.SendPollMessageUseCase,
	editMessageUseCase *messageUsecases.EditMessageUseCase,
	deleteMessageUseCase *messageUsecases.DeleteMessageUseCase,
	reactMessageUseCase *messageUsecases.ReactMessageUseCase,
) *ChatHandler {
	return &ChatHandler{
		logger:               logger.WithComponent("chat-handler"),
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
	}
}

// ChatPresenceRequest representa a requisição para definir presença no chat
type ChatPresenceRequest struct {
	Phone string                  `json:"phone" validate:"required"`
	State string                  `json:"state" validate:"required,oneof=composing recording paused"`
	Media types.ChatPresenceMedia `json:"media,omitempty"`
}

// MarkReadRequest representa a requisição para marcar mensagens como lidas
type MarkReadRequest struct {
	MessageIDs []string  `json:"message_ids" validate:"required"`
	Chat       types.JID `json:"chat" validate:"required"`
	Sender     types.JID `json:"sender,omitempty"`
}

// DownloadMediaRequest representa a requisição para download de mídia
type DownloadMediaRequest struct {
	MessageID string `json:"message_id" validate:"required"`
	Phone     string `json:"phone" validate:"required"`
}

// SendChatPresence define a presença no chat (digitando, gravando, etc.)
// @Summary Definir presença no chat
// @Description Define o status de presença no chat (digitando, gravando áudio, pausado)
// @Tags Chat
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados da presença no chat"
// @Success 200 {object} responses.SuccessResponse "Presença definida com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /chat/{sessionID}/presence [post]
func (h *ChatHandler) SendChatPresence(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		responses.Error400(w, "Session ID é obrigatório", "SESSION_ID_REQUIRED", "Session ID é obrigatório")
		return
	}

	var req ChatPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode chat presence request")
		responses.Error400(w, "Dados da requisição inválidos", "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Implementar lógica de presença no chat
	// Aqui você implementaria a lógica usando o whatsmeow client

	h.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"phone":      req.Phone,
		"state":      req.State,
	}).Info().Msg("Chat presence request")

	responses.Success200(w, "Operação realizada com sucesso", map[string]interface{}{
		"success": true,
		"message": "Presença no chat definida com sucesso",
		"data": map[string]interface{}{
			"session_id": sessionID,
			"phone":      req.Phone,
			"state":      req.State,
		},
	})
}

// MarkAsRead marca mensagens como lidas
// @Summary Marcar mensagens como lidas
// @Description Marca uma ou mais mensagens como lidas em um chat específico
// @Tags Chat
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados das mensagens para marcar como lidas"
// @Success 200 {object} responses.SuccessResponse "Mensagens marcadas como lidas com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /chat/{sessionID}/markread [post]
func (h *ChatHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		responses.Error400(w, "Session ID é obrigatório", "SESSION_ID_REQUIRED", "Session ID é obrigatório")
		return
	}

	var req MarkReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode mark read request")
		responses.Error400(w, "Dados da requisição inválidos", "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Implementar lógica de marcar como lida
	// Aqui você implementaria a lógica usando o whatsmeow client

	h.logger.WithFields(map[string]interface{}{
		"session_id":  sessionID,
		"message_ids": req.MessageIDs,
		"chat":        req.Chat.String(),
	}).Info().Msg("Mark messages as read request")

	responses.Success200(w, "Operação realizada com sucesso", map[string]interface{}{
		"success": true,
		"message": "Mensagens marcadas como lidas com sucesso",
		"data": map[string]interface{}{
			"session_id":  sessionID,
			"message_ids": req.MessageIDs,
			"chat":        req.Chat.String(),
		},
	})
}

// DownloadImage faz download de uma imagem
// @Summary Download de imagem
// @Description Faz download de uma imagem de uma mensagem específica
// @Tags Chat
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados para download da imagem"
// @Success 200 {object} responses.SuccessResponse "Download iniciado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /chat/{sessionID}/downloadimage [post]
func (h *ChatHandler) DownloadImage(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		responses.Error400(w, "Session ID é obrigatório", "SESSION_ID_REQUIRED", "Session ID é obrigatório")
		return
	}

	var req DownloadMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode download image request")
		responses.Error400(w, "Dados da requisição inválidos", "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Implementar lógica de download de imagem
	// Aqui você implementaria a lógica usando o whatsmeow client

	h.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"message_id": req.MessageID,
		"phone":      req.Phone,
	}).Info().Msg("Download image request")

	responses.Success200(w, "Operação realizada com sucesso", map[string]interface{}{
		"success": true,
		"message": "Download de imagem iniciado",
		"data": map[string]interface{}{
			"session_id": sessionID,
			"message_id": req.MessageID,
			"phone":      req.Phone,
			"type":       "image",
		},
	})
}

// DownloadVideo faz download de um vídeo
// @Summary Download de vídeo
// @Description Faz download de um vídeo de uma mensagem específica
// @Tags Chat
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados para download do vídeo"
// @Success 200 {object} responses.SuccessResponse "Download iniciado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /chat/{sessionID}/downloadvideo [post]
func (h *ChatHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		responses.Error400(w, "Session ID é obrigatório", "SESSION_ID_REQUIRED", "Session ID é obrigatório")
		return
	}

	var req DownloadMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode download video request")
		responses.Error400(w, "Dados da requisição inválidos", "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Implementar lógica de download de vídeo
	// Aqui você implementaria a lógica usando o whatsmeow client

	h.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"message_id": req.MessageID,
		"phone":      req.Phone,
	}).Info().Msg("Download video request")

	responses.Success200(w, "Operação realizada com sucesso", map[string]interface{}{
		"success": true,
		"message": "Download de vídeo iniciado",
		"data": map[string]interface{}{
			"session_id": sessionID,
			"message_id": req.MessageID,
			"phone":      req.Phone,
			"type":       "video",
		},
	})
}

// DownloadAudio faz download de um áudio
// @Summary Download de áudio
// @Description Faz download de um áudio de uma mensagem específica
// @Tags Chat
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados para download do áudio"
// @Success 200 {object} responses.SuccessResponse "Download iniciado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /chat/{sessionID}/downloadaudio [post]
func (h *ChatHandler) DownloadAudio(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		responses.Error400(w, "Session ID é obrigatório", "SESSION_ID_REQUIRED", "Session ID é obrigatório")
		return
	}

	var req DownloadMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode download audio request")
		responses.Error400(w, "Dados da requisição inválidos", "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Implementar lógica de download de áudio
	// Aqui você implementaria a lógica usando o whatsmeow client

	h.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"message_id": req.MessageID,
		"phone":      req.Phone,
	}).Info().Msg("Download audio request")

	responses.Success200(w, "Operação realizada com sucesso", map[string]interface{}{
		"success": true,
		"message": "Download de áudio iniciado",
		"data": map[string]interface{}{
			"session_id": sessionID,
			"message_id": req.MessageID,
			"phone":      req.Phone,
			"type":       "audio",
		},
	})
}

// DownloadDocument faz download de um documento
// @Summary Download de documento
// @Description Faz download de um documento de uma mensagem específica
// @Tags Chat
// @Accept json
// @Produce json
// @Param sessionID path string true "ID da sessão (UUID)"
// @Param request body object true "Dados para download do documento"
// @Success 200 {object} responses.SuccessResponse "Download iniciado com sucesso"
// @Failure 400 {object} responses.ErrorResponse "Dados inválidos"
// @Failure 500 {object} responses.ErrorResponse "Erro interno do servidor"
// @Router /chat/{sessionID}/downloaddocument [post]
func (h *ChatHandler) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		responses.Error400(w, "Session ID é obrigatório", "SESSION_ID_REQUIRED", "Session ID é obrigatório")
		return
	}

	var req DownloadMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode download document request")
		responses.Error400(w, "Dados da requisição inválidos", "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Implementar lógica de download de documento
	// Aqui você implementaria a lógica usando o whatsmeow client

	h.logger.WithFields(map[string]interface{}{
		"session_id": sessionID,
		"message_id": req.MessageID,
		"phone":      req.Phone,
	}).Info().Msg("Download document request")

	responses.Success200(w, "Operação realizada com sucesso", map[string]interface{}{
		"success": true,
		"message": "Download de documento iniciado",
		"data": map[string]interface{}{
			"session_id": sessionID,
			"message_id": req.MessageID,
			"phone":      req.Phone,
			"type":       "document",
		},
	})
}
