package message

import (
	"time"

	"github.com/google/uuid"
)

// SendTextMessageRequest representa a requisição para envio de mensagem de texto
type SendTextMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Message     string                 `json:"message" validate:"required"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendMediaMessageRequest representa a requisição para envio de mídia
type SendMediaMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	MediaType   string                 `json:"mediaType" validate:"required,oneof=image audio video document"`
	Media       string                 `json:"media" validate:"required"` // URL da mídia ou dados Base64
	Caption     string                 `json:"caption,omitempty"`
	FileName    string                 `json:"fileName,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendImageMessageRequest representa a requisição para envio de imagem
type SendImageMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Image       string                 `json:"image" validate:"required"` // Base64 data URL
	Caption     string                 `json:"caption,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendAudioMessageRequest representa a requisição para envio de áudio
type SendAudioMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Audio       string                 `json:"audio" validate:"required"` // Base64 data URL
	Caption     string                 `json:"caption,omitempty"`
	PTT         bool                   `json:"ptt,omitempty"` // Push to talk (voice message)
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendVideoMessageRequest representa a requisição para envio de vídeo
type SendVideoMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Video       string                 `json:"video" validate:"required"` // Base64 data URL
	Caption     string                 `json:"caption,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendDocumentMessageRequest representa a requisição para envio de documento
type SendDocumentMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Document    string                 `json:"document" validate:"required"` // Base64 data URL
	FileName    string                 `json:"fileName" validate:"required"`
	Caption     string                 `json:"caption,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendStickerMessageRequest representa a requisição para envio de sticker
type SendStickerMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Sticker     string                 `json:"sticker" validate:"required"` // Base64 data URL
	MimeType    string                 `json:"mimeType,omitempty"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendLocationMessageRequest representa a requisição para envio de localização
type SendLocationMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Latitude    float64                `json:"latitude" validate:"required"`
	Longitude   float64                `json:"longitude" validate:"required"`
	Name        string                 `json:"name,omitempty"`
	Address     string                 `json:"address,omitempty"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendContactMessageRequest representa a requisição para envio de contato
type SendContactMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	ContactName string                 `json:"contactName" validate:"required"`
	ContactJID  string                 `json:"contactJID" validate:"required"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendButtonsMessageRequest representa a requisição para envio de mensagem com botões
type SendButtonsMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Text        string                 `json:"text" validate:"required"`
	Footer      string                 `json:"footer,omitempty"`
	Buttons     []MessageButton        `json:"buttons" validate:"required,min=1,max=3"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendListMessageRequest representa a requisição para envio de mensagem com lista
type SendListMessageRequest struct {
	Number      string                 `json:"number" validate:"required"`
	Text        string                 `json:"text" validate:"required"`
	Footer      string                 `json:"footer,omitempty"`
	Title       string                 `json:"title" validate:"required"`
	ButtonText  string                 `json:"buttonText" validate:"required"`
	Sections    []MessageListSection   `json:"sections" validate:"required,min=1"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendPollMessageRequest representa a requisição para envio de enquete
type SendPollMessageRequest struct {
	Number                 string                 `json:"number" validate:"required"`
	Name                   string                 `json:"name" validate:"required"`
	Options                []string               `json:"options" validate:"required,min=2,max=12"`
	SelectableOptionsCount int                    `json:"selectableOptionsCount" validate:"min=1"`
	ContextInfo            *MessageContextInfo    `json:"contextInfo,omitempty"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
}

// EditMessageRequest representa a requisição para edição de mensagem
type EditMessageRequest struct {
	Number  string `json:"number" validate:"required"`
	ID      string `json:"id" validate:"required"`
	NewText string `json:"newText" validate:"required"`
}

// MessageContextInfo representa informações de contexto da mensagem (reply, mention)
type MessageContextInfo struct {
	StanzaID      *string  `json:"stanzaId,omitempty"`      // ID da mensagem sendo respondida
	Participant   *string  `json:"participant,omitempty"`   // JID do participante (para grupos)
	MentionedJIDs []string `json:"mentionedJids,omitempty"` // JIDs mencionados na mensagem
}

// MessageButton representa um botão em mensagem interativa
type MessageButton struct {
	ID          string `json:"id" validate:"required"`
	DisplayText string `json:"displayText" validate:"required"`
	Type        string `json:"type,omitempty"` // "RESPONSE" por padrão
}

// MessageListSection representa uma seção em mensagem de lista
type MessageListSection struct {
	Title string           `json:"title" validate:"required"`
	Rows  []MessageListRow `json:"rows" validate:"required,min=1"`
}

// MessageListRow representa uma linha em seção de lista
type MessageListRow struct {
	ID          string `json:"id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description,omitempty"`
}

// SendMessageResponse representa a resposta de envio de mensagem
type SendMessageResponse struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// MessageHistoryRequest representa a requisição para histórico de mensagens
type MessageHistoryRequest struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	ChatID string `json:"chatId,omitempty"`
	Before string `json:"before,omitempty"` // Message ID
	After  string `json:"after,omitempty"`  // Message ID
}

// MessageHistoryResponse representa a resposta do histórico de mensagens
type MessageHistoryResponse struct {
	Messages   []MessageInfo `json:"messages"`
	TotalCount int           `json:"totalCount"`
	HasMore    bool          `json:"hasMore"`
}

// DeleteMessageRequest representa a requisição para deletar mensagem
type DeleteMessageRequest struct {
	Number string `json:"number" validate:"required"`
	ID     string `json:"id" validate:"required"`
	ForMe  bool   `json:"forMe,omitempty"` // true = deletar só para mim, false = deletar para todos
}

// ReactMessageRequest representa a requisição para reagir a uma mensagem
type ReactMessageRequest struct {
	Number   string `json:"number" validate:"required"`
	ID       string `json:"id" validate:"required"` // ID da mensagem. Use "me:" como prefixo para suas próprias mensagens
	Reaction string `json:"reaction"`               // Reação/emoji (ex: "👍", "❤️", "" para remover)
}

// MessageInfo representa informações de uma mensagem
type MessageInfo struct {
	ID        string                 `json:"id"`
	SessionID uuid.UUID              `json:"sessionId"`
	ChatID    string                 `json:"chatId"`
	SenderID  string                 `json:"senderId"`
	Type      string                 `json:"type"`
	Content   map[string]interface{} `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	IsFromMe  bool                   `json:"isFromMe"`
	Status    string                 `json:"status,omitempty"`
}
