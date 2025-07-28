package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"zmeow/internal/http/responses"
	"zmeow/internal/usecases/session"
	"zmeow/pkg/logger"
)

// SessionHandler implementa os handlers para sessões
type SessionHandler struct {
	createUseCase     *session.CreateSessionUseCase
	listUseCase       *session.ListSessionsUseCase
	getUseCase        *session.GetSessionUseCase
	deleteUseCase     *session.DeleteSessionUseCase
	connectUseCase    *session.ConnectSessionUseCase
	disconnectUseCase *session.DisconnectSessionUseCase
	qrUseCase         *session.GetQRCodeUseCase
	pairUseCase       *session.PairPhoneUseCase
	proxyUseCase      *session.SetProxyUseCase
	statusUseCase     *session.GetStatusUseCase
	logger            logger.Logger
}

// NewSessionHandler cria uma nova instância do session handler
func NewSessionHandler(
	createUseCase *session.CreateSessionUseCase,
	listUseCase *session.ListSessionsUseCase,
	getUseCase *session.GetSessionUseCase,
	deleteUseCase *session.DeleteSessionUseCase,
	connectUseCase *session.ConnectSessionUseCase,
	disconnectUseCase *session.DisconnectSessionUseCase,
	qrUseCase *session.GetQRCodeUseCase,
	pairUseCase *session.PairPhoneUseCase,
	proxyUseCase *session.SetProxyUseCase,
	statusUseCase *session.GetStatusUseCase,
	logger logger.Logger,
) *SessionHandler {
	return &SessionHandler{
		createUseCase:     createUseCase,
		listUseCase:       listUseCase,
		getUseCase:        getUseCase,
		deleteUseCase:     deleteUseCase,
		connectUseCase:    connectUseCase,
		disconnectUseCase: disconnectUseCase,
		qrUseCase:         qrUseCase,
		pairUseCase:       pairUseCase,
		proxyUseCase:      proxyUseCase,
		statusUseCase:     statusUseCase,
		logger:            logger.WithComponent("session-handler"),
	}
}

// AddSession cria uma nova sessão
// @Summary      Criar Nova Sessão
// @Description  Cria uma nova sessão WhatsApp com nome único
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param request body object true  "Dados da sessão"
// @Success      201      {object}  responses.CreatedResponse  "Sessão criada com sucesso"
// @Failure      400      {object}  responses.ErrorResponse  "Dados inválidos"
// @Failure      409      {object}  responses.ErrorResponse  "Sessão já existe"
// @Failure      500      {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/add [post]
func (h *SessionHandler) AddSession(w http.ResponseWriter, r *http.Request) {
	var req session.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode create session request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	sess, err := h.createUseCase.Execute(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to create session")
		responses.InternalError(w, "Failed to create session")
		return
	}

	responses.Created(w, "Sessão criada com sucesso", sess)
}

// ListSessions lista todas as sessões
// @Summary      Listar Sessões
// @Description  Lista todas as sessões WhatsApp cadastradas
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Success      200  {object}  responses.SuccessResponse  "Lista de sessões"
// @Failure      500  {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/list [get]
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.listUseCase.Execute(r.Context())
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to list sessions")
		responses.InternalError(w, "Failed to list sessions")
		return
	}

	responses.Success(w, "Sessões listadas com sucesso", sessions)
}

// GetSession obtém uma sessão específica
// @Summary      Obter Sessão
// @Description  Obtém informações de uma sessão específica pelo ID
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string  true  "ID da sessão (UUID)"
// @Success      200        {object}  responses.SuccessResponse  "Sessão encontrada"
// @Failure      400        {object}  responses.ErrorResponse  "ID inválido"
// @Failure      404        {object}  responses.ErrorResponse  "Sessão não encontrada"
// @Failure      500        {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/{sessionID} [get]
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	sess, err := h.getUseCase.Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to get session")
		responses.NotFound(w, "Session not found")
		return
	}

	responses.Success(w, "Sessão encontrada", sess)
}

// DeleteSession remove uma sessão
// @Summary      Deletar Sessão
// @Description  Remove uma sessão WhatsApp permanentemente
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string  true  "ID da sessão (UUID)"
// @Success      200        {object}  responses.SuccessResponse  "Sessão removida com sucesso"
// @Failure      400        {object}  responses.ErrorResponse  "ID inválido"
// @Failure      404        {object}  responses.ErrorResponse  "Sessão não encontrada"
// @Failure      500        {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/{sessionID} [delete]
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	err = h.deleteUseCase.Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to delete session")
		responses.InternalError(w, "Failed to delete session")
		return
	}

	responses.Success(w, "Sessão removida com sucesso", nil)
}

// ConnectSession conecta uma sessão
// @Summary      Conectar Sessão
// @Description  Inicia conexão de uma sessão WhatsApp
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string  true  "ID da sessão (UUID)"
// @Success      200        {object}  responses.SuccessResponse  "Conexão iniciada"
// @Failure      400        {object}  responses.ErrorResponse  "ID inválido"
// @Failure      404        {object}  responses.ErrorResponse  "Sessão não encontrada"
// @Failure      500        {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/{sessionID}/connect [post]
func (h *SessionHandler) ConnectSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	result, err := h.connectUseCase.Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to connect session")
		responses.InternalError(w, "Failed to connect session")
		return
	}

	responses.Success(w, "Conexão iniciada", result)
}

