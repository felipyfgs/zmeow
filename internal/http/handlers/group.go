package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"zmeow/internal/domain/group"
	"zmeow/internal/http/responses"
	groupUseCases "zmeow/internal/usecases/group"
	"zmeow/pkg/logger"
)

// GroupHandler implementa os handlers para grupos
type GroupHandler struct {
	createGroupUC               *groupUseCases.CreateGroupUseCase
	listGroupsUC                *groupUseCases.ListGroupsUseCase
	getGroupInfoUC              *groupUseCases.GetGroupInfoUseCase
	updateParticipantsUC        *groupUseCases.UpdateParticipantsUseCase
	leaveGroupUC                *groupUseCases.LeaveGroupUseCase
	setGroupNameUseCase         *groupUseCases.SetGroupNameUseCase
	setGroupTopicUseCase        *groupUseCases.SetGroupTopicUseCase
	setGroupPhotoUseCase        *groupUseCases.SetGroupPhotoUseCase
	removeGroupPhotoUseCase     *groupUseCases.RemoveGroupPhotoUseCase
	setGroupAnnounceUseCase     *groupUseCases.SetGroupAnnounceUseCase
	setGroupLockedUseCase       *groupUseCases.SetGroupLockedUseCase
	setDisappearingTimerUseCase *groupUseCases.SetDisappearingTimerUseCase
	getInviteLinkUseCase        *groupUseCases.GetInviteLinkUseCase
	joinGroupUseCase            *groupUseCases.JoinGroupUseCase
	getInviteInfoUseCase        *groupUseCases.GetInviteInfoUseCase
	logger                      logger.Logger
}

// NewGroupHandler cria uma nova instância do group handler
func NewGroupHandler(
	createGroupUC *groupUseCases.CreateGroupUseCase,
	listGroupsUC *groupUseCases.ListGroupsUseCase,
	getGroupInfoUC *groupUseCases.GetGroupInfoUseCase,
	updateParticipantsUC *groupUseCases.UpdateParticipantsUseCase,
	leaveGroupUC *groupUseCases.LeaveGroupUseCase,
	setGroupNameUseCase *groupUseCases.SetGroupNameUseCase,
	setGroupTopicUseCase *groupUseCases.SetGroupTopicUseCase,
	setGroupPhotoUseCase *groupUseCases.SetGroupPhotoUseCase,
	removeGroupPhotoUseCase *groupUseCases.RemoveGroupPhotoUseCase,
	setGroupAnnounceUseCase *groupUseCases.SetGroupAnnounceUseCase,
	setGroupLockedUseCase *groupUseCases.SetGroupLockedUseCase,
	setDisappearingTimerUseCase *groupUseCases.SetDisappearingTimerUseCase,
	getInviteLinkUseCase *groupUseCases.GetInviteLinkUseCase,
	joinGroupUseCase *groupUseCases.JoinGroupUseCase,
	getInviteInfoUseCase *groupUseCases.GetInviteInfoUseCase,
	logger logger.Logger,
) *GroupHandler {
	return &GroupHandler{
		createGroupUC:               createGroupUC,
		listGroupsUC:                listGroupsUC,
		getGroupInfoUC:              getGroupInfoUC,
		updateParticipantsUC:        updateParticipantsUC,
		leaveGroupUC:                leaveGroupUC,
		setGroupNameUseCase:         setGroupNameUseCase,
		setGroupTopicUseCase:        setGroupTopicUseCase,
		setGroupPhotoUseCase:        setGroupPhotoUseCase,
		removeGroupPhotoUseCase:     removeGroupPhotoUseCase,
		setGroupAnnounceUseCase:     setGroupAnnounceUseCase,
		setGroupLockedUseCase:       setGroupLockedUseCase,
		setDisappearingTimerUseCase: setDisappearingTimerUseCase,
		getInviteLinkUseCase:        getInviteLinkUseCase,
		joinGroupUseCase:            joinGroupUseCase,
		getInviteInfoUseCase:        getInviteInfoUseCase,
		logger:                      logger.WithComponent("group-handler"),
	}
}

