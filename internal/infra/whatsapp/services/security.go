package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"zmeow/internal/domain/whatsapp"
	"zmeow/pkg/logger"
)

// SecurityServiceImpl implementa o serviço de segurança
type SecurityServiceImpl struct {
	logger        logger.Logger
	encryptionKey []byte
}

// NewSecurityService cria uma nova instância do SecurityService
func NewSecurityService(log logger.Logger) whatsapp.SecurityService {
	// Obter chave de criptografia do ambiente ou gerar uma padrão
	encryptionKey := getEncryptionKey()

	return &SecurityServiceImpl{
		logger:        log.WithComponent("security-service"),
		encryptionKey: encryptionKey,
	}
}

// getEncryptionKey obtém a chave de criptografia do ambiente
func getEncryptionKey() []byte {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		// Chave padrão para desenvolvimento (NÃO usar em produção)
		keyStr = "zmeow-default-encryption-key-32b"
	}

	// Garantir que a chave tenha 32 bytes (AES-256)
	key := []byte(keyStr)
	if len(key) < 32 {
		// Pad com zeros se for menor
		padded := make([]byte, 32)
		copy(padded, key)
		return padded
	} else if len(key) > 32 {
		// Truncar se for maior
		return key[:32]
	}

	return key
}

// GenerateSignature gera uma assinatura HMAC-SHA256 para webhook
func (ss *SecurityServiceImpl) GenerateSignature(data []byte, secret string) string {
	if secret == "" {
		ss.logger.Warn().Msg("Empty secret provided for signature generation")
		return ""
	}

	// Criar HMAC com SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	signature := h.Sum(nil)

	// Retornar como hex com prefixo sha256=
	return "sha256=" + hex.EncodeToString(signature)
}

// ValidateSignature valida uma assinatura de webhook
func (ss *SecurityServiceImpl) ValidateSignature(data []byte, signature, secret string) bool {
	if secret == "" || signature == "" {
		ss.logger.Warn().Msg("Empty secret or signature provided for validation")
		return false
	}

	// Gerar assinatura esperada
	expectedSignature := ss.GenerateSignature(data, secret)

	// Comparar usando hmac.Equal para evitar timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// EncryptData criptografa dados sensíveis usando AES-256-GCM
func (ss *SecurityServiceImpl) EncryptData(data string) (string, error) {
	if data == "" {
		return "", fmt.Errorf("data cannot be empty")
	}

	// Criar cipher AES
	block, err := aes.NewCipher(ss.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Criar GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Gerar nonce aleatório
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Criptografar dados
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Codificar em base64
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	ss.logger.Debug().Msg("Data encrypted successfully")
	return encoded, nil
}

// DecryptData descriptografa dados sensíveis
func (ss *SecurityServiceImpl) DecryptData(encryptedData string) (string, error) {
	if encryptedData == "" {
		return "", fmt.Errorf("encrypted data cannot be empty")
	}

	// Decodificar base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Criar cipher AES
	block, err := aes.NewCipher(ss.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Criar GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Verificar tamanho mínimo
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extrair nonce e dados criptografados
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Descriptografar
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	ss.logger.Debug().Msg("Data decrypted successfully")
	return string(plaintext), nil
}

// GenerateRandomKey gera uma chave aleatória para uso em tokens ou secrets
func (ss *SecurityServiceImpl) GenerateRandomKey(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("key length must be positive")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// HashPassword cria um hash seguro de uma senha (para uso futuro)
func (ss *SecurityServiceImpl) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Usar SHA256 com salt (implementação básica)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(password))
	h.Write(salt)
	hash := h.Sum(nil)

	// Combinar salt + hash e codificar em base64
	combined := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword verifica uma senha contra seu hash
func (ss *SecurityServiceImpl) VerifyPassword(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, fmt.Errorf("password and hash cannot be empty")
	}

	// Decodificar hash
	combined, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	if len(combined) < 16+32 { // 16 bytes salt + 32 bytes SHA256
		return false, fmt.Errorf("invalid hash format")
	}

	// Extrair salt e hash
	salt := combined[:16]
	expectedHash := combined[16:]

	// Calcular hash da senha fornecida
	h := sha256.New()
	h.Write([]byte(password))
	h.Write(salt)
	actualHash := h.Sum(nil)

	// Comparar hashes
	return hmac.Equal(expectedHash, actualHash), nil
}
