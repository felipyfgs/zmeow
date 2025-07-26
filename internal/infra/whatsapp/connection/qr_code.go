package connection

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"

	"zmeow/pkg/logger"
)

// QRCodeData representa os dados de um QR code
type QRCodeData struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// QRCodeManager gerencia QR codes para autenticação
type QRCodeManager struct {
	qrCodes map[uuid.UUID]*QRCodeData
	mutex   sync.RWMutex
	logger  logger.Logger
}

// NewQRCodeManager cria uma nova instância do QRCodeManager
func NewQRCodeManager(log logger.Logger) *QRCodeManager {
	return &QRCodeManager{
		qrCodes: make(map[uuid.UUID]*QRCodeData),
		logger:  log.WithComponent("qr-manager"),
	}
}

// GenerateQRCode inicia o processo de geração de QR code para uma sessão
func (qm *QRCodeManager) GenerateQRCode(ctx context.Context, sessionID uuid.UUID, client *whatsmeow.Client) (<-chan whatsmeow.QRChannelItem, error) {
	qm.logger.WithField("sessionId", sessionID).Info().Msg("Starting QR code generation")

	// Obter canal de QR code
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		qm.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("Failed to get QR channel")
		return nil, err
	}

	// Processar eventos QR em goroutine
	go qm.processQREvents(sessionID, qrChan)

	return qrChan, nil
}

// processQREvents processa eventos do canal QR
func (qm *QRCodeManager) processQREvents(sessionID uuid.UUID, qrChan <-chan whatsmeow.QRChannelItem) {
	qm.logger.WithField("sessionId", sessionID).Debug().Msg("Starting QR event processing")

	for evt := range qrChan {
		qm.logger.WithFields(map[string]interface{}{
			"sessionId": sessionID,
			"event":     evt.Event,
		}).Debug().Msg("QR event received")

		switch evt.Event {
		case "code":
			qm.handleQRCode(sessionID, evt.Code)
		case "success":
			qm.handleQRSuccess(sessionID)
		case "timeout":
			qm.handleQRTimeout(sessionID)
		case "error":
			qm.handleQRError(sessionID, evt.Error)
		default:
			qm.logger.WithFields(map[string]interface{}{
				"sessionId": sessionID,
				"event":     evt.Event,
			}).Debug().Msg("Unknown QR event")
		}
	}

	qm.logger.WithField("sessionId", sessionID).Debug().Msg("QR event processing finished")
}

// handleQRCode processa um novo QR code
func (qm *QRCodeManager) handleQRCode(sessionID uuid.UUID, code string) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// Criar dados do QR code
	qrData := &QRCodeData{
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Second), // QR codes expiram em 30 segundos
	}

	qm.qrCodes[sessionID] = qrData

	qm.logger.WithField("sessionId", sessionID).Info().Msg("QR code generated")

	// Exibir QR code no terminal
	qm.displayQRCodeInTerminal(code)
}

// handleQRSuccess processa sucesso na autenticação via QR
func (qm *QRCodeManager) handleQRSuccess(sessionID uuid.UUID) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// Remover QR code após sucesso
	delete(qm.qrCodes, sessionID)

	qm.logger.WithField("sessionId", sessionID).Info().Msg("QR code authentication successful")
}

// handleQRTimeout processa timeout do QR code
func (qm *QRCodeManager) handleQRTimeout(sessionID uuid.UUID) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// Remover QR code expirado
	delete(qm.qrCodes, sessionID)

	qm.logger.WithField("sessionId", sessionID).Warn().Msg("QR code expired")
}

// handleQRError processa erro no QR code
func (qm *QRCodeManager) handleQRError(sessionID uuid.UUID, err error) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// Remover QR code com erro
	delete(qm.qrCodes, sessionID)

	qm.logger.WithError(err).WithField("sessionId", sessionID).Error().Msg("QR code error")
}

// GetQRCode retorna o QR code atual de uma sessão
func (qm *QRCodeManager) GetQRCode(sessionID uuid.UUID) (string, error) {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	qrData, exists := qm.qrCodes[sessionID]
	if !exists {
		return "", nil // Não há QR code para esta sessão
	}

	// Verificar se o QR code expirou
	if time.Now().After(qrData.ExpiresAt) {
		// Remover QR code expirado
		delete(qm.qrCodes, sessionID)
		return "", nil
	}

	return qrData.Code, nil
}

