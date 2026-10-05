package service

import (
	"context"
	"errors"
	"slices"
	"testing"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
)

var limits = domain.Limits{MaxLimit: 100, MaxStrLength: 10}

// fakeRepo inplements FizzbuzzRepository interface.
// It returns canned values and errors for repository
type fakeRepo struct {
	calls   int
	addErr  error
	stat    *domain.ParamStat
	statErr error
}

// AddRequest just increase the # of calls
func (f *fakeRepo) AddRequest(_ context.Context, _ *domain.FizzbuzzParams) error {
	f.calls++
	return f.addErr
}

// GetMostFrequent returnr a set stat
func (f *fakeRepo) GetMostFrequent(_ context.Context) (*domain.ParamStat, error) {
	return f.stat, f.statErr
}

func TestGenerate(t *testing.T) {
	errRepo := errors.New("repo down")

	tests := []struct {
		name              string
		params            domain.FizzbuzzParams
		addErr            error
		expectedGenerated []string
		expectedErr       error
		expectedNbCalls   int
	}{
		{
			name:              "valid params",
			params:            domain.FizzbuzzParams{Int1: 2, Int2: 3, Limit: 6, Str1: "fizz", Str2: "buzz"},
			expectedGenerated: []string{"1", "fizz", "buzz", "fizz", "5", "fizzbuzz"},
			expectedNbCalls:   1,
		},
		{
			name:            "invalid params are not recorded",
			params:          domain.FizzbuzzParams{Int1: 0, Int2: 3, Limit: 6, Str1: "fizz", Str2: "buzz"},
			expectedErr:     domain.ErrInvalidParams,
			expectedNbCalls: 0,
		},
		{
			name:            "repo error propagates",
			params:          domain.FizzbuzzParams{Int1: 2, Int2: 3, Limit: 6, Str1: "fizz", Str2: "buzz"},
			addErr:          errRepo,
			expectedErr:     errRepo,
			expectedNbCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{addErr: tt.addErr}
			svc := NewService(repo, limits)

			generated, err := svc.Generate(context.Background(), &tt.params)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expects error %v, got: %v", tt.expectedErr, err)
			}
			if !slices.Equal(generated, tt.expectedGenerated) {
				t.Fatalf("expects %v, got %v", tt.expectedGenerated, generated)
			}
			if repo.calls != tt.expectedNbCalls {
				t.Fatalf("expects %d AddRequest calls, got %d", tt.expectedNbCalls, repo.calls)
			}
		})
	}
}

func TestGetMostFrequent(t *testing.T) {
	errRepo := errors.New("repo down")
	stat := &domain.ParamStat{Params: &domain.FizzbuzzParams{Int1: 3, Int2: 5, Limit: 16, Str1: "fizz", Str2: "buzz"}, Hits: 2}

	tests := []struct {
		name              string
		repo              *fakeRepo
		expectedParamStat *domain.ParamStat
		ExpectedStatErr   error
	}{
		{
			name:              "returns repo stat",
			repo:              &fakeRepo{stat: stat},
			expectedParamStat: stat,
		},
		{
			name:            "repo error propagates",
			repo:            &fakeRepo{statErr: errRepo},
			ExpectedStatErr: errRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, limits)

			paramStats, err := svc.GetMostFrequent(context.Background())
			if !errors.Is(err, tt.ExpectedStatErr) {
				t.Fatalf("expects error %v, got: %v", tt.ExpectedStatErr, err)
			}
			if paramStats != tt.expectedParamStat {
				t.Fatalf("expects %+v, got %+v", tt.expectedParamStat, paramStats)
			}
		})
	}
}
