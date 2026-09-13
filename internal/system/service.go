package system

import "context"

type Repository interface {
	Health(ctx context.Context) map[string]string
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Health(ctx context.Context) map[string]string {
	return s.repo.Health(ctx)
}