// GetQRCodeData retorna os dados completos do QR code
func (qm *QRCodeManager) GetQRCodeData(sessionID uuid.UUID) (*QRCodeData, error) {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	qrData, exists := qm.qrCodes[sessionID]
	if !exists {
		return nil, nil
	}

	// Verificar se o QR code expirou
	if time.Now().After(qrData.ExpiresAt) {
		// Remover QR code expirado
		delete(qm.qrCodes, sessionID)
		return nil, nil
	}

	// Retornar cópia para evitar modificações
	dataCopy := *qrData
	return &dataCopy, nil
}

// ClearQRCode remove o QR code de uma sessão
func (qm *QRCodeManager) ClearQRCode(sessionID uuid.UUID) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	delete(qm.qrCodes, sessionID)
	qm.logger.WithField("sessionId", sessionID).Debug().Msg("QR code cleared")
}

// IsQRCodeValid verifica se há um QR code válido para a sessão
func (qm *QRCodeManager) IsQRCodeValid(sessionID uuid.UUID) bool {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	qrData, exists := qm.qrCodes[sessionID]
	if !exists {
		return false
	}

	return time.Now().Before(qrData.ExpiresAt)
}

// GetAllQRCodes retorna todos os QR codes ativos
func (qm *QRCodeManager) GetAllQRCodes() map[uuid.UUID]*QRCodeData {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	result := make(map[uuid.UUID]*QRCodeData)
	now := time.Now()

	for sessionID, qrData := range qm.qrCodes {
		// Só incluir QR codes válidos
		if now.Before(qrData.ExpiresAt) {
			dataCopy := *qrData
			result[sessionID] = &dataCopy
		}
	}

	return result
}

// CleanupExpiredQRCodes remove QR codes expirados
func (qm *QRCodeManager) CleanupExpiredQRCodes() {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for sessionID, qrData := range qm.qrCodes {
		if now.After(qrData.ExpiresAt) {
			delete(qm.qrCodes, sessionID)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		qm.logger.WithField("count", expiredCount).Debug().Msg("Cleaned up expired QR codes")
	}
}

// StartCleanupRoutine inicia uma rotina de limpeza automática
func (qm *QRCodeManager) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Limpeza a cada 10 segundos
	defer ticker.Stop()

	qm.logger.Info().Msg("QR code cleanup routine started")

	for {
		select {
		case <-ctx.Done():
			qm.logger.Info().Msg("QR code cleanup routine stopped")
			return
		case <-ticker.C:
			qm.CleanupExpiredQRCodes()
		}
	}
}

// displayQRCodeInTerminal exibe o QR code no terminal
func (qm *QRCodeManager) displayQRCodeInTerminal(code string) {
	qm.logger.Info().Msg("=== QR CODE PARA WHATSAPP ===")
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🔗 QR CODE PARA CONECTAR AO WHATSAPP")
	fmt.Println("📱 Escaneie com seu WhatsApp:")
	fmt.Println("   1. Abra o WhatsApp no seu celular")
	fmt.Println("   2. Vá em Configurações > Aparelhos conectados")
	fmt.Println("   3. Toque em 'Conectar um aparelho'")
	fmt.Println("   4. Escaneie o QR Code abaixo")
	fmt.Println(strings.Repeat("=", 50))
	
	qrterminal.GenerateHalfBlock(code, qrterminal.L, os.Stdout)
	
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("⏰ Este QR Code expira em 30 segundos")
	fmt.Println("🔄 Se expirar, faça uma nova requisição para /qr")
	fmt.Println(strings.Repeat("=", 50) + "\n")
}

// Close encerra o QRCodeManager
func (qm *QRCodeManager) Close() {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	count := len(qm.qrCodes)
	qm.qrCodes = make(map[uuid.UUID]*QRCodeData)

	qm.logger.WithField("clearedCount", count).Info().Msg("QRCodeManager closed")
}
