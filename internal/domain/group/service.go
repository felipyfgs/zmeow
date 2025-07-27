package group

import (
	"time"

	"go.mau.fi/whatsmeow/types"
)

// WhatsAppGroupClient define as operações do cliente WhatsApp para grupos
// Esta interface abstrai as operações diretas com o WhatsApp
type WhatsAppGroupClient interface {
	// Operações básicas de grupo
	CreateGroup(name string, participants []types.JID) (*Group, error)
	GetJoinedGroups() ([]Group, error)
	GetGroupInfo(groupJID types.JID) (*Group, error)
	LeaveGroup(groupJID types.JID) error

	// Gerenciamento de participantes
	AddParticipants(groupJID types.JID, participants []types.JID) error
	RemoveParticipants(groupJID types.JID, participants []types.JID) error
	PromoteParticipants(groupJID types.JID, participants []types.JID) error
	DemoteParticipants(groupJID types.JID, participants []types.JID) error

	// Configurações do grupo
	SetGroupName(groupJID types.JID, name string) error
	SetGroupTopic(groupJID types.JID, topic string) error
	SetGroupPhoto(groupJID types.JID, imageData []byte) (string, error)
	RemoveGroupPhoto(groupJID types.JID) error
	SetGroupAnnounce(groupJID types.JID, announce bool) error
	SetGroupLocked(groupJID types.JID, locked bool) error
	SetDisappearingTimer(groupJID types.JID, duration time.Duration) error

	// Sistema de convites
	GetGroupInviteLink(groupJID types.JID, reset bool) (string, error)
	JoinGroupWithLink(code string) (*Group, error)
	GetGroupInfoFromLink(code string) (*InviteInfo, error)
}

// GroupService define operações de negócio para grupos
type GroupService interface {
	// Validações de negócio
	ValidateGroupName(name string) error
	ValidateParticipantList(participants []string) error
	ValidateGroupJID(groupJID string) (types.JID, error)
	ValidateParticipantJID(participantJID string) (types.JID, error)
	ValidateParticipantAction(action string) (ParticipantAction, error)
	ValidateDisappearingTimer(duration string) (time.Duration, error)
	ValidateImageData(imageData string) ([]byte, error)
	ValidateInviteCode(code string) error

	// Operações de permissão
	CanUserManageGroup(group *Group, userJID types.JID) bool
	CanUserManageParticipants(group *Group, userJID types.JID) bool
	CanUserChangeSettings(group *Group, userJID types.JID) bool
	CanUserPromoteParticipant(group *Group, userJID types.JID, targetJID types.JID) bool
	CanUserRemoveParticipant(group *Group, userJID types.JID, targetJID types.JID) bool

	// Operações de transformação
	ConvertPhonesToJIDs(phones []string) ([]types.JID, error)
	ConvertJIDToPhone(jid types.JID) string
	ParseGroupJID(groupJIDStr string) (types.JID, error)
	ParseParticipantJID(participantJIDStr string) (types.JID, error)

	// Operações de formatação
	FormatGroupResponse(group *Group) *GroupResponse
	FormatGroupListResponse(groups []Group) *GroupListResponse
	FormatInviteLinkResponse(link, code string) *InviteLinkResponse
	FormatInviteInfoResponse(info *InviteInfo) *InviteInfoResponse
}

// InviteInfo representa informações de um convite
type InviteInfo struct {
	Code        string    `json:"code"`
	GroupJID    types.JID `json:"groupJid"`
	GroupName   string    `json:"groupName"`
	CreatedBy   types.JID `json:"createdBy"`
	CreatedAt   int64     `json:"createdAt"`
	Description string    `json:"description,omitempty"`
}
