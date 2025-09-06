package server

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"telecommunications_repair_hub/config"
	"telecommunications_repair_hub/http"
	"telecommunications_repair_hub/pkg/logger"
)

func NewTelecommunicationsServer() {
	cfg := config.InitConfig()
	logger.NewLogger(cfg.App.LogLevel,
		cfg.App.LogOutput,
		cfg.App.Logger.Rotation,
		cfg.App.Logger.RotationTime,
		cfg.App.Logger.RotationSize,
		cfg.App.Logger.RotationCount,
	).Init()

	TelecommunicationsServer := http.NewHttpServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	address := net.JoinHostPort(cfg.App.Host, cfg.App.Port)
	go func() {
		slog.Info("TelecommunicationsServer is starting", "address", address)
		TelecommunicationsServer.Start(ctx)
	}()

	<-signalChan

	cancel()
}
