package events

import (
	"reflect"

	"go.mau.fi/whatsmeow/types/events"
)

// EventTypeMapper mapeia tipos de eventos para strings
type EventTypeMapper struct {
	typeMap map[reflect.Type]string
}

// NewEventTypeMapper cria um novo mapeador de tipos de evento
func NewEventTypeMapper() *EventTypeMapper {
	mapper := &EventTypeMapper{
		typeMap: make(map[reflect.Type]string),
	}

	mapper.initializeTypeMap()
	return mapper
}

// initializeTypeMap inicializa o mapeamento de tipos
func (etm *EventTypeMapper) initializeTypeMap() {
	// Eventos principais
	etm.typeMap[reflect.TypeOf(&events.Connected{})] = "Connected"
	etm.typeMap[reflect.TypeOf(&events.Disconnected{})] = "Disconnected"
	etm.typeMap[reflect.TypeOf(&events.LoggedOut{})] = "LoggedOut"
	etm.typeMap[reflect.TypeOf(&events.Message{})] = "Message"
	etm.typeMap[reflect.TypeOf(&events.QR{})] = "QR"
	etm.typeMap[reflect.TypeOf(&events.PairSuccess{})] = "PairSuccess"
	etm.typeMap[reflect.TypeOf(&events.PairError{})] = "PairError"

	// Eventos de recibo e presença
	etm.typeMap[reflect.TypeOf(&events.Receipt{})] = "Receipt"
	etm.typeMap[reflect.TypeOf(&events.Presence{})] = "Presence"
	etm.typeMap[reflect.TypeOf(&events.ChatPresence{})] = "ChatPresence"

	// Eventos de sincronização
	etm.typeMap[reflect.TypeOf(&events.HistorySync{})] = "HistorySync"
	etm.typeMap[reflect.TypeOf(&events.AppStateSyncComplete{})] = "AppStateSyncComplete"
	etm.typeMap[reflect.TypeOf(&events.AppState{})] = "AppState"
	etm.typeMap[reflect.TypeOf(&events.OfflineSyncCompleted{})] = "OfflineSyncCompleted"
	etm.typeMap[reflect.TypeOf(&events.OfflineSyncPreview{})] = "OfflineSyncPreview"

	// Eventos de configuração
	etm.typeMap[reflect.TypeOf(&events.PushNameSetting{})] = "PushNameSetting"
	etm.typeMap[reflect.TypeOf(&events.PushName{})] = "PushName"
	etm.typeMap[reflect.TypeOf(&events.PrivacySettings{})] = "PrivacySettings"
	etm.typeMap[reflect.TypeOf(&events.UnarchiveChatsSetting{})] = "UnarchiveChatsSetting"

	// Eventos de chat
	etm.typeMap[reflect.TypeOf(&events.Archive{})] = "Archive"
	etm.typeMap[reflect.TypeOf(&events.ClearChat{})] = "ClearChat"
	etm.typeMap[reflect.TypeOf(&events.DeleteChat{})] = "DeleteChat"
	etm.typeMap[reflect.TypeOf(&events.DeleteForMe{})] = "DeleteForMe"
	etm.typeMap[reflect.TypeOf(&events.MarkChatAsRead{})] = "MarkChatAsRead"
	etm.typeMap[reflect.TypeOf(&events.Mute{})] = "Mute"
	etm.typeMap[reflect.TypeOf(&events.Pin{})] = "Pin"
	etm.typeMap[reflect.TypeOf(&events.Star{})] = "Star"

	// Eventos de contato
	etm.typeMap[reflect.TypeOf(&events.Contact{})] = "Contact"
	etm.typeMap[reflect.TypeOf(&events.Blocklist{})] = "Blocklist"
	etm.typeMap[reflect.TypeOf(&events.Picture{})] = "Picture"
	etm.typeMap[reflect.TypeOf(&events.UserAbout{})] = "UserAbout"
	etm.typeMap[reflect.TypeOf(&events.UserStatusMute{})] = "UserStatusMute"

	// Eventos de negócios
	etm.typeMap[reflect.TypeOf(&events.BusinessName{})] = "BusinessName"

	// Eventos de chamada
	etm.typeMap[reflect.TypeOf(&events.CallAccept{})] = "CallAccept"
	etm.typeMap[reflect.TypeOf(&events.CallOffer{})] = "CallOffer"
	etm.typeMap[reflect.TypeOf(&events.CallOfferNotice{})] = "CallOfferNotice"
	etm.typeMap[reflect.TypeOf(&events.CallPreAccept{})] = "CallPreAccept"
	etm.typeMap[reflect.TypeOf(&events.CallReject{})] = "CallReject"
	etm.typeMap[reflect.TypeOf(&events.CallRelayLatency{})] = "CallRelayLatency"
	etm.typeMap[reflect.TypeOf(&events.CallTerminate{})] = "CallTerminate"
	etm.typeMap[reflect.TypeOf(&events.CallTransport{})] = "CallTransport"
	etm.typeMap[reflect.TypeOf(&events.UnknownCallEvent{})] = "UnknownCallEvent"

	// Eventos de grupo
	etm.typeMap[reflect.TypeOf(&events.GroupInfo{})] = "GroupInfo"
	etm.typeMap[reflect.TypeOf(&events.JoinedGroup{})] = "JoinedGroup"

	// Eventos de newsletter
	etm.typeMap[reflect.TypeOf(&events.NewsletterJoin{})] = "NewsletterJoin"
	etm.typeMap[reflect.TypeOf(&events.NewsletterLeave{})] = "NewsletterLeave"
	etm.typeMap[reflect.TypeOf(&events.NewsletterLiveUpdate{})] = "NewsletterLiveUpdate"
	etm.typeMap[reflect.TypeOf(&events.NewsletterMuteChange{})] = "NewsletterMuteChange"

	// Eventos de label
	etm.typeMap[reflect.TypeOf(&events.LabelAssociationChat{})] = "LabelAssociationChat"
	etm.typeMap[reflect.TypeOf(&events.LabelAssociationMessage{})] = "LabelAssociationMessage"
	etm.typeMap[reflect.TypeOf(&events.LabelEdit{})] = "LabelEdit"

	// Eventos de mídia
	etm.typeMap[reflect.TypeOf(&events.MediaRetry{})] = "MediaRetry"
	etm.typeMap[reflect.TypeOf(&events.MediaRetryError{})] = "MediaRetryError"

	// Eventos de segurança
	etm.typeMap[reflect.TypeOf(&events.IdentityChange{})] = "IdentityChange"
	etm.typeMap[reflect.TypeOf(&events.UndecryptableMessage{})] = "UndecryptableMessage"

	// Eventos de conexão
	etm.typeMap[reflect.TypeOf(&events.KeepAliveRestored{})] = "KeepAliveRestored"
	etm.typeMap[reflect.TypeOf(&events.KeepAliveTimeout{})] = "KeepAliveTimeout"
	etm.typeMap[reflect.TypeOf(&events.StreamError{})] = "StreamError"
	etm.typeMap[reflect.TypeOf(&events.StreamReplaced{})] = "StreamReplaced"
	etm.typeMap[reflect.TypeOf(&events.TemporaryBan{})] = "TemporaryBan"

	// Eventos especiais
	etm.typeMap[reflect.TypeOf(&events.QRScannedWithoutMultidevice{})] = "QRScannedWithoutMultidevice"
}