// CreateGroup cria um novo grupo
// POST /groups/{sessionID}/create
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode create group request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	// Adicionar session ID da URL ao request
	req.SessionID = sessionID.String()

	// Validar request básico
	if err := h.validateCreateGroupRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid create group request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar use case
	groupResult, err := h.createGroupUC.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to create group")
		responses.InternalError(w, "Falha ao criar grupo")
		return
	}

	// Formatar resposta
	response := &group.GroupResponse{
		Details: "Grupo criado com sucesso",
		Group:   groupResult,
	}

	responses.Created(w, "Grupo criado com sucesso", response)
}

// ListGroups lista todos os grupos da sessão
// GET /groups/{sessionID}/list
func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	// Executar use case
	groups, err := h.listGroupsUC.Execute(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to list groups")
		responses.InternalError(w, "Falha ao listar grupos")
		return
	}

	// Formatar resposta
	response := &group.GroupListResponse{
		Details: "Grupos listados com sucesso",
		Groups:  groups,
		Count:   len(groups),
	}

	responses.Success(w, "Grupos listados com sucesso", response)
}

// GetGroupInfo obtém informações de um grupo específico
// GET /groups/{sessionID}/info?groupJid={groupJid}
func (h *GroupHandler) GetGroupInfo(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	// Obter groupJid dos query parameters
	groupJIDStr := r.URL.Query().Get("groupJid")
	if groupJIDStr == "" {
		h.logger.Error().Msg("Missing groupJid parameter")
		responses.BadRequest(w, "Parâmetro groupJid é obrigatório", "")
		return
	}

	// Executar use case
	groupInfo, err := h.getGroupInfoUC.Execute(r.Context(), sessionID, groupJIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to get group info")
		responses.InternalError(w, "Falha ao obter informações do grupo")
		return
	}

	// Formatar resposta
	response := &group.GroupResponse{
		Details: "Informações do grupo obtidas com sucesso",
		Group:   groupInfo,
	}

	responses.Success(w, "Informações do grupo obtidas com sucesso", response)
}

// validateCreateGroupRequest valida a requisição de criação de grupo
func (h *GroupHandler) validateCreateGroupRequest(req group.CreateGroupRequest) error {
	if req.Name == "" {
		return group.NewValidationError("name", req.Name, "nome do grupo é obrigatório")
	}

	if len(req.Name) > 25 {
		return group.NewValidationError("name", req.Name, "nome do grupo deve ter no máximo 25 caracteres")
	}

	if len(req.Participants) == 0 {
		return group.NewValidationError("participants", "", "pelo menos um participante é obrigatório")
	}

	if len(req.Participants) > 256 {
		return group.NewValidationError("participants", "", "máximo de 256 participantes permitidos")
	}

	return nil
}

// UpdateParticipants atualiza participantes do grupo (adicionar, remover, promover, rebaixar)
// POST /groups/{sessionID}/participants/update
func (h *GroupHandler) UpdateParticipants(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.UpdateParticipantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode update participants request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	// Adicionar session ID da URL ao request
	req.SessionID = sessionID.String()

	// Validar request básico
	if err := h.validateUpdateParticipantsRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid update participants request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar use case
	err = h.updateParticipantsUC.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to update participants")
		responses.InternalError(w, "Falha ao atualizar participantes")
		return
	}

	// Formatar resposta
	response := map[string]interface{}{
		"details": fmt.Sprintf("Participantes %s com sucesso", getActionDescription(req.Action)),
		"action":  req.Action,
		"count":   len(req.Phones),
	}

	responses.Success(w, fmt.Sprintf("Participantes %s com sucesso", getActionDescription(req.Action)), response)
}

