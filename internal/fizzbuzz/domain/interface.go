package domain

import "context"

type FizzbuzzService interface {
	Generate(ctx context.Context, params *FizzbuzzParams) ([]string, error)
	GetMostFrequent(ctx context.Context) (*ParamStat, error)
}

type FizzbuzzRepository interface {
	AddRequest(ctx context.Context, params *FizzbuzzParams) error
	GetMostFrequent(ctx context.Context) (*ParamStat, error)
}
