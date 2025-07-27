package app

import (
	"github.com/uptrace/bun"

	"zmeow/internal/domain/group"
	"zmeow/internal/domain/session"
	"zmeow/internal/domain/whatsapp"
	"zmeow/internal/http/handlers"
	"zmeow/internal/infra/database"
	"zmeow/internal/infra/media"
	groupUseCases "zmeow/internal/usecases/group"
	messageUseCases "zmeow/internal/usecases/message"
	sessionUseCases "zmeow/internal/usecases/session"
	"zmeow/pkg/logger"
)

// Container gerencia todas as dependências da aplicação
type Container struct {
	// Database
	DB *bun.DB

	// Repositories
	SessionRepo session.SessionRepository

	// WhatsApp
	WhatsAppManager whatsapp.WhatsAppManager

	// Use Cases
	CreateSessionUC     *sessionUseCases.CreateSessionUseCase
	ListSessionsUC      *sessionUseCases.ListSessionsUseCase
	GetSessionUC        *sessionUseCases.GetSessionUseCase
	DeleteSessionUC     *sessionUseCases.DeleteSessionUseCase
	ConnectSessionUC    *sessionUseCases.ConnectSessionUseCase
	DisconnectSessionUC *sessionUseCases.DisconnectSessionUseCase
	QRCodeUC            *sessionUseCases.GetQRCodeUseCase
	PairPhoneUC         *sessionUseCases.PairPhoneUseCase
	SetProxyUC          *sessionUseCases.SetProxyUseCase
	GetStatusUC         *sessionUseCases.GetStatusUseCase

	// Message Use Cases
	SendTextMessageUC     *messageUseCases.SendTextMessageUseCase
	SendMediaMessageUC    *messageUseCases.SendMediaMessageUseCase
	SendLocationMessageUC *messageUseCases.SendLocationMessageUseCase
	SendContactMessageUC  *messageUseCases.SendContactMessageUseCase
	SendStickerMessageUC  *messageUseCases.SendStickerMessageUseCase
	SendButtonsMessageUC  *messageUseCases.SendButtonsMessageUseCase
	SendListMessageUC     *messageUseCases.SendListMessageUseCase
	SendPollMessageUC     *messageUseCases.SendPollMessageUseCase
	EditMessageUC         *messageUseCases.EditMessageUseCase
	DeleteMessageUC       *messageUseCases.DeleteMessageUseCase
	ReactMessageUC        *messageUseCases.ReactMessageUseCase

	// Group Use Cases
	CreateGroupUC          *groupUseCases.CreateGroupUseCase
	ListGroupsUC           *groupUseCases.ListGroupsUseCase
	GetGroupInfoUC         *groupUseCases.GetGroupInfoUseCase
	UpdateParticipantsUC   *groupUseCases.UpdateParticipantsUseCase
	LeaveGroupUC           *groupUseCases.LeaveGroupUseCase
	SetGroupNameUC         *groupUseCases.SetGroupNameUseCase
	SetGroupTopicUC        *groupUseCases.SetGroupTopicUseCase
	SetGroupPhotoUC        *groupUseCases.SetGroupPhotoUseCase
	RemoveGroupPhotoUC     *groupUseCases.RemoveGroupPhotoUseCase
	SetGroupAnnounceUC     *groupUseCases.SetGroupAnnounceUseCase
	SetGroupLockedUC       *groupUseCases.SetGroupLockedUseCase
	SetDisappearingTimerUC *groupUseCases.SetDisappearingTimerUseCase
	GetInviteLinkUC        *groupUseCases.GetInviteLinkUseCase
	JoinGroupUC            *groupUseCases.JoinGroupUseCase
	GetInviteInfoUC        *groupUseCases.GetInviteInfoUseCase

	// Handlers
	SessionHandler *handlers.SessionHandler
	HealthHandler  *handlers.HealthHandler
	MessageHandler *handlers.MessageHandler
	ChatHandler    *handlers.ChatHandler
	GroupHandler   *handlers.GroupHandler

	// Logger
	Logger logger.Logger
}

// NewContainer cria um novo container de dependências
func NewContainer(db *bun.DB, whatsappManager whatsapp.WhatsAppManager) (*Container, error) {
	c := &Container{
		DB:              db,
		WhatsAppManager: whatsappManager,
		Logger:          logger.WithComponent("di-container"),
	}

	// Inicializar repositórios
	if err := c.initRepositories(); err != nil {
		return nil, err
	}

	// Inicializar use cases
	c.initUseCases()

	// Inicializar handlers
	c.initHandlers()

	c.Logger.Info().Msg("Container initialized successfully")
	return c, nil
}

// initRepositories inicializa os repositórios
func (c *Container) initRepositories() error {
	c.SessionRepo = database.NewSessionRepository(c.DB)
	return nil
}

// initUseCases inicializa os casos de uso
func (c *Container) initUseCases() {
	c.CreateSessionUC = sessionUseCases.NewCreateSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.ListSessionsUC = sessionUseCases.NewListSessionsUseCase(
		c.SessionRepo,
		c.Logger,
	)

	c.GetSessionUC = sessionUseCases.NewGetSessionUseCase(
		c.SessionRepo,
		c.Logger,
	)

	c.DeleteSessionUC = sessionUseCases.NewDeleteSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.ConnectSessionUC = sessionUseCases.NewConnectSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.QRCodeUC = sessionUseCases.NewQRCodeUseCase(
		c.WhatsAppManager,
		c.Logger,
	)

	c.PairPhoneUC = sessionUseCases.NewPairPhoneUseCase(
		c.WhatsAppManager,
		c.Logger,
	)

	c.DisconnectSessionUC = sessionUseCases.NewDisconnectSessionUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SetProxyUC = sessionUseCases.NewSetProxyUseCase(
		c.WhatsAppManager,
		c.Logger,
	)

	c.GetStatusUC = sessionUseCases.NewGetStatusUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	// Inicializar casos de uso de mensagem
	c.initMessageUseCases()

	// Inicializar casos de uso de grupo
	c.initGroupUseCases()
}