// LeaveGroup sai de um grupo
// POST /groups/{sessionID}/leave
func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.LeaveGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode leave group request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	// Adicionar session ID da URL ao request
	req.SessionID = sessionID.String()

	// Validar request básico
	if err := h.validateLeaveGroupRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid leave group request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar use case
	err = h.leaveGroupUC.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to leave group")
		responses.InternalError(w, "Falha ao sair do grupo")
		return
	}

	// Formatar resposta
	response := map[string]interface{}{
		"details":  "Saiu do grupo com sucesso",
		"groupJid": req.GroupJID,
	}

	responses.Success(w, "Saiu do grupo com sucesso", response)
}

// validateUpdateParticipantsRequest valida a requisição de atualização de participantes
func (h *GroupHandler) validateUpdateParticipantsRequest(req group.UpdateParticipantsRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "JID do grupo é obrigatório")
	}

	if req.Action == "" {
		return group.NewValidationError("action", req.Action, "ação é obrigatória")
	}

	// Validar ação
	validActions := []string{"add", "remove", "promote", "demote"}
	isValidAction := false
	for _, validAction := range validActions {
		if req.Action == validAction {
			isValidAction = true
			break
		}
	}
	if !isValidAction {
		return group.NewValidationError("action", req.Action, "ação deve ser: add, remove, promote ou demote")
	}

	if len(req.Phones) == 0 {
		return group.NewValidationError("phones", "", "pelo menos um telefone é obrigatório")
	}

	if len(req.Phones) > 50 {
		return group.NewValidationError("phones", "", "máximo de 50 participantes por operação")
	}

	return nil
}

// validateLeaveGroupRequest valida a requisição de sair do grupo
func (h *GroupHandler) validateLeaveGroupRequest(req group.LeaveGroupRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("groupJid", req.GroupJID, "JID do grupo é obrigatório")
	}

	return nil
}

// getActionDescription retorna a descrição da ação em português
func getActionDescription(action string) string {
	switch action {
	case "add":
		return "adicionados"
	case "remove":
		return "removidos"
	case "promote":
		return "promovidos"
	case "demote":
		return "rebaixados"
	default:
		return "atualizados"
	}
}

// SetGroupName define o nome do grupo
func (h *GroupHandler) SetGroupName(w http.ResponseWriter, r *http.Request) {
	// Extrair session ID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.SetGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode set group name request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	req.SessionID = sessionID.String()

	// Validar requisição
	if err := h.validateSetGroupNameRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid set group name request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar caso de uso
	if err := h.setGroupNameUseCase.Execute(r.Context(), sessionID, req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to set group name")
		responses.InternalError(w, "Falha ao definir nome do grupo")
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"Details": "Group Name set successfully",
	}

	responses.Success(w, "Nome do grupo definido com sucesso", response)
}

// validateSetGroupNameRequest valida a requisição de definir nome do grupo
func (h *GroupHandler) validateSetGroupNameRequest(req group.SetGroupNameRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("group_jid", req.GroupJID, "JID do grupo é obrigatório")
	}

	if req.Name == "" {
		return group.NewValidationError("name", req.Name, "Nome do grupo é obrigatório")
	}

	if len(req.Name) > 25 {
		return group.NewValidationError("name", req.Name, "Nome do grupo não pode exceder 25 caracteres")
	}

	return nil
}

// SetGroupTopic define o tópico do grupo
func (h *GroupHandler) SetGroupTopic(w http.ResponseWriter, r *http.Request) {
	// Extrair session ID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.SetGroupTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode set group topic request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	req.SessionID = sessionID.String()

	// Validar requisição
	if err := h.validateSetGroupTopicRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid set group topic request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar caso de uso
	if err := h.setGroupTopicUseCase.Execute(r.Context(), sessionID, req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to set group topic")
		responses.InternalError(w, "Falha ao definir tópico do grupo")
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"Details": "Group Topic set successfully",
	}

	responses.Success(w, "Tópico do grupo definido com sucesso", response)
}

