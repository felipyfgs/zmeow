package group

import (
	"time"

	"go.mau.fi/whatsmeow/types"
)

// Group representa um grupo WhatsApp
type Group struct {
	JID              types.JID     `json:"jid"`
	Name             string        `json:"name"`
	Topic            string        `json:"topic"`
	Participants     []Participant `json:"participants"`
	Admins           []types.JID   `json:"admins"`
	Owner            types.JID     `json:"owner"`
	CreatedAt        time.Time     `json:"createdAt"`
	IsAnnounce       bool          `json:"isAnnounce"`
	IsLocked         bool          `json:"isLocked"`
	IsEphemeral      bool          `json:"isEphemeral"`
	EphemeralTimer   time.Duration `json:"ephemeralTimer"`
	PictureID        string        `json:"pictureId,omitempty"`
	InviteCode       string        `json:"inviteCode,omitempty"`
	ParticipantCount int           `json:"participantCount"`
}

// Participant representa um participante do grupo
type Participant struct {
	JID          types.JID `json:"jid"`
	IsAdmin      bool      `json:"isAdmin"`
	IsSuperAdmin bool      `json:"isSuperAdmin"`
	JoinedAt     time.Time `json:"joinedAt"`
}

// CreateGroupRequest representa a requisição para criar grupo
type CreateGroupRequest struct {
	SessionID    string   `json:"session_id" validate:"required,uuid"`
	Name         string   `json:"name" validate:"required,min=1,max=25"`
	Participants []string `json:"participants" validate:"required,min=1,dive,phone"`
}

// UpdateParticipantsRequest representa a requisição para atualizar participantes
type UpdateParticipantsRequest struct {
	SessionID string   `json:"session_id" validate:"required,uuid"`
	GroupJID  string   `json:"group_jid" validate:"required"`
	Phones    []string `json:"phones" validate:"required,min=1,dive,phone"`
	Action    string   `json:"action" validate:"required,oneof=add remove promote demote"`
}

// SetGroupPhotoRequest representa a requisição para definir foto
type SetGroupPhotoRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
	Image     string `json:"image,omitempty"`     // Base64 data URL
	ImageURL  string `json:"image_url,omitempty"` // URL da imagem
}

// RemoveGroupPhotoRequest representa a requisição para remover foto
type RemoveGroupPhotoRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
}

// SetGroupNameRequest representa a requisição para definir nome
type SetGroupNameRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
	Name      string `json:"name" validate:"required,min=1,max=25"`
}

// SetGroupTopicRequest representa a requisição para definir tópico
type SetGroupTopicRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
	Topic     string `json:"topic" validate:"max=512"`
}

// SetGroupAnnounceRequest representa a requisição para configurar anúncios
type SetGroupAnnounceRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
	Announce  bool   `json:"announce"`
}

// SetGroupLockedRequest representa a requisição para configurar bloqueio
type SetGroupLockedRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
	Locked    bool   `json:"locked"`
}

// SetDisappearingTimerRequest representa a requisição para timer de desaparecimento
type SetDisappearingTimerRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
	Duration  string `json:"duration" validate:"required,oneof=off 24h 7d 90d"`
}

// JoinGroupRequest representa a requisição para entrar no grupo
type JoinGroupRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	Code      string `json:"code" validate:"required"`
}

// GetGroupInviteInfoRequest representa a requisição para info do convite
type GetGroupInviteInfoRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	Code      string `json:"code" validate:"required"`
}

// GetGroupInviteLinkRequest representa a requisição para obter link de convite
type GetGroupInviteLinkRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
	Reset     bool   `json:"reset"`
}

// LeaveGroupRequest representa a requisição para sair do grupo
type LeaveGroupRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
}

// GetGroupInfoRequest representa a requisição para obter informações do grupo
type GetGroupInfoRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
	GroupJID  string `json:"group_jid" validate:"required"`
}

// ListGroupsRequest representa a requisição para listar grupos
type ListGroupsRequest struct {
	SessionID string `json:"session_id" validate:"required,uuid"`
}

// GroupResponse representa a resposta padrão de grupo
type GroupResponse struct {
	Details string `json:"details"`
	Group   *Group `json:"group,omitempty"`
}

// GroupListResponse representa a resposta de lista de grupos
type GroupListResponse struct {
	Details string  `json:"details"`
	Groups  []Group `json:"groups"`
	Count   int     `json:"count"`
}