// GetEventType retorna o tipo do evento como string
func (etm *EventTypeMapper) GetEventType(evt interface{}) string {
	eventType := reflect.TypeOf(evt)
	if eventName, exists := etm.typeMap[eventType]; exists {
		return eventName
	}
	return "Unknown"
}

// IsKnownEvent verifica se o evento é conhecido
func (etm *EventTypeMapper) IsKnownEvent(evt interface{}) bool {
	eventType := reflect.TypeOf(evt)
	_, exists := etm.typeMap[eventType]
	return exists
}

// GetAllEventTypes retorna todos os tipos de evento conhecidos
func (etm *EventTypeMapper) GetAllEventTypes() []string {
	types := make([]string, 0, len(etm.typeMap))
	for _, eventName := range etm.typeMap {
		types = append(types, eventName)
	}
	return types
}

// AddCustomEventType adiciona um tipo de evento customizado
func (etm *EventTypeMapper) AddCustomEventType(eventType reflect.Type, eventName string) {
	etm.typeMap[eventType] = eventName
}

// RemoveEventType remove um tipo de evento
func (etm *EventTypeMapper) RemoveEventType(eventType reflect.Type) {
	delete(etm.typeMap, eventType)
}

// GetEventTypeCount retorna o número de tipos de evento registrados
func (etm *EventTypeMapper) GetEventTypeCount() int {
	return len(etm.typeMap)
}

