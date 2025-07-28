package main

import (
	"context"

	"strings"

	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/GoYoko/web"

	"github.com/chaitin/MonkeyCode/backend/config"
	"github.com/chaitin/MonkeyCode/backend/docs"
	"github.com/chaitin/MonkeyCode/backend/internal"
	"github.com/chaitin/MonkeyCode/backend/pkg"
	"github.com/chaitin/MonkeyCode/backend/pkg/service"
	"github.com/chaitin/MonkeyCode/backend/pkg/store"
)

// @title MonkeyCode API
// @version 1.0
// @description MonkeyCode API
func main() {
	s, err := newServer()
	if err != nil {
		panic(err)
	}

	s.version.Print()
	s.logger.With("config", s.config).Debug("config")

	if err := store.MigrateSQL(s.config, s.logger); err != nil {
		panic(err)
	}

	if err := s.userV1.InitAdmin(); err != nil {
		panic(err)
	}

	if s.config.Debug {
		s.web.Swagger("MonkeyCode API", "/reference", string(docs.SwaggerJSON), web.WithBasicAuth("mc", "mc88"))
	}

	s.web.PrintRoutes()

	// Serve frontend static files
	e := s.web.Echo()
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "/app/ui/dist",
		Index:      "index.html",
		HTML5:      true,
		Browse:     false,
		IgnoreBase: true,
		Skipper: func(c echo.Context) bool {
			// Skip API routes
			return strings.HasPrefix(c.Request().URL.Path, "/api") ||
				strings.HasPrefix(c.Request().URL.Path, "/v1") ||
				strings.HasPrefix(c.Request().URL.Path, "/reference")
		},
	}))

	if err := s.report.ReportInstallation(); err != nil {
		panic(err)
	}

	s.euse.SyncLatest()

	svc := service.NewService(service.WithPprof())
	svc.Add(s)
	if err := svc.Run(); err != nil {
		panic(err)
	}
}

// Name implements service.Servicer.
func (s *Server) Name() string {
	return "Server"
}

// Start implements service.Servicer.
func (s *Server) Start() error {
	return s.web.Run(s.config.Server.Addr)
}

// Stop implements service.Servicer.
func (s *Server) Stop() error {
	return s.web.Echo().Shutdown(context.Background())
}

//lint:ignore U1000 unused for wire
var appSet = wire.NewSet(
	wire.FieldsOf(new(*config.Config), "Logger"),
	config.Init,
	pkg.Provider,
	internal.Provider,
)