// PhotoResponse representa a resposta com ID da foto
type PhotoResponse struct {
	Details   string `json:"details"`
	PictureID string `json:"pictureId"`
}

// InviteLinkResponse representa a resposta com link de convite
type InviteLinkResponse struct {
	Details    string `json:"details"`
	InviteLink string `json:"inviteLink"`
	Code       string `json:"code"`
}

// InviteInfoResponse representa a resposta com informações do convite
type InviteInfoResponse struct {
	Details     string `json:"details"`
	GroupName   string `json:"groupName"`
	GroupJID    string `json:"groupJid"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
	Description string `json:"description,omitempty"`
}

// ParticipantAction representa as ações possíveis com participantes
type ParticipantAction string

const (
	ParticipantActionAdd     ParticipantAction = "add"
	ParticipantActionRemove  ParticipantAction = "remove"
	ParticipantActionPromote ParticipantAction = "promote"
	ParticipantActionDemote  ParticipantAction = "demote"
)

// DisappearingTimerDuration representa as durações válidas para timer
type DisappearingTimerDuration string

const (
	DisappearingTimerOff DisappearingTimerDuration = "off"
	DisappearingTimer24h DisappearingTimerDuration = "24h"
	DisappearingTimer7d  DisappearingTimerDuration = "7d"
	DisappearingTimer90d DisappearingTimerDuration = "90d"
)

// Métodos da entidade Group

// IsUserAdmin verifica se um usuário é admin do grupo
func (g *Group) IsUserAdmin(userJID types.JID) bool {
	for _, admin := range g.Admins {
		if admin.User == userJID.User {
			return true
		}
	}
	return false
}

// IsUserOwner verifica se um usuário é o dono do grupo
func (g *Group) IsUserOwner(userJID types.JID) bool {
	return g.Owner.User == userJID.User
}

// IsUserParticipant verifica se um usuário é participante do grupo
func (g *Group) IsUserParticipant(userJID types.JID) bool {
	for _, participant := range g.Participants {
		if participant.JID.User == userJID.User {
			return true
		}
	}
	return false
}

// GetParticipant retorna um participante específico
func (g *Group) GetParticipant(userJID types.JID) *Participant {
	for i, participant := range g.Participants {
		if participant.JID.User == userJID.User {
			return &g.Participants[i]
		}
	}
	return nil
}

// CanUserManageParticipants verifica se um usuário pode gerenciar participantes
func (g *Group) CanUserManageParticipants(userJID types.JID) bool {
	return g.IsUserOwner(userJID) || g.IsUserAdmin(userJID)
}

// CanUserChangeSettings verifica se um usuário pode alterar configurações
func (g *Group) CanUserChangeSettings(userJID types.JID) bool {
	return g.IsUserOwner(userJID) || g.IsUserAdmin(userJID)
}

// CanUserSendMessages verifica se um usuário pode enviar mensagens
func (g *Group) CanUserSendMessages(userJID types.JID) bool {
	if !g.IsAnnounce {
		return g.IsUserParticipant(userJID)
	}
	return g.IsUserOwner(userJID) || g.IsUserAdmin(userJID)
}

// UpdateParticipantCount atualiza o contador de participantes
func (g *Group) UpdateParticipantCount() {
	g.ParticipantCount = len(g.Participants)
}

// Métodos para DisappearingTimerDuration

// ToDuration converte a duração string para time.Duration
func (d DisappearingTimerDuration) ToDuration() time.Duration {
	switch d {
	case DisappearingTimer24h:
		return 24 * time.Hour
	case DisappearingTimer7d:
		return 7 * 24 * time.Hour
	case DisappearingTimer90d:
		return 90 * 24 * time.Hour
	default:
		return 0
	}
}

// IsValid verifica se a duração é válida
func (d DisappearingTimerDuration) IsValid() bool {
	switch d {
	case DisappearingTimerOff, DisappearingTimer24h, DisappearingTimer7d, DisappearingTimer90d:
		return true
	default:
		return false
	}
}

// ParseDisappearingTimerDuration converte string para DisappearingTimerDuration
func ParseDisappearingTimerDuration(duration string) (DisappearingTimerDuration, bool) {
	d := DisappearingTimerDuration(duration)
	return d, d.IsValid()
}
