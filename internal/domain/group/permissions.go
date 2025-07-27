package group

import (
	"fmt"

	"go.mau.fi/whatsmeow/types"
)

// PermissionLevel representa os níveis de permissão em um grupo
type PermissionLevel int

const (
	// PermissionNone - usuário não tem permissões especiais
	PermissionNone PermissionLevel = iota
	// PermissionMember - usuário é membro do grupo
	PermissionMember
	// PermissionAdmin - usuário é administrador do grupo
	PermissionAdmin
	// PermissionOwner - usuário é o dono do grupo
	PermissionOwner
)

// String retorna a representação string do nível de permissão
func (p PermissionLevel) String() string {
	switch p {
	case PermissionNone:
		return "none"
	case PermissionMember:
		return "member"
	case PermissionAdmin:
		return "admin"
	case PermissionOwner:
		return "owner"
	default:
		return "unknown"
	}
}

// PermissionValidator implementa validações de permissões para grupos
type PermissionValidator struct{}

// NewPermissionValidator cria uma nova instância do validador de permissões
func NewPermissionValidator() *PermissionValidator {
	return &PermissionValidator{}
}

// GetUserPermissionLevel retorna o nível de permissão de um usuário no grupo
func (pv *PermissionValidator) GetUserPermissionLevel(group *Group, userJID types.JID) PermissionLevel {
	// Verificar se é o dono
	if group.IsUserOwner(userJID) {
		return PermissionOwner
	}

	// Verificar se é admin
	if group.IsUserAdmin(userJID) {
		return PermissionAdmin
	}

	// Verificar se é membro
	if group.IsUserParticipant(userJID) {
		return PermissionMember
	}

	// Não é membro do grupo
	return PermissionNone
}

// CanAddParticipants verifica se o usuário pode adicionar participantes
func (pv *PermissionValidator) CanAddParticipants(group *Group, userJID types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	if level < PermissionAdmin {
		return NewPermissionError(
			userJID,
			group.JID,
			"add participants",
			"admin or owner",
		)
	}

	// Verificar se o grupo está bloqueado para adições
	if group.IsLocked && level < PermissionOwner {
		return NewPermissionError(
			userJID,
			group.JID,
			"add participants to locked group",
			"owner",
		)
	}

	return nil
}

// CanRemoveParticipants verifica se o usuário pode remover participantes
func (pv *PermissionValidator) CanRemoveParticipants(group *Group, userJID types.JID, targetJIDs []types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	if level < PermissionAdmin {
		return NewPermissionError(
			userJID,
			group.JID,
			"remove participants",
			"admin or owner",
		)
	}

	// Verificar cada participante a ser removido
	for _, targetJID := range targetJIDs {
		// Não pode remover o dono
		if group.IsUserOwner(targetJID) {
			return NewPermissionError(
				userJID,
				group.JID,
				"remove group owner",
				"impossible",
			)
		}

		// Admin só pode remover membros comuns, não outros admins
		if level == PermissionAdmin && group.IsUserAdmin(targetJID) {
			return NewPermissionError(
				userJID,
				group.JID,
				"remove admin participant",
				"owner",
			)
		}
	}

	return nil
}

// CanPromoteParticipants verifica se o usuário pode promover participantes
func (pv *PermissionValidator) CanPromoteParticipants(group *Group, userJID types.JID, targetJIDs []types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	// Apenas o dono pode promover participantes
	if level < PermissionOwner {
		return NewPermissionError(
			userJID,
			group.JID,
			"promote participants",
			"owner",
		)
	}

	// Verificar cada participante a ser promovido
	for _, targetJID := range targetJIDs {
		// Verificar se o usuário está no grupo
		if !group.IsUserParticipant(targetJID) {
			return fmt.Errorf("user %s is not a member of the group", targetJID.String())
		}

		// Verificar se já é admin
		if group.IsUserAdmin(targetJID) {
			return fmt.Errorf("user %s is already an admin", targetJID.String())
		}

		// Não pode promover o próprio dono
		if group.IsUserOwner(targetJID) {
			return fmt.Errorf("cannot promote the group owner")
		}
	}

	return nil
}

// CanDemoteParticipants verifica se o usuário pode rebaixar participantes
func (pv *PermissionValidator) CanDemoteParticipants(group *Group, userJID types.JID, targetJIDs []types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	// Apenas o dono pode rebaixar participantes
	if level < PermissionOwner {
		return NewPermissionError(
			userJID,
			group.JID,
			"demote participants",
			"owner",
		)
	}

	// Verificar cada participante a ser rebaixado
	for _, targetJID := range targetJIDs {
		// Não pode rebaixar o dono
		if group.IsUserOwner(targetJID) {
			return NewPermissionError(
				userJID,
				group.JID,
				"demote group owner",
				"impossible",
			)
		}

		// Verificar se é admin
		if !group.IsUserAdmin(targetJID) {
			return fmt.Errorf("user %s is not an admin", targetJID.String())
		}
	}

	return nil
}

