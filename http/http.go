package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"telecommunications_repair_hub/config"
	"telecommunications_repair_hub/pkg/response"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type HttpServer struct {
	Port   string
	Host   string
	config *config.Config
}

func NewHttpServer(config *config.Config) *HttpServer {
	return &HttpServer{
		Port:   config.App.Port,
		Host:   config.App.Host,
		config: config,
	}
}

func (h *HttpServer) init(e *Server) {
	e.HTTPErrorHandler = func(err error, ctx echo.Context) {
		if ctx.Response().Committed {
			return
		}
		response.NewResponse(ctx).
			SetStatus(http.StatusInternalServerError).
			SetMessage(err.Error()).
			Error(err)

		slog.Error("[HttpServer] HTTPErrorHandler", "Method", ctx.Request().Method,
			"Path", ctx.Request().URL.Path, "Error", err)

	}

}

func (h *HttpServer) Start(ctx context.Context) error {
	e := NewServer(h.config)
	h.init(e)

	go func() {
		<-ctx.Done()
		e.Shutdown(ctx)
	}()

	slog.Info("[HttpServer] Register Routes")
	NewBaseRouter(e).RegisterRoutes()

	slog.Info("[HttpServer] Start", "Host", h.Host, "Port", h.Port)
	err := e.Start(h.Host, h.Port)
	if err != nil {
		slog.Error("[HttpServer] Start", "Error", err)
		return err
	}
	return err
}

func (h *HttpServer) HttpLogLevel() log.Lvl {
	switch strings.ToLower(h.config.App.LogLevel) {
	case "debug":
		return log.DEBUG
	case "info":
		return log.INFO
	case "warn":
		return log.WARN
	case "error":
		return log.ERROR
	default:
		return log.INFO
	}
}
