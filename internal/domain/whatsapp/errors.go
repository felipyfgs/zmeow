package whatsapp

import "errors"

var (
	// ErrSessionNotFound indica que a sessão não foi encontrada
	ErrSessionNotFound = errors.New("session not found")
	
	// ErrSessionAlreadyExists indica que a sessão já existe
	ErrSessionAlreadyExists = errors.New("session already exists")
	
	// ErrSessionAlreadyConnected indica que a sessão já está conectada
	ErrSessionAlreadyConnected = errors.New("session already connected")
	
	// ErrSessionNotConnected indica que a sessão não está conectada
	ErrSessionNotConnected = errors.New("session not connected")
	
	// ErrSessionAlreadyAuthenticated indica que a sessão já está autenticada
	ErrSessionAlreadyAuthenticated = errors.New("session already authenticated")
	
	// ErrInvalidPhoneNumber indica que o número de telefone é inválido
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	
	// ErrPairingFailed indica que o pareamento falhou
	ErrPairingFailed = errors.New("pairing failed")
	
	// ErrQRCodeExpired indica que o QR code expirou
	ErrQRCodeExpired = errors.New("qr code expired")
	
	// ErrInvalidJID indica que o JID é inválido
	ErrInvalidJID = errors.New("invalid jid")
)