// LogoutSession desconecta uma sessão
// @Summary      Logout da Sessão
// @Description  Realiza logout de uma sessão WhatsApp
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string  true  "ID da sessão (UUID)"
// @Success      200        {object}  responses.SuccessResponse  "Logout realizado com sucesso"
// @Failure      400        {object}  responses.ErrorResponse  "ID inválido"
// @Failure      404        {object}  responses.ErrorResponse  "Sessão não encontrada"
// @Failure      500        {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/{sessionID}/logout [post]
func (h *SessionHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	result, err := h.disconnectUseCase.Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to logout session")
		responses.InternalError(w, "Failed to logout session")
		return
	}

	responses.Success(w, "Logout realizado com sucesso", result)
}

// GetSessionStatus obtém o status de uma sessão
// @Summary      Status da Sessão
// @Description  Obtém o status atual de uma sessão WhatsApp
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string  true  "ID da sessão (UUID)"
// @Success      200        {object}  responses.SuccessResponse  "Status da sessão"
// @Failure      400        {object}  responses.ErrorResponse  "ID inválido"
// @Failure      404        {object}  responses.ErrorResponse  "Sessão não encontrada"
// @Failure      500        {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/{sessionID}/status [get]
func (h *SessionHandler) GetSessionStatus(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	status, err := h.statusUseCase.Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to get session status")
		responses.NotFound(w, "Session not found")
		return
	}

	responses.Success(w, "Status obtido", status)
}

// GetQRCode obtém o QR code de uma sessão
// @Summary      QR Code da Sessão
// @Description  Obtém o QR Code para autenticação da sessão WhatsApp
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string  true  "ID da sessão (UUID)"
// @Success      200        {object}  responses.SuccessResponse  "QR Code obtido"
// @Failure      400        {object}  responses.ErrorResponse  "ID inválido"
// @Failure      404        {object}  responses.ErrorResponse  "Sessão não encontrada"
// @Failure      500        {object}  responses.ErrorResponse  "Erro interno"
// @Router       /sessions/{sessionID}/qr [get]
func (h *SessionHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	result, err := h.qrUseCase.Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to get QR code")
		responses.InternalError(w, "Failed to get QR code")
		return
	}

	if result.QRCode != "" {
		responses.Success(w, "QR Code gerado e exibido no terminal do servidor", result)
	} else {
		if result.Status == "connected" {
			responses.Success(w, "Sessão já está conectada. QR Code não é necessário.", result)
		} else {
			responses.Success(w, "Nenhum QR Code disponível. Conecte a sessão primeiro para gerar um novo QR Code.", result)
		}
	}
}

// PairPhone realiza pareamento por telefone
// @Summary      Pareamento por Telefone
// @Description  Realiza pareamento da sessão WhatsApp usando número de telefone
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID    path      string                           true   "ID da sessão (UUID)"
// @Param request body object true   "Dados do pareamento"
// @Success      200          {object}  responses.SuccessResponse  "Pareamento iniciado"
// @Failure      400          {object}  responses.ErrorResponse          "Dados inválidos"
// @Failure      404          {object}  responses.ErrorResponse          "Sessão não encontrada"
// @Failure      500          {object}  responses.ErrorResponse          "Erro interno"
// @Router       /sessions/{sessionID}/pairphone [post]
func (h *SessionHandler) PairPhone(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req struct {
		PhoneNumber string `json:"phoneNumber" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode pair phone request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	result, err := h.pairUseCase.Execute(r.Context(), sessionID, req.PhoneNumber)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to pair phone")
		responses.InternalError(w, "Failed to pair phone")
		return
	}

	responses.Success(w, "Código de pareamento enviado", result)
}

// SetProxy configura proxy para uma sessão
// @Summary      Configurar Proxy
// @Description  Configura proxy para uma sessão WhatsApp
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        sessionID  path      string                    true  "ID da sessão (UUID)"
// @Param request body object true  "Configurações do proxy"
// @Success      200        {object}  responses.SuccessResponse  "Proxy configurado"
// @Failure      400        {object}  responses.ErrorResponse   "Dados inválidos"
// @Failure      404        {object}  responses.ErrorResponse   "Sessão não encontrada"
// @Failure      500        {object}  responses.ErrorResponse   "Erro interno"
// @Router       /sessions/{sessionID}/proxy/set [post]
func (h *SessionHandler) SetProxy(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Invalid session ID format", err.Error())
		return
	}

	var req session.SetProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode set proxy request")
		responses.BadRequest(w, "Invalid request body", err.Error())
		return
	}

	result, err := h.proxyUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to set proxy")
		responses.InternalError(w, "Failed to set proxy")
		return
	}

	responses.Success(w, "Proxy configurado com sucesso", result)
}
