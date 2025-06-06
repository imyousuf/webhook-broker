// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/newscred/webhook-broker/config"
	"github.com/newscred/webhook-broker/controllers"
	"github.com/newscred/webhook-broker/dispatcher"
	"github.com/newscred/webhook-broker/scheduler"
	"github.com/newscred/webhook-broker/storage"
)

// Injectors from wire.go:

// GetAppVersion retrieves the app version
func GetAppVersion() config.AppVersion {
	appVersion := config.GetVersion()
	return appVersion
}

// GetHTTPServer returns the server container with all adjacent data
func GetHTTPServer(cliConfig *config.CLIConfig) (*HTTPServiceContainer, error) {
	configConfig, err := config.GetConfigurationFromCLIConfig(cliConfig)
	if err != nil {
		return nil, err
	}
	serverLifecycleListenerImpl := NewServerListener()
	migrationConfig := GetMigrationConfig(cliConfig)
	dataAccessor, err := storage.GetNewDataAccessor(configConfig, migrationConfig, configConfig)
	if err != nil {
		return nil, err
	}
	appRepository := newAppRepository(dataAccessor)
	statusController := controllers.NewStatusController(appRepository)
	producerRepository := newProducerRepository(dataAccessor)
	producerController := controllers.NewProducerController(producerRepository)
	producersController := controllers.NewProducersController(producerRepository, producerController)
	channelRepository := newChannelRepository(dataAccessor)
	consumerRepository := newConsumerRepository(dataAccessor)
	messageRepository := newMessageRepository(dataAccessor)
	deliveryJobRepository := newDeliveryJobRepository(dataAccessor)
	messageController := controllers.NewMessageController(messageRepository, deliveryJobRepository)
	jobRequeueController := controllers.NewJobRequeueController(deliveryJobRepository, channelRepository, consumerRepository)
	dlqController := controllers.NewDLQController(messageController, jobRequeueController, deliveryJobRepository, consumerRepository)
	consumerController := controllers.NewConsumerController(channelRepository, consumerRepository, dlqController)
	consumersController := controllers.NewConsumersController(consumerController, consumerRepository)
	messagesController := controllers.NewMessagesController(messageController, messageRepository)
	scheduledMessageRepository := newScheduledMessageRepository(dataAccessor)
	lockRepository := newLockRepository(dataAccessor)
	metricsContainer := dispatcher.NewMetricsContainer()
	configuration := &dispatcher.Configuration{
		DeliveryJobRepo:          deliveryJobRepository,
		ConsumerRepo:             consumerRepository,
		LockRepo:                 lockRepository,
		BrokerConfig:             configConfig,
		ConsumerConnectionConfig: configConfig,
		MsgRepo:                  messageRepository,
		MetricsCollector:         metricsContainer,
	}
	messageDispatcher := dispatcher.NewMessageDispatcher(configuration)
	broadcastController := controllers.NewBroadcastController(channelRepository, messageRepository, producerRepository, scheduledMessageRepository, messageDispatcher)
	scheduledMessagesController := controllers.NewScheduledMessagesController(scheduledMessageRepository, channelRepository)
	messagesStatusController := controllers.NewMessagesStatusController(messagesController, scheduledMessagesController, messageRepository, scheduledMessageRepository)
	channelController := controllers.NewChannelController(consumersController, messagesController, broadcastController, messagesStatusController, channelRepository)
	jobsController := controllers.NewJobsController(channelRepository, consumerRepository, deliveryJobRepository)
	jobController := controllers.NewJobController(messageController, channelController, producerController, consumerController, channelRepository, consumerRepository, deliveryJobRepository)
	channelsController := controllers.NewChannelsController(channelRepository, channelController)
	jobStatusController := controllers.NewJobStatusController(deliveryJobRepository)
	scheduledMessageController := controllers.NewScheduledMessageController(scheduledMessageRepository, channelRepository)
	handler := dispatcher.NewPrometheusHandler()
	controllersControllers := &controllers.Controllers{
		StatusController:            statusController,
		ProducersController:         producersController,
		ProducerController:          producerController,
		ChannelController:           channelController,
		ConsumerController:          consumerController,
		ConsumersController:         consumersController,
		JobsController:              jobsController,
		JobController:               jobController,
		BroadcastController:         broadcastController,
		MessageController:           messageController,
		MessagesController:          messagesController,
		DLQController:               dlqController,
		ChannelsController:          channelsController,
		MessagesStatusController:    messagesStatusController,
		JobRequeueController:        jobRequeueController,
		JobStatusController:         jobStatusController,
		ScheduledMessageController:  scheduledMessageController,
		ScheduledMessagesController: scheduledMessagesController,
		MetricsHandler:              handler,
	}
	router := controllers.NewRouter(controllersControllers)
	server := controllers.ConfigureAPI(configConfig, serverLifecycleListenerImpl, router)
	schedulerConfiguration := scheduler.NewSchedulerConfiguration(scheduledMessageRepository, messageRepository, messageDispatcher, lockRepository, configConfig)
	messageScheduler := scheduler.NewMessageScheduler(schedulerConfiguration)
	httpServiceContainer := &HTTPServiceContainer{
		Configuration: configConfig,
		Server:        server,
		DataAccessor:  dataAccessor,
		Listener:      serverLifecycleListenerImpl,
		Dispatcher:    messageDispatcher,
		Scheduler:     messageScheduler,
	}
	return httpServiceContainer, nil
}
