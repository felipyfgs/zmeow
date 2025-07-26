package session

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Erros de domínio específicos para sessões
var (
	// ErrSessionNotFound indica que a sessão não foi encontrada
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionAlreadyExists indica que uma sessão com o nome já existe
	ErrSessionAlreadyExists = errors.New("session already exists")

	// ErrSessionAlreadyConnected indica que a sessão já está conectada
	ErrSessionAlreadyConnected = errors.New("session already connected")

	// ErrSessionNotConnected indica que a sessão não está conectada
	ErrSessionNotConnected = errors.New("session not connected")

	// ErrSessionConnecting indica que a sessão está em processo de conexão
	ErrSessionConnecting = errors.New("session is connecting")

	// ErrSessionInactive indica que a sessão está inativa
	ErrSessionInactive = errors.New("session is inactive")

	// ErrInvalidSessionName indica que o nome da sessão é inválido
	ErrInvalidSessionName = errors.New("invalid session name")

	// ErrInvalidPhoneNumber indica que o número de telefone é inválido
	ErrInvalidPhoneNumber = errors.New("invalid phone number")

	// ErrInvalidProxyURL indica que a URL do proxy é inválida
	ErrInvalidProxyURL = errors.New("invalid proxy URL")

	// ErrQRCodeNotAvailable indica que o QR code não está disponível
	ErrQRCodeNotAvailable = errors.New("QR code not available")

	// ErrPairingCodeNotAvailable indica que o código de pareamento não está disponível
	ErrPairingCodeNotAvailable = errors.New("pairing code not available")
)

// SessionError representa um erro específico de sessão com contexto adicional
type SessionError struct {
	SessionID uuid.UUID
	Op        string
	Err       error
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("session %s: %s: %v", e.SessionID, e.Op, e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}

// NewSessionError cria um novo erro de sessão
func NewSessionError(sessionID uuid.UUID, op string, err error) *SessionError {
	return &SessionError{
		SessionID: sessionID,
		Op:        op,
		Err:       err,
	}
}

// ValidationError representa um erro de validação
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