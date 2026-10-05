package repository

import (
	"context"
	"fmt"
	"sync"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
)

type InmemoryRepository struct {
	Requests map[domain.FizzbuzzParams]int
	mutex    sync.RWMutex
}

func NewInmemoryRepository() *InmemoryRepository {
	return &InmemoryRepository{
		Requests: make(map[domain.FizzbuzzParams]int),
		mutex:    sync.RWMutex{},
	}
}

func (r *InmemoryRepository) AddRequest(_ctx context.Context, params *domain.FizzbuzzParams) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if params == nil {
		return fmt.Errorf("%w: params cannot be nil", domain.ErrInvalidParams)
	}

	r.Requests[*params] += 1
	return nil
}

func (r *InmemoryRepository) GetMostFrequent(_ctx context.Context) (*domain.ParamStat, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var mostFrequent domain.ParamStat
	for key, count := range r.Requests {
		if mostFrequent.Params == nil {
			mostFrequent.Params = &key
			mostFrequent.Hits = count
			continue
		}

		if mostFrequent.Hits < count {
			mostFrequent.Params = &key
			mostFrequent.Hits = count
		}
	}
	return &mostFrequent, nil
}
