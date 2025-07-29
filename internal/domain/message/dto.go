package message

import (
	"time"

	"github.com/google/uuid"
)

// SendTextMessageRequest representa a requisição para envio de mensagem de texto
type SendTextMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" swaggertype:"string" format:"phone" description:"Número do destinatário (formato: código do país + DDD + número)"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Text        string                 `json:"text" validate:"required" example:"Olá, isso é um teste!" description:"Texto da mensagem a ser enviada"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem (reply, menções)"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados para a mensagem"`
}

// SendMediaMessageRequest representa a requisição para envio de mídia
type SendMediaMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	MediaType   string                 `json:"mediaType" validate:"required,oneof=image audio video document" example:"image" enum:"image,audio,video,document" description:"Tipo de mídia a ser enviada"`
	Media       string                 `json:"media" validate:"required" example:"data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD..." description:"URL da mídia ou dados Base64 (formato: data:tipo/mime;base64,dados)"`
	Caption     string                 `json:"caption,omitempty" example:"Legenda da imagem" description:"Legenda opcional para a mídia"`
	FileName    string                 `json:"fileName,omitempty" example:"documento.pdf" description:"Nome do arquivo (obrigatório para documentos)"`
	MimeType    string                 `json:"mimeType,omitempty" example:"image/jpeg" description:"Tipo MIME da mídia (detectado automaticamente se não fornecido)"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendImageMessageRequest representa a requisição para envio de imagem
type SendImageMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Image       string                 `json:"image" validate:"required" example:"data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD..." description:"Imagem em Base64 data URL ou URL pública"`
	Caption     string                 `json:"caption,omitempty" example:"Olha essa imagem!" description:"Legenda opcional da imagem"`
	MimeType    string                 `json:"mimeType,omitempty" example:"image/jpeg" description:"Tipo MIME da imagem (image/jpeg, image/png, etc.)"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendAudioMessageRequest representa a requisição para envio de áudio
type SendAudioMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Audio       string                 `json:"audio" validate:"required" example:"data:audio/mpeg;base64,SUQzAwAAAAAfdlBSSVYAAAAgAAAAUGVhY2UuLi4..." description:"Áudio em Base64 data URL ou URL pública"`
	Caption     string                 `json:"caption,omitempty" example:"Mensagem de áudio" description:"Legenda opcional do áudio"`
	PTT         bool                   `json:"ptt,omitempty" example:"true" description:"Push to talk - true para mensagem de voz, false para áudio normal"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendVideoMessageRequest representa a requisição para envio de vídeo
type SendVideoMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Video       string                 `json:"video" validate:"required" example:"data:video/mp4;base64,AAAAIGZ0eXBpc29tAAACAGlzb21pc28yYXZjMW1wNDE..." description:"Vídeo em Base64 data URL ou URL pública"`
	Caption     string                 `json:"caption,omitempty" example:"Vídeo interessante!" description:"Legenda opcional do vídeo"`
	MimeType    string                 `json:"mimeType,omitempty" example:"video/mp4" description:"Tipo MIME do vídeo (video/mp4, video/avi, etc.)"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendDocumentMessageRequest representa a requisição para envio de documento
type SendDocumentMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Document    string                 `json:"document" validate:"required" example:"data:application/pdf;base64,JVBERi0xLjQKJdP0zOEKMSAwIG9iag..." description:"Documento em Base64 data URL ou URL pública"`
	FileName    string                 `json:"fileName" validate:"required" example:"relatorio.pdf" description:"Nome do arquivo com extensão (obrigatório)"`
	Caption     string                 `json:"caption,omitempty" example:"Relatório mensal" description:"Legenda opcional do documento"`
	MimeType    string                 `json:"mimeType,omitempty" example:"application/pdf" description:"Tipo MIME do documento"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendStickerMessageRequest representa a requisição para envio de sticker
type SendStickerMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Sticker     string                 `json:"sticker" validate:"required" example:"data:image/webp;base64,UklGRh4AAABXRUJQVlA4TBIAAAAvAAAAAAfQ..." description:"Sticker em Base64 data URL (preferencialmente WebP)"`
	MimeType    string                 `json:"mimeType,omitempty" example:"image/webp" description:"Tipo MIME do sticker (image/webp recomendado)"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendLocationMessageRequest representa a requisição para envio de localização
type SendLocationMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Latitude    float64                `json:"latitude" validate:"required" example:"-23.550520" description:"Latitude da localização (coordenadas decimais)"`
	Longitude   float64                `json:"longitude" validate:"required" example:"-46.633309" description:"Longitude da localização (coordenadas decimais)"`
	Name        string                 `json:"name,omitempty" example:"Avenida Paulista" description:"Nome opcional do local"`
	Address     string                 `json:"address,omitempty" example:"Av. Paulista, 1578 - Bela Vista, São Paulo - SP" description:"Endereço opcional do local"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendContactMessageRequest representa a requisição para envio de contato
type SendContactMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	ContactName string                 `json:"contactName" validate:"required" example:"João Silva" description:"Nome do contato a ser compartilhado"`
	ContactJID  string                 `json:"contactJID" validate:"required" example:"559987654321@s.whatsapp.net" description:"JID do contato no WhatsApp"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendButtonsMessageRequest representa a requisição para envio de mensagem com botões
type SendButtonsMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Text        string                 `json:"text" validate:"required" example:"Escolha uma opção:" description:"Texto principal da mensagem"`
	Footer      string                 `json:"footer,omitempty" example:"Powered by ZMeow" description:"Texto opcional no rodapé da mensagem"`
	Buttons     []MessageButton        `json:"buttons" validate:"required,min=1,max=3" description:"Lista de botões (mínimo 1, máximo 3)"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendListMessageRequest representa a requisição para envio de mensagem com lista
type SendListMessageRequest struct {
	Number      string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid    string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Text        string                 `json:"text" validate:"required" example:"Escolha uma das opções abaixo:" description:"Texto principal da mensagem"`
	Footer      string                 `json:"footer,omitempty" example:"Powered by ZMeow" description:"Texto opcional no rodapé"`
	Title       string                 `json:"title" validate:"required" example:"Opções disponíveis" description:"Título da lista"`
	ButtonText  string                 `json:"buttonText" validate:"required" example:"Ver opções" description:"Texto do botão que abre a lista"`
	Sections    []MessageListSection   `json:"sections" validate:"required,min=1" description:"Seções da lista com itens"`
	ContextInfo *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// SendPollMessageRequest representa a requisição para envio de enquete
type SendPollMessageRequest struct {
	Number                 string                 `json:"number,omitempty" example:"559981769536" description:"Número do destinatário"`
	GroupJid               string                 `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo de destino"`
	Name                   string                 `json:"name" validate:"required" example:"Qual sua cor favorita?" description:"Pergunta da enquete"`
	Options                []string               `json:"options" validate:"required,min=2,max=12" example:"[\"Azul\", \"Verde\", \"Vermelho\"]" description:"Opções da enquete (mínimo 2, máximo 12)"`
	SelectableOptionsCount int                    `json:"selectableOptionsCount" validate:"min=1" example:"1" description:"Número de opções que podem ser selecionadas"`
	ContextInfo            *MessageContextInfo    `json:"contextInfo,omitempty" description:"Informações de contexto da mensagem"`
	Metadata               map[string]interface{} `json:"metadata,omitempty" description:"Metadados customizados"`
}

// EditMessageRequest representa a requisição para edição de mensagem
type EditMessageRequest struct {
	Number   string `json:"number,omitempty" example:"559981769536" description:"Número do chat onde a mensagem será editada"`
	GroupJid string `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo onde a mensagem será editada"`
	ID       string `json:"id" validate:"required" example:"ABCD123456" description:"ID da mensagem a ser editada"`
	NewText  string `json:"newText" validate:"required" example:"Texto corrigido" description:"Novo texto da mensagem"`
}

// MessageContextInfo representa informações de contexto da mensagem (reply, mention)
type MessageContextInfo struct {
	StanzaID      *string  `json:"stanzaId,omitempty" example:"ABCD123456" description:"ID da mensagem sendo respondida (para reply)"`
	Participant   *string  `json:"participant,omitempty" example:"558199999999@s.whatsapp.net" description:"JID do participante que enviou a mensagem original (necessário para reply em grupos)"`
	MentionedJIDs []string `json:"mentionedJids,omitempty" example:"[\"558199999999@s.whatsapp.net\"]" description:"Lista de JIDs mencionados na mensagem (@mencionar)"`
}

// MessageButton representa um botão em mensagem interativa
type MessageButton struct {
	ID          string `json:"id" validate:"required" example:"btn1" description:"ID único do botão (usado na resposta)"`
	DisplayText string `json:"displayText" validate:"required" example:"Sim" description:"Texto exibido no botão"`
	Type        string `json:"type,omitempty" example:"RESPONSE" description:"Tipo do botão (RESPONSE por padrão)"`
}

// MessageListSection representa uma seção em mensagem de lista
type MessageListSection struct {
	Title string           `json:"title" validate:"required" example:"Categoria A" description:"Título da seção"`
	Rows  []MessageListRow `json:"rows" validate:"required,min=1" description:"Itens da seção"`
}

// MessageListRow representa uma linha em seção de lista
type MessageListRow struct {
	ID          string `json:"id" validate:"required" example:"row1" description:"ID único do item (usado na resposta)"`
	Title       string `json:"title" validate:"required" example:"Opção 1" description:"Título do item"`
	Description string `json:"description,omitempty" example:"Descrição da opção 1" description:"Descrição opcional do item"`
}

// SendMessageResponse representa a resposta de envio de mensagem
type SendMessageResponse struct {
	ID        string                 `json:"id" example:"ABCD123456" description:"ID da mensagem enviada"`
	Timestamp time.Time              `json:"timestamp" example:"2023-12-01T15:30:00Z" description:"Timestamp do envio"`
	Status    string                 `json:"status" example:"sent" description:"Status da mensagem (sent, delivered, read)"`
	Details   map[string]interface{} `json:"details,omitempty" description:"Detalhes adicionais do envio"`
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
	Number   string `json:"number,omitempty" example:"559981769536" description:"Número do chat onde a mensagem será deletada"`
	GroupJid string `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo onde a mensagem será deletada"`
	ID       string `json:"id" validate:"required" example:"ABCD123456" description:"ID da mensagem a ser deletada"`
	ForMe    bool   `json:"forMe,omitempty" example:"false" description:"true = deletar só para mim, false = deletar para todos (padrão: false)"`
}

// ReactMessageRequest representa a requisição para reagir a uma mensagem
type ReactMessageRequest struct {
	Number   string `json:"number,omitempty" example:"559981769536" description:"Número do chat onde a mensagem será reagida"`
	GroupJid string `json:"groupJid,omitempty" example:"120363123456789012@g.us" description:"JID do grupo onde a mensagem será reagida"`
	ID       string `json:"id" validate:"required" example:"ABCD123456" description:"ID da mensagem a ser reagida (use prefixo \"me:\" para suas próprias mensagens)"`
	Reaction string `json:"reaction" example:"👍" description:"Emoji da reação (ex: 👍, ❤️) ou string vazia para remover reação"`
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