// validateSetGroupTopicRequest valida a requisição de definir tópico do grupo
func (h *GroupHandler) validateSetGroupTopicRequest(req group.SetGroupTopicRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("group_jid", req.GroupJID, "JID do grupo é obrigatório")
	}

	if len(req.Topic) > 512 {
		return group.NewValidationError("topic", req.Topic, "Tópico do grupo não pode exceder 512 caracteres")
	}

	return nil
}

// SetGroupPhoto define a foto do grupo (suporta JSON, form-data, base64, URL)
func (h *GroupHandler) SetGroupPhoto(w http.ResponseWriter, r *http.Request) {
	// Extrair session ID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.SetGroupPhotoRequest
	req.SessionID = sessionID.String()

	// Detectar tipo de conteúdo e processar adequadamente
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		// Processar form-data
		if err := h.parseFormDataPhoto(r, &req); err != nil {
			h.logger.WithError(err).Error().Msg("Failed to parse form-data photo request")
			responses.BadRequest(w, "Erro ao processar form-data", err.Error())
			return
		}
	} else {
		// Processar JSON (base64 ou URL)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.WithError(err).Error().Msg("Failed to decode set group photo request")
			responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
			return
		}
	}

	// Validar requisição
	if err := h.validateSetGroupPhotoRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid set group photo request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar caso de uso
	pictureID, err := h.setGroupPhotoUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to set group photo")
		responses.InternalError(w, "Falha ao definir foto do grupo")
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"Details":   "Group Photo set successfully",
		"PictureID": pictureID,
	}

	responses.Success(w, "Foto do grupo definida com sucesso", response)
}

// RemoveGroupPhoto remove a foto do grupo
func (h *GroupHandler) RemoveGroupPhoto(w http.ResponseWriter, r *http.Request) {
	// Extrair session ID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.RemoveGroupPhotoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode remove group photo request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	req.SessionID = sessionID.String()

	// Validar requisição
	if err := h.validateRemoveGroupPhotoRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid remove group photo request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar caso de uso
	if err := h.removeGroupPhotoUseCase.Execute(r.Context(), sessionID, req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to remove group photo")
		responses.InternalError(w, "Falha ao remover foto do grupo")
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"Details": "Group Photo removed successfully",
	}

	responses.Success(w, "Foto do grupo removida com sucesso", response)
}

// SetGroupAnnounce configura o modo anúncio do grupo
func (h *GroupHandler) SetGroupAnnounce(w http.ResponseWriter, r *http.Request) {
	// Extrair session ID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.SetGroupAnnounceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode set group announce request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	req.SessionID = sessionID.String()

	// Validar requisição
	if err := h.validateSetGroupAnnounceRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid set group announce request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar caso de uso
	if err := h.setGroupAnnounceUseCase.Execute(r.Context(), sessionID, req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to set group announce mode")
		responses.InternalError(w, "Falha ao configurar modo anúncio do grupo")
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"Details": "Group Announce mode set successfully",
	}

	responses.Success(w, "Modo anúncio do grupo configurado com sucesso", response)
}

// SetGroupLocked configura o modo bloqueado do grupo
func (h *GroupHandler) SetGroupLocked(w http.ResponseWriter, r *http.Request) {
	// Extrair session ID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.SetGroupLockedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode set group locked request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	req.SessionID = sessionID.String()

	// Validar requisição
	if err := h.validateSetGroupLockedRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid set group locked request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar caso de uso
	if err := h.setGroupLockedUseCase.Execute(r.Context(), sessionID, req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to set group locked mode")
		responses.InternalError(w, "Falha ao configurar modo bloqueado do grupo")
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"Details": "Group Locked mode set successfully",
	}

	responses.Success(w, "Modo bloqueado do grupo configurado com sucesso", response)
}