// initMessageUseCases inicializa os casos de uso de mensagem
func (c *Container) initMessageUseCases() {
	c.SendTextMessageUC = messageUseCases.NewSendTextMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SendMediaMessageUC = messageUseCases.NewSendMediaMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SendLocationMessageUC = messageUseCases.NewSendLocationMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SendContactMessageUC = messageUseCases.NewSendContactMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SendStickerMessageUC = messageUseCases.NewSendStickerMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SendButtonsMessageUC = messageUseCases.NewSendButtonsMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SendListMessageUC = messageUseCases.NewSendListMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SendPollMessageUC = messageUseCases.NewSendPollMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.EditMessageUC = messageUseCases.NewEditMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.DeleteMessageUC = messageUseCases.NewDeleteMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.ReactMessageUC = messageUseCases.NewReactMessageUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)
}

// initGroupUseCases inicializa os casos de uso de grupo
func (c *Container) initGroupUseCases() {
	c.CreateGroupUC = groupUseCases.NewCreateGroupUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.ListGroupsUC = groupUseCases.NewListGroupsUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.GetGroupInfoUC = groupUseCases.NewGetGroupInfoUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.UpdateParticipantsUC = groupUseCases.NewUpdateParticipantsUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.LeaveGroupUC = groupUseCases.NewLeaveGroupUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.SetGroupNameUC = groupUseCases.NewSetGroupNameUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		&group.PermissionValidator{},
		c.Logger,
	)

	c.SetGroupTopicUC = groupUseCases.NewSetGroupTopicUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		&group.PermissionValidator{},
		c.Logger,
	)

	c.SetGroupPhotoUC = groupUseCases.NewSetGroupPhotoUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		media.NewImageProcessor(c.Logger),
		&group.PermissionValidator{},
		c.Logger,
	)

	c.RemoveGroupPhotoUC = groupUseCases.NewRemoveGroupPhotoUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		&group.PermissionValidator{},
		c.Logger,
	)

	c.SetGroupAnnounceUC = groupUseCases.NewSetGroupAnnounceUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		&group.PermissionValidator{},
		c.Logger,
	)

	c.SetGroupLockedUC = groupUseCases.NewSetGroupLockedUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		&group.PermissionValidator{},
		c.Logger,
	)

	c.SetDisappearingTimerUC = groupUseCases.NewSetDisappearingTimerUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		&group.PermissionValidator{},
		c.Logger,
	)

	c.GetInviteLinkUC = groupUseCases.NewGetInviteLinkUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		&group.PermissionValidator{},
		c.Logger,
	)

	c.JoinGroupUC = groupUseCases.NewJoinGroupUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)

	c.GetInviteInfoUC = groupUseCases.NewGetInviteInfoUseCase(
		c.SessionRepo,
		c.WhatsAppManager,
		c.Logger,
	)
}

// initHandlers inicializa os handlers
func (c *Container) initHandlers() {
	c.SessionHandler = handlers.NewSessionHandler(
		c.CreateSessionUC,
		c.ListSessionsUC,
		c.GetSessionUC,
		c.DeleteSessionUC,
		c.ConnectSessionUC,
		c.DisconnectSessionUC,
		c.QRCodeUC,
		c.PairPhoneUC,
		c.SetProxyUC,
		c.GetStatusUC,
		c.Logger,
	)

	c.HealthHandler = handlers.NewHealthHandler()

	c.MessageHandler = handlers.NewMessageHandler(
		c.SendTextMessageUC,
		c.SendMediaMessageUC,
		c.SendLocationMessageUC,
		c.SendContactMessageUC,
		c.SendStickerMessageUC,
		c.SendButtonsMessageUC,
		c.SendListMessageUC,
		c.SendPollMessageUC,
		c.EditMessageUC,
		c.DeleteMessageUC,
		c.ReactMessageUC,
		c.Logger,
	)

	c.ChatHandler = handlers.NewChatHandler(
		c.SendTextMessageUC,
		c.SendMediaMessageUC,
		c.SendLocationMessageUC,
		c.SendContactMessageUC,
		c.SendStickerMessageUC,
		c.SendButtonsMessageUC,
		c.SendListMessageUC,
		c.SendPollMessageUC,
		c.EditMessageUC,
		c.DeleteMessageUC,
		c.ReactMessageUC,
	)

	c.GroupHandler = handlers.NewGroupHandler(
		c.CreateGroupUC,
		c.ListGroupsUC,
		c.GetGroupInfoUC,
		c.UpdateParticipantsUC,
		c.LeaveGroupUC,
		c.SetGroupNameUC,
		c.SetGroupTopicUC,
		c.SetGroupPhotoUC,
		c.RemoveGroupPhotoUC,
		c.SetGroupAnnounceUC,
		c.SetGroupLockedUC,
		c.SetDisappearingTimerUC,
		c.GetInviteLinkUC,
		c.JoinGroupUC,
		c.GetInviteInfoUC,
		c.Logger,
	)
}

// Close encerra o container e todos os seus recursos
func (c *Container) Close() error {
	c.Logger.Info().Msg("Closing container")

	// Fechar WhatsApp Manager se implementar interface de Close
	if closer, ok := c.WhatsAppManager.(interface{ Close() }); ok {
		closer.Close()
	}

	// Fechar banco de dados
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			c.Logger.WithError(err).Error().Msg("Failed to close database")
			return err
		}
	}

	c.Logger.Info().Msg("Container closed successfully")
	return nil
}
