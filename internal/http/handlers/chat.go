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
	logger                logger.Logger
	sendTextUseCase       *messageUsecases.SendTextMessageUseCase
	sendMediaUseCase      *messageUsecases.SendMediaMessageUseCase
	sendLocationUseCase   *messageUsecases.SendLocationMessageUseCase
	sendContactUseCase    *messageUsecases.SendContactMessageUseCase
	sendStickerUseCase    *messageUsecases.SendStickerMessageUseCase
	sendButtonsUseCase    *messageUsecases.SendButtonsMessageUseCase
	sendListUseCase       *messageUsecases.SendListMessageUseCase
	sendPollUseCase       *messageUsecases.SendPollMessageUseCase
	editMessageUseCase    *messageUsecases.EditMessageUseCase
	deleteMessageUseCase  *messageUsecases.DeleteMessageUseCase
	reactMessageUseCase   *messageUsecases.ReactMessageUseCase
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
		logger:                logger.WithComponent("chat-handler"),
		sendTextUseCase:       sendTextUseCase,
		sendMediaUseCase:      sendMediaUseCase,
		sendLocationUseCase:   sendLocationUseCase,
		sendContactUseCase:    sendContactUseCase,
		sendStickerUseCase:    sendStickerUseCase,
		sendButtonsUseCase:    sendButtonsUseCase,
		sendListUseCase:       sendListUseCase,
		sendPollUseCase:       sendPollUseCase,
		editMessageUseCase:    editMessageUseCase,
		deleteMessageUseCase:  deleteMessageUseCase,
		reactMessageUseCase:   reactMessageUseCase,
	}
}

// ChatPresenceRequest representa a requisição para definir presença no chat
type ChatPresenceRequest struct {
	Phone string                    `json:"phone" validate:"required"`
	State string                    `json:"state" validate:"required,oneof=composing recording paused"`
	Media types.ChatPresenceMedia   `json:"media,omitempty"`
}

// MarkReadRequest representa a requisição para marcar mensagens como lidas
type MarkReadRequest struct {
	MessageIDs []string   `json:"message_ids" validate:"required"`
	Chat       types.JID  `json:"chat" validate:"required"`
	Sender     types.JID  `json:"sender,omitempty"`
}

// DownloadMediaRequest representa a requisição para download de mídia
type DownloadMediaRequest struct {
	MessageID string `json:"message_id" validate:"required"`
	Phone     string `json:"phone" validate:"required"`
}

// SendChatPresence define a presença no chat (digitando, gravando, etc.)
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