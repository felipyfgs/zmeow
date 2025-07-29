package message

import (
	"fmt"
	"regexp"
	"strings"
)

// NumberValidator contém métodos utilitários para validação e normalização de números e JIDs
type NumberValidator struct{}

// NewNumberValidator cria uma nova instância do validador
func NewNumberValidator() *NumberValidator {
	return &NumberValidator{}
}

// ValidateDestination valida se pelo menos um dos campos (number ou groupJid) está preenchido
func (nv *NumberValidator) ValidateDestination(number, groupJid string) error {
	if number == "" && groupJid == "" {
		return fmt.Errorf("either number or groupJid must be provided")
	}
	
	if number != "" && groupJid != "" {
		return fmt.Errorf("only one of number or groupJid should be provided")
	}
	
	if number != "" {
		if !nv.IsValidNumber(number) {
			return fmt.Errorf("invalid number format")
		}
	}
	
	if groupJid != "" {
		if !nv.IsValidGroupJid(groupJid) {
			return fmt.Errorf("invalid groupJid format")
		}
	}
	
	return nil
}

// GetDestination retorna o destinatário normalizado (number ou groupJid)
func (nv *NumberValidator) GetDestination(number, groupJid string) string {
	if number != "" {
		return nv.NormalizeNumber(number)
	}
	if groupJid != "" {
		return groupJid // GroupJid já deve vir no formato correto
	}
	return ""
}

// IsValidNumber valida o formato do número de telefone
func (nv *NumberValidator) IsValidNumber(number string) bool {
	// Se já é um JID individual, validar
	if strings.HasSuffix(number, "@s.whatsapp.net") {
		numericPart := strings.TrimSuffix(number, "@s.whatsapp.net")
		if matched, _ := regexp.MatchString(`^\d+$`, numericPart); matched && len(numericPart) >= 10 && len(numericPart) <= 15 {
			return true
		}
		return false
	}

	// Remover caracteres não numéricos
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(number, "")

	// Verificar se tem pelo menos 10 dígitos e no máximo 15
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}

	// Verificar se começa com código de país válido
	if len(cleaned) >= 11 && (strings.HasPrefix(cleaned, "55") || strings.HasPrefix(cleaned, "1")) {
		return true
	}

	// Aceitar números com 10-11 dígitos (formato nacional)
	if len(cleaned) >= 10 && len(cleaned) <= 11 {
		return true
	}

	return false
}

// IsValidGroupJid valida o formato do JID de grupo
func (nv *NumberValidator) IsValidGroupJid(groupJid string) bool {
	// Verificar se é um JID de grupo (formato: numero@g.us)
	if strings.HasSuffix(groupJid, "@g.us") {
		// Extrair a parte numérica antes do @g.us
		numericPart := strings.TrimSuffix(groupJid, "@g.us")
		// Verificar se a parte numérica contém apenas dígitos e tem tamanho adequado
		if matched, _ := regexp.MatchString(`^\d+$`, numericPart); matched && len(numericPart) >= 10 && len(numericPart) <= 25 {
			return true
		}
	}
	return false
}

// NormalizeNumber normaliza o número de telefone para o formato WhatsApp
func (nv *NumberValidator) NormalizeNumber(number string) string {
	// Se já é um JID válido individual, retornar como está
	if strings.HasSuffix(number, "@s.whatsapp.net") {
		return number
	}

	// Remover caracteres não numéricos para números simples
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(number, "")

	// Se não tem código de país, assumir Brasil (55)
	if len(cleaned) == 10 || len(cleaned) == 11 {
		if !strings.HasPrefix(cleaned, "55") {
			cleaned = "55" + cleaned
		}
	}

	// Adicionar sufixo @s.whatsapp.net para números individuais
	return cleaned + "@s.whatsapp.net"
}
