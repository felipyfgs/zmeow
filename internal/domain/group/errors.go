package group

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow/types"
)

// Erros de domínio específicos para grupos
var (
	// ErrGroupNotFound indica que o grupo não foi encontrado
	ErrGroupNotFound = errors.New("group not found")

	// ErrGroupAlreadyExists indica que um grupo com o nome já existe
	ErrGroupAlreadyExists = errors.New("group already exists")

	// ErrInvalidGroupName indica que o nome do grupo é inválido
	ErrInvalidGroupName = errors.New("invalid group name")

	// ErrInvalidGroupJID indica que o JID do grupo é inválido
	ErrInvalidGroupJID = errors.New("invalid group JID")

	// ErrInvalidParticipantJID indica que o JID do participante é inválido
	ErrInvalidParticipantJID = errors.New("invalid participant JID")

	// ErrParticipantNotFound indica que o participante não foi encontrado
	ErrParticipantNotFound = errors.New("participant not found")

	// ErrParticipantAlreadyExists indica que o participante já existe no grupo
	ErrParticipantAlreadyExists = errors.New("participant already exists in group")

	// ErrInsufficientPermissions indica que o usuário não tem permissões suficientes
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// ErrCannotRemoveOwner indica que não é possível remover o dono do grupo
	ErrCannotRemoveOwner = errors.New("cannot remove group owner")

	// ErrCannotDemoteOwner indica que não é possível rebaixar o dono do grupo
	ErrCannotDemoteOwner = errors.New("cannot demote group owner")

	// ErrInvalidParticipantAction indica que a ação do participante é inválida
	ErrInvalidParticipantAction = errors.New("invalid participant action")

	// ErrInvalidImageFormat indica que o formato da imagem é inválido
	ErrInvalidImageFormat = errors.New("invalid image format")

	// ErrImageTooLarge indica que a imagem é muito grande
	ErrImageTooLarge = errors.New("image too large")

	// ErrInvalidBase64 indica que os dados Base64 são inválidos
	ErrInvalidBase64 = errors.New("invalid base64 data")

	// ErrInvalidDisappearingTimer indica que o timer de desaparecimento é inválido
	ErrInvalidDisappearingTimer = errors.New("invalid disappearing timer")

	// ErrInvalidInviteCode indica que o código de convite é inválido
	ErrInvalidInviteCode = errors.New("invalid invite code")

	// ErrInviteExpired indica que o convite expirou
	ErrInviteExpired = errors.New("invite expired")

	// ErrGroupFull indica que o grupo está cheio
	ErrGroupFull = errors.New("group is full")

	// ErrGroupLocked indica que o grupo está bloqueado
	ErrGroupLocked = errors.New("group is locked")

	// ErrUserNotInGroup indica que o usuário não está no grupo
	ErrUserNotInGroup = errors.New("user not in group")

	// ErrUserAlreadyInGroup indica que o usuário já está no grupo
	ErrUserAlreadyInGroup = errors.New("user already in group")

	// ErrCannotLeaveAsOwner indica que o dono não pode sair do grupo
	ErrCannotLeaveAsOwner = errors.New("owner cannot leave group")

	// ErrInvalidPhoneNumber indica que o número de telefone é inválido
	ErrInvalidPhoneNumber = errors.New("invalid phone number")

	// ErrTooManyParticipants indica que há muitos participantes
	ErrTooManyParticipants = errors.New("too many participants")

	// ErrEmptyParticipantList indica que a lista de participantes está vazia
	ErrEmptyParticipantList = errors.New("empty participant list")

	// ErrInvalidTopicLength indica que o tópico é muito longo
	ErrInvalidTopicLength = errors.New("topic too long")

	// ErrInvalidNameLength indica que o nome é muito longo
	ErrInvalidNameLength = errors.New("name too long")
)

// GroupError representa um erro específico de grupo com contexto adicional
type GroupError struct {
	SessionID uuid.UUID
	GroupJID  types.JID
	Op        string
	Err       error
}

func (e *GroupError) Error() string {
	if e.GroupJID.IsEmpty() {
		return fmt.Sprintf("session %s: %s: %v", e.SessionID, e.Op, e.Err)
	}
	return fmt.Sprintf("session %s: group %s: %s: %v", e.SessionID, e.GroupJID, e.Op, e.Err)
}

func (e *GroupError) Unwrap() error {
	return e.Err
}

// NewGroupError cria um novo erro de grupo
func NewGroupError(sessionID uuid.UUID, groupJID types.JID, op string, err error) *GroupError {
	return &GroupError{
		SessionID: sessionID,
		GroupJID:  groupJID,
		Op:        op,
		Err:       err,
	}
}

// ParticipantError representa um erro específico de participante
type ParticipantError struct {
	SessionID     uuid.UUID
	GroupJID      types.JID
	ParticipantJID types.JID
	Action        ParticipantAction
	Err           error
}

func (e *ParticipantError) Error() string {
	return fmt.Sprintf("session %s: group %s: participant %s: action %s: %v",
		e.SessionID, e.GroupJID, e.ParticipantJID, e.Action, e.Err)
}

func (e *ParticipantError) Unwrap() error {
	return e.Err
}

// NewParticipantError cria um novo erro de participante
func NewParticipantError(sessionID uuid.UUID, groupJID, participantJID types.JID, action ParticipantAction, err error) *ParticipantError {
	return &ParticipantError{
		SessionID:      sessionID,
		GroupJID:       groupJID,
		ParticipantJID: participantJID,
		Action:         action,
		Err:            err,
	}
}

// ValidationError representa um erro de validação específico de grupos
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError cria um novo erro de validação
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// PermissionError representa um erro de permissão
type PermissionError struct {
	UserJID   types.JID
	GroupJID  types.JID
	Operation string
	Required  string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: user %s cannot %s in group %s (required: %s)",
		e.UserJID, e.Operation, e.GroupJID, e.Required)
}

// NewPermissionError cria um novo erro de permissão
func NewPermissionError(userJID, groupJID types.JID, operation, required string) *PermissionError {
	return &PermissionError{
		UserJID:   userJID,
		GroupJID:  groupJID,
		Operation: operation,
		Required:  required,
	}
}

// MediaError representa um erro relacionado a mídia (fotos, etc.)
type MediaError struct {
	MediaType string
	Size      int64
	Format    string
	Err       error
}

func (e *MediaError) Error() string {
	return fmt.Sprintf("media error: type=%s, size=%d, format=%s: %v",
		e.MediaType, e.Size, e.Format, e.Err)
}

func (e *MediaError) Unwrap() error {
	return e.Err
}

// NewMediaError cria um novo erro de mídia
func NewMediaError(mediaType string, size int64, format string, err error) *MediaError {
	return &MediaError{
		MediaType: mediaType,
		Size:      size,
		Format:    format,
		Err:       err,
	}
}

// InviteError representa um erro relacionado a convites
type InviteError struct {
	Code      string
	GroupJID  types.JID
	Operation string
	Err       error
}

func (e *InviteError) Error() string {
	return fmt.Sprintf("invite error: code=%s, group=%s, operation=%s: %v",
		e.Code, e.GroupJID, e.Operation, e.Err)
}

func (e *InviteError) Unwrap() error {
	return e.Err
}

// NewInviteError cria um novo erro de convite
func NewInviteError(code string, groupJID types.JID, operation string, err error) *InviteError {
	return &InviteError{
		Code:      code,
		GroupJID:  groupJID,
		Operation: operation,
		Err:       err,
	}
}
