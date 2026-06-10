package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexandrKudryavtsev/go-job-queue/config"
	"github.com/AlexandrKudryavtsev/go-job-queue/internal/queue"
	"github.com/AlexandrKudryavtsev/go-job-queue/pkg/httpserver"
	"github.com/AlexandrKudryavtsev/go-job-queue/pkg/logger"
)

func Run(cfg *config.Config) error {
	log := logger.New(cfg.Logger)
	log.Info("create logger")

	appQueue := queue.New(queue.Config{
		VisibilityTimeout: cfg.Queue.VisibilityTimeout.Duration,
		RetryBaseDelay:    cfg.Queue.RetryBaseDelay.Duration,
		MaxPayloadSize:    cfg.Queue.MaxPayloadSize,
		SweepInterval:     cfg.Queue.SweepInterval.Duration,
	})

	queueContext, cancelQueueContext := context.WithCancel(context.Background())
	defer cancelQueueContext()

	go appQueue.Start(queueContext)

	mx := http.NewServeMux()

	httpServer := httpserver.New(
		mx,
		httpserver.Port(cfg.Server.Port),
		httpserver.ReadTimeout(cfg.Server.ReadTimeout.Duration),
		httpserver.WriteTimeout(cfg.Server.WriteTimeout.Duration),
		httpserver.ShutdownTimeout(cfg.Server.ShutdownTimeout.Duration),
	)

	httpServer.Start()
	log.Info("http server starting", "port", cfg.Server.Port)

	notify := httpServer.Notify()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(quit)

	select {
	case err, ok := <-notify:

		if ok && err != nil {
			log.Error("http server", "error", err)
			return err
		}
		log.Info("http server stopped")
	case sig := <-quit:
		log.Info("received signal", "signal", sig)

		if err := httpServer.Shutdown(); err != nil {
			log.Error("shutdown http server", "error", err)
			return err
		}
		log.Info("http server stopped")
	}

	return nil
}
