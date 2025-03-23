package web

import (
	"context"
	"log/slog"
	"strings"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	slogecho "github.com/samber/slog-echo"
	"github.com/wasilak/cloudflare-ddns/libs"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

type Server struct {
	Server *echo.Echo
	*WebServer
}

func (s *Server) Start(ctx context.Context, frameworkOptions FrameworkOptions) {
	go func() {
		s.setup()
		slog.DebugContext(ctx, "Starting server", "address", s.FrameworkOptions.ListenAddr)
		s.Server.Start(s.FrameworkOptions.ListenAddr)
	}()
}

func (s *Server) setup() {
	s.Server = echo.New()

	s.Server.HideBanner = true
	s.Server.HidePort = true

	s.Server.Debug = strings.EqualFold(s.FrameworkOptions.LogLevelConfig.Level().String(), "debug")

	s.Server.Use(slogecho.New(slog.Default()))

	if s.FrameworkOptions.OtelEnabled {
		s.Server.Use(otelecho.Middleware(libs.GetAppName(), otelecho.WithSkipper(func(c echo.Context) bool {
			return strings.Contains(c.Path(), "metrics") || strings.Contains(c.Path(), "health")
		})))
	}

	s.Server.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "metrics")
		},
	}))

	echoprometheusConfig := echoprometheus.MiddlewareConfig{
		Subsystem:  strings.ReplaceAll(libs.GetAppName(), "-", "_"),
		Registerer: prometheus.Registerer(prometheus.NewRegistry()),
	}
	s.Server.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheusConfig))

	s.Server.GET("/health", s.healthRoute)
	s.Server.GET("/api/list", s.apiList)
	s.Server.PUT("/api/", s.apiCreate)
	s.Server.POST("/api/", s.apiUpdate)
	s.Server.DELETE("/api/:zone_name/:record_name", s.apiDelete)

	s.Server.GET("/metrics", echoprometheus.NewHandler())
}
