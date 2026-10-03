// Package service TODO
package service

import (
	"context"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
)

type Service struct {
	repo domain.FizzbuzzRepository
}

func NewService(repo domain.FizzbuzzRepository) *Service {
	return &Service{repo}
}

func (s *Service) Generate(ctx context.Context, params *domain.FizzbuzzParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	fizzbuzzed := domain.GenerateFizzbuzz(params)

	if err := s.repo.AddRequest(ctx, params); err != nil {
		return nil, err
	}
	return fizzbuzzed, nil
}

func (s *Service) GetMostFrequent(ctx context.Context) (*domain.ParamStat, error) {
	return s.repo.GetMostFrequent(ctx)
}
