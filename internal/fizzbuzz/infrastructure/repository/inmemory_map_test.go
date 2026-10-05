package repository

import (
	"context"
	"errors"
	"sync"
	"testing"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
)

var (
	paramsA = domain.FizzbuzzParams{Int1: 3, Int2: 5, Limit: 16, Str1: "fizz", Str2: "buzz"}
	paramsB = domain.FizzbuzzParams{Int1: 2, Int2: 7, Limit: 10, Str1: "foo", Str2: "bar"}
)

func TestGetMostFrequent(t *testing.T) {
	tests := []struct {
		name           string
		adds           []domain.FizzbuzzParams
		expectedParams *domain.FizzbuzzParams
		expectedHits   int
	}{
		{
			name:           "empty repo returns zero stat",
			adds:           nil,
			expectedParams: nil,
			expectedHits:   0,
		},
		{
			name:           "single request",
			adds:           []domain.FizzbuzzParams{paramsA},
			expectedParams: &paramsA,
			expectedHits:   1,
		},
		{
			name:           "most frequent wins",
			adds:           []domain.FizzbuzzParams{paramsA, paramsB, paramsB, paramsA, paramsA},
			expectedParams: &paramsA,
			expectedHits:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInmemoryRepository()
			for _, p := range tt.adds {
				if err := repo.AddRequest(context.Background(), &p); err != nil {
					t.Fatalf("expects no error on AddRequest, got: %v", err)
				}
			}

			stat, err := repo.GetMostFrequent(context.Background())
			if err != nil {
				t.Fatalf("expects no error, got: %v", err)
			}
			if stat.Hits != tt.expectedHits {
				t.Fatalf("expects %d hits, got %d", tt.expectedHits, stat.Hits)
			}
			if (tt.expectedParams == nil) != (stat.Params == nil) || (tt.expectedParams != nil && *stat.Params != *tt.expectedParams) {
				t.Fatalf("expects params %+v, got %+v", tt.expectedParams, stat.Params)
			}
		})
	}
}

func TestAddRequestNil(t *testing.T) {
	repo := NewInmemoryRepository()

	err := repo.AddRequest(context.Background(), nil)
	if !errors.Is(err, domain.ErrInvalidParams) {
		t.Fatalf("expects error %v, got: %v", domain.ErrInvalidParams, err)
	}
}

// TestConcurrentAddRequest is meaningful only under `go test -race`.
func TestConcurrentAddRequest(t *testing.T) {
	const goroutines = 100
	repo := NewInmemoryRepository()

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			if err := repo.AddRequest(context.Background(), &paramsA); err != nil {
				t.Errorf("expects no error, got: %v", err)
			}
		})
	}
	wg.Wait()

	stat, err := repo.GetMostFrequent(context.Background())
	if err != nil {
		t.Fatalf("expects no error, got: %v", err)
	}
	if stat.Hits != goroutines {
		t.Fatalf("expects %d hits, got %d", goroutines, stat.Hits)
	}
}
