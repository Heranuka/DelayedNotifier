// Package app wires together the application's configuration, infrastructure,
// repositories, services, HTTP handlers, and server startup.
package app

import (
	"context"
	"delay/internal/app/workers"
	"delay/internal/infra/mailer"
	"delay/internal/infra/rabbitmq"
	"delay/internal/infra/redis"
	"delay/internal/service/cache"
	"fmt"

	"delay/internal/config"
	"delay/internal/infra/postgres"
	logg "delay/internal/logger"
	reponotifications "delay/internal/repository/notifications"
	svcnotifications "delay/internal/service/notifications"
	"delay/internal/transport"
	"delay/internal/transport/http/handler/notifications"

	"github.com/wb-go/wbf/logger"
)

// App wires together the application's configuration, infrastructure,
// repositories, services, background worker, and HTTP server.
type App struct {
	Config       *config.Config
	Logger       *logger.ZerologAdapter
	Repositories *Repositories
	Services     *Services
	Worker       *workers.Worker
	Server       *transport.Server
}

// Repositories groups all repository dependencies used by the application.
type Repositories struct {
	Notifications *reponotifications.Repository
}

// Services groups all service dependencies used by the application.
type Services struct {
	Notifications *svcnotifications.Service
}

// New builds the application container by loading config, creating
// infrastructure clients, wiring repositories and services, and preparing
// the worker and HTTP server.
func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load("./configs/config.yml")
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log := logg.New(cfg)

	pgPool, err := postgres.New(cfg.Postgres.ConnectionURL, log)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	redisClient := redis.New(&cfg.Redis)

	rabbitClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, log)
	if err != nil {
		return nil, fmt.Errorf("create rabbitmq client: %w", err)
	}

	if err := rabbitClient.SetupQueues(); err != nil {
		return nil, fmt.Errorf("setup rabbitmq queues: %w", err)
	}

	email := mailer.NewSMTPMailer(&cfg.Mailer, log)

	var telegramChannel *mailer.TelegramChannel

	if cfg.Telegram.TelegramToken != "" {
		ch, err := mailer.NewTelegramChannel(cfg.Telegram.TelegramToken, 12)
		if err != nil {
			log.Warn("telegram disabled: %v", err)
		} else {
			telegramChannel = ch
		}
	}

	cacheService := cache.New(redisClient)

	repos := &Repositories{
		Notifications: reponotifications.New(pgPool),
	}

	publisher := rabbitmq.NewPublisher(rabbitClient, cfg.RabbitMQ.QueueName, cfg.RabbitMQ.ExchangeName, log)

	svcs := &Services{
		Notifications: svcnotifications.New(repos.Notifications, publisher, cacheService, cfg, log),
	}

	handler := workers.NewNotificationHandler(svcs.Notifications, email, telegramChannel, log)

	ch, err := rabbitClient.GetChannel()
	if err != nil {
		return nil, fmt.Errorf("create rabbitmq channel: %w", err)
	}

	consumer := rabbitmq.NewConsumer(ch, cfg.RabbitMQ.QueueName, handler, log).
		WithPrefetch(50).
		WithConsumerTag("worker-1")

	worker := workers.New(consumer)

	handlers := transport.Handlers{
		Notifications: notifications.New(svcs.Notifications, log),
	}

	router := transport.NewRouter(handlers)
	server := transport.New(cfg, log, router)

	return &App{
		Config:       cfg,
		Logger:       log,
		Repositories: repos,
		Services:     svcs,
		Worker:       worker,
		Server:       server,
	}, nil
}

// Run starts the HTTP server and handles graceful shutdown.
func (a *App) Run(ctx context.Context) error {
	a.Logger.Info("starting server on %s", a.Config.HTTP.Port)
	return a.Server.Run(ctx)
}