// EventCategory representa uma categoria de eventos
type EventCategory string

const (
	CategoryConnection    EventCategory = "connection"
	CategoryMessage       EventCategory = "message"
	CategoryCall          EventCategory = "call"
	CategoryChat          EventCategory = "chat"
	CategoryContact       EventCategory = "contact"
	CategoryGroup         EventCategory = "group"
	CategoryNewsletter    EventCategory = "newsletter"
	CategorySecurity      EventCategory = "security"
	CategorySync          EventCategory = "sync"
	CategoryConfiguration EventCategory = "configuration"
	CategoryMedia         EventCategory = "media"
	CategoryLabel         EventCategory = "label"
	CategoryBusiness      EventCategory = "business"
	CategoryUnknown       EventCategory = "unknown"
)

// GetEventCategory retorna a categoria de um evento
func (etm *EventTypeMapper) GetEventCategory(evt interface{}) EventCategory {
	eventType := etm.GetEventType(evt)

	switch eventType {
	case "Connected", "Disconnected", "LoggedOut", "KeepAliveRestored", "KeepAliveTimeout", "StreamError", "StreamReplaced", "TemporaryBan":
		return CategoryConnection
	case "Message", "Receipt", "UndecryptableMessage":
		return CategoryMessage
	case "CallAccept", "CallOffer", "CallOfferNotice", "CallPreAccept", "CallReject", "CallRelayLatency", "CallTerminate", "CallTransport", "UnknownCallEvent":
		return CategoryCall
	case "Archive", "ClearChat", "DeleteChat", "DeleteForMe", "MarkChatAsRead", "Mute", "Pin", "Star":
		return CategoryChat
	case "Contact", "Blocklist", "Picture", "UserAbout", "UserStatusMute":
		return CategoryContact
	case "GroupInfo", "JoinedGroup":
		return CategoryGroup
	case "NewsletterJoin", "NewsletterLeave", "NewsletterLiveUpdate", "NewsletterMuteChange":
		return CategoryNewsletter
	case "IdentityChange", "QRScannedWithoutMultidevice", "QR", "PairSuccess", "PairError":
		return CategorySecurity
	case "HistorySync", "AppStateSyncComplete", "AppState", "OfflineSyncCompleted", "OfflineSyncPreview":
		return CategorySync
	case "PushNameSetting", "PushName", "PrivacySettings", "UnarchiveChatsSetting":
		return CategoryConfiguration
	case "MediaRetry", "MediaRetryError":
		return CategoryMedia
	case "LabelAssociationChat", "LabelAssociationMessage", "LabelEdit":
		return CategoryLabel
	case "BusinessName":
		return CategoryBusiness
	default:
		return CategoryUnknown
	}
}

// Global instance for easy access
var GlobalEventMapper = NewEventTypeMapper()

// GetEventType função global para compatibilidade
func GetEventType(evt interface{}) string {
	return GlobalEventMapper.GetEventType(evt)
}

// GetEventCategory função global para obter categoria
func GetEventCategory(evt interface{}) EventCategory {
	return GlobalEventMapper.GetEventCategory(evt)
}