// CanChangeGroupSettings verifica se o usuário pode alterar configurações do grupo
func (pv *PermissionValidator) CanChangeGroupSettings(group *Group, userJID types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	if level < PermissionAdmin {
		return NewPermissionError(
			userJID,
			group.JID,
			"change group settings",
			"admin or owner",
		)
	}

	return nil
}

// CanChangeGroupName verifica se o usuário pode alterar o nome do grupo
func (pv *PermissionValidator) CanChangeGroupName(group *Group, userJID types.JID) error {
	return pv.CanChangeGroupSettings(group, userJID)
}

// CanChangeGroupTopic verifica se o usuário pode alterar o tópico do grupo
func (pv *PermissionValidator) CanChangeGroupTopic(group *Group, userJID types.JID) error {
	return pv.CanChangeGroupSettings(group, userJID)
}

// CanChangeGroupPhoto verifica se o usuário pode alterar a foto do grupo
func (pv *PermissionValidator) CanChangeGroupPhoto(group *Group, userJID types.JID) error {
	return pv.CanChangeGroupSettings(group, userJID)
}

// CanChangeGroupAnnounceMode verifica se o usuário pode alterar o modo de anúncio
func (pv *PermissionValidator) CanChangeGroupAnnounceMode(group *Group, userJID types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	// Apenas o dono pode alterar o modo de anúncio
	if level < PermissionOwner {
		return NewPermissionError(
			userJID,
			group.JID,
			"change announce mode",
			"owner",
		)
	}

	return nil
}

// CanChangeGroupLockedMode verifica se o usuário pode alterar o modo bloqueado
func (pv *PermissionValidator) CanChangeGroupLockedMode(group *Group, userJID types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	// Apenas o dono pode alterar o modo bloqueado
	if level < PermissionOwner {
		return NewPermissionError(
			userJID,
			group.JID,
			"change locked mode",
			"owner",
		)
	}

	return nil
}

// CanSetDisappearingTimer verifica se o usuário pode configurar timer de desaparecimento
func (pv *PermissionValidator) CanSetDisappearingTimer(group *Group, userJID types.JID) error {
	return pv.CanChangeGroupSettings(group, userJID)
}

// CanGetGroupInviteLink verifica se o usuário pode obter link de convite
func (pv *PermissionValidator) CanGetGroupInviteLink(group *Group, userJID types.JID) error {
	return pv.CanChangeGroupSettings(group, userJID)
}

// CanLeaveGroup verifica se o usuário pode sair do grupo
func (pv *PermissionValidator) CanLeaveGroup(group *Group, userJID types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	// Verificar se é membro do grupo
	if level < PermissionMember {
		return NewPermissionError(
			userJID,
			group.JID,
			"leave group",
			"be a member",
		)
	}

	// O dono não pode sair do grupo
	if level == PermissionOwner {
		return ErrCannotLeaveAsOwner
	}

	return nil
}

// CanSendMessages verifica se o usuário pode enviar mensagens no grupo
func (pv *PermissionValidator) CanSendMessages(group *Group, userJID types.JID) error {
	level := pv.GetUserPermissionLevel(group, userJID)
	
	// Verificar se é membro do grupo
	if level < PermissionMember {
		return NewPermissionError(
			userJID,
			group.JID,
			"send messages",
			"be a member",
		)
	}

	// Se o grupo está em modo anúncio, apenas admins e dono podem enviar
	if group.IsAnnounce && level < PermissionAdmin {
		return NewPermissionError(
			userJID,
			group.JID,
			"send messages in announce mode",
			"admin or owner",
		)
	}

	return nil
}

// ValidateParticipantAction valida se uma ação de participante é permitida
func (pv *PermissionValidator) ValidateParticipantAction(group *Group, userJID types.JID, action ParticipantAction, targetJIDs []types.JID) error {
	switch action {
	case ParticipantActionAdd:
		return pv.CanAddParticipants(group, userJID)
	case ParticipantActionRemove:
		return pv.CanRemoveParticipants(group, userJID, targetJIDs)
	case ParticipantActionPromote:
		return pv.CanPromoteParticipants(group, userJID, targetJIDs)
	case ParticipantActionDemote:
		return pv.CanDemoteParticipants(group, userJID, targetJIDs)
	default:
		return fmt.Errorf("invalid participant action: %s", action)
	}
}
