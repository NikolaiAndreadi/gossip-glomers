package main

import (
	"maps"
	"slices"
	"sync"
)

type StoredItem int

type StoredSet map[StoredItem]struct{}

type Store struct {
	mu  sync.RWMutex
	set StoredSet
}

func NewStore() *Store {
	return &Store{
		set: make(StoredSet),
	}
}

func (s *Store) Add(values ...StoredItem) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, value := range values {
		s.set[value] = struct{}{}
	}
}

func (s *Store) All() []StoredItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]StoredItem, 0, len(s.set))
	values = slices.AppendSeq(values, maps.Keys(s.set))
	return values
}
