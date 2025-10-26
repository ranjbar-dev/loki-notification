package srv

import (
	"github.com/ranjbar-dev/loki-notification/internal/config"
	"github.com/ranjbar-dev/loki-notification/internal/httpserver"
	"go.uber.org/zap"
)

type Service struct {
	cfg *config.Config
	log *zap.Logger
	hs  *httpserver.HttpServer
}

func NewService(cfg *config.Config, log *zap.Logger, hs *httpserver.HttpServer) *Service {

	return &Service{cfg: cfg, log: log, hs: hs}
}

func (s *Service) Start() error {

	s.RegisterRoutes()

	return s.hs.Serve()
}
