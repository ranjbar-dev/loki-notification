package srv

func (s *Service) RegisterRoutes() {

	s.hs.RegisterPostRoute("/loki/api/v1/push", s.handleLokiPush)
}