// SetDisappearingTimer configura o timer de mensagens temporárias do grupo
func (h *GroupHandler) SetDisappearingTimer(w http.ResponseWriter, r *http.Request) {
	// Extrair session ID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID format")
		responses.BadRequest(w, "Formato de session ID inválido", err.Error())
		return
	}

	var req group.SetDisappearingTimerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode set disappearing timer request")
		responses.BadRequest(w, "Dados da requisição inválidos", err.Error())
		return
	}

	req.SessionID = sessionID.String()

	// Validar requisição
	if err := h.validateSetDisappearingTimerRequest(req); err != nil {
		h.logger.WithError(err).Error().Msg("Invalid set disappearing timer request")
		responses.BadRequest(w, "Dados inválidos", err.Error())
		return
	}

	// Executar caso de uso
	if err := h.setDisappearingTimerUseCase.Execute(r.Context(), sessionID, req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to set disappearing timer")
		responses.InternalError(w, "Falha ao configurar timer de mensagens temporárias")
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"Details": "Disappearing timer set successfully",
	}

	responses.Success(w, "Timer de mensagens temporárias configurado com sucesso", response)
}

// validateSetGroupPhotoRequest valida a requisição de definir foto do grupo
func (h *GroupHandler) validateSetGroupPhotoRequest(req group.SetGroupPhotoRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("group_jid", req.GroupJID, "JID do grupo é obrigatório")
	}

	if req.Image == "" && req.ImageURL == "" {
		return group.NewValidationError("image", "", "Imagem (base64) ou URL da imagem é obrigatória")
	}

	return nil
}

// validateRemoveGroupPhotoRequest valida a requisição de remover foto do grupo
func (h *GroupHandler) validateRemoveGroupPhotoRequest(req group.RemoveGroupPhotoRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("group_jid", req.GroupJID, "JID do grupo é obrigatório")
	}

	return nil
}

// validateSetGroupAnnounceRequest valida a requisição de configurar modo anúncio
func (h *GroupHandler) validateSetGroupAnnounceRequest(req group.SetGroupAnnounceRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("group_jid", req.GroupJID, "JID do grupo é obrigatório")
	}

	return nil
}

// validateSetGroupLockedRequest valida a requisição de configurar modo bloqueado
func (h *GroupHandler) validateSetGroupLockedRequest(req group.SetGroupLockedRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("group_jid", req.GroupJID, "JID do grupo é obrigatório")
	}

	return nil
}

// validateSetDisappearingTimerRequest valida a requisição de configurar timer de mensagens temporárias
func (h *GroupHandler) validateSetDisappearingTimerRequest(req group.SetDisappearingTimerRequest) error {
	if req.GroupJID == "" {
		return group.NewValidationError("group_jid", req.GroupJID, "JID do grupo é obrigatório")
	}

	if req.Duration == "" {
		return group.NewValidationError("duration", req.Duration, "Duração é obrigatória")
	}

	return nil
}

// parseFormDataPhoto processa form-data para SetGroupPhoto
func (h *GroupHandler) parseFormDataPhoto(r *http.Request, req *group.SetGroupPhotoRequest) error {
	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Obter group_jid
	req.GroupJID = r.FormValue("group_jid")
	if req.GroupJID == "" {
		return fmt.Errorf("group_jid é obrigatório")
	}

	// Verificar se há arquivo de imagem
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Ler dados do arquivo
		imageData, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read image file: %w", err)
		}

		// Converter para base64 data URL
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg" // default
		}

		req.Image = fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(imageData))
		return nil
	}

	// Verificar se há URL de imagem
	imageURL := r.FormValue("image_url")
	if imageURL != "" {
		req.ImageURL = imageURL
		return nil
	}

	// Verificar se há base64
	imageBase64 := r.FormValue("image")
	if imageBase64 != "" {
		req.Image = imageBase64
		return nil
	}

	return fmt.Errorf("nenhuma imagem fornecida (image, image_url ou arquivo)")
}

