package ai

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]Envelope
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{entries: make(map[string]Envelope)} }

func (s *MemoryStore) Get(_ context.Context, key CacheKey) (Envelope, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for storedKey, value := range s.entries {
		if storedKey == cacheID(key, value.Model) {
			return value, true, nil
		}
	}
	return Envelope{}, false, nil
}

func (s *MemoryStore) Put(_ context.Context, key CacheKey, value Envelope, _ GenerateResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[cacheID(key, key.Model)] = value
	return nil
}

func cacheID(key CacheKey, model string) string {
	return key.Task + "|" + key.PromptVersion + "|" + key.InputFingerprint + "|" + key.Provider + "|" + model
}

type MemoryBudget struct {
	mu        sync.Mutex
	Limit     int64
	Spent     int64
	Responses int
}

func (b *MemoryBudget) Allow(_ context.Context, _ string, maximum int64) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Limit <= 0 || b.Spent+maximum <= b.Limit, nil
}

func (b *MemoryBudget) Record(_ context.Context, _ string, response GenerateResponse) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Spent += response.CostMicro
	b.Responses++
	return nil
}