// GetGroupInviteLink obtém o link de convite do grupo
func (h *GroupHandler) GetGroupInviteLink(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("Getting group invite link")

	// Extrair sessionID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID")
		responses.BadRequest(w, "ID de sessão inválido", "Formato de UUID inválido")
		return
	}

	// Obter parâmetros da query string
	groupJID := r.URL.Query().Get("groupJID")
	if groupJID == "" {
		h.logger.Error().Msg("Missing groupJID parameter")
		responses.BadRequest(w, "Parâmetro groupJID é obrigatório", "O parâmetro groupJID deve ser fornecido na query string")
		return
	}

	resetParam := r.URL.Query().Get("reset")
	reset := false
	if resetParam != "" {
		reset = resetParam == "true"
	}

	// Criar requisição
	req := group.GetGroupInviteLinkRequest{
		SessionID: sessionIDStr,
		GroupJID:  groupJID,
		Reset:     reset,
	}

	// Executar caso de uso
	result, err := h.getInviteLinkUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to get group invite link")
		responses.InternalError(w, "Falha ao obter link de convite do grupo")
		return
	}

	h.logger.Info().
		Str("sessionId", sessionID.String()).
		Str("groupJid", groupJID).
		Str("inviteLink", result.InviteLink).
		Msg("Group invite link obtained successfully")

	responses.Success(w, "Link de convite obtido com sucesso", result)
}

// JoinGroupWithLink entra em um grupo usando um link de convite
func (h *GroupHandler) JoinGroupWithLink(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("Joining group with invite link")

	// Extrair sessionID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID")
		responses.BadRequest(w, "ID de sessão inválido", "Formato de UUID inválido")
		return
	}

	// Decodificar corpo da requisição
	var req group.JoinGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode request body")
		responses.BadRequest(w, "Corpo da requisição inválido", "JSON malformado")
		return
	}

	// Validar requisição
	if req.Code == "" {
		h.logger.Error().Msg("Missing invite code")
		responses.BadRequest(w, "Código de convite é obrigatório", "O campo 'code' deve ser fornecido")
		return
	}

	// Executar caso de uso
	result, err := h.joinGroupUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to join group with link")
		responses.InternalError(w, "Falha ao entrar no grupo")
		return
	}

	h.logger.Info().
		Str("sessionId", sessionID.String()).
		Str("code", req.Code).
		Str("groupJid", result.JID.String()).
		Str("groupName", result.Name).
		Msg("Joined group successfully")

	responses.Success(w, "Entrou no grupo com sucesso", result)
}

// GetGroupInviteInfo obtém informações de um convite de grupo
func (h *GroupHandler) GetGroupInviteInfo(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("Getting group invite info")

	// Extrair sessionID da URL
	sessionIDStr := chi.URLParam(r, "sessionID")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Invalid session ID")
		responses.BadRequest(w, "ID de sessão inválido", "Formato de UUID inválido")
		return
	}

	// Decodificar corpo da requisição
	var req group.GetGroupInviteInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error().Msg("Failed to decode request body")
		responses.BadRequest(w, "Corpo da requisição inválido", "JSON malformado")
		return
	}

	// Validar requisição
	if req.Code == "" {
		h.logger.Error().Msg("Missing invite code")
		responses.BadRequest(w, "Código de convite é obrigatório", "O campo 'code' deve ser fornecido")
		return
	}

	// Executar caso de uso
	result, err := h.getInviteInfoUseCase.Execute(r.Context(), sessionID, req)
	if err != nil {
		h.logger.WithError(err).Error().Msg("Failed to get group invite info")
		responses.InternalError(w, "Falha ao obter informações do convite")
		return
	}

	h.logger.Info().
		Str("sessionId", sessionID.String()).
		Str("code", req.Code).
		Str("groupJid", result.GroupJID).
		Str("groupName", result.GroupName).
		Msg("Group invite info obtained successfully")

	responses.Success(w, "Informações do convite obtidas com sucesso", result)
}
