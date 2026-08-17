package main

import "sync"

type Store struct {
	mu        sync.Mutex
	registers map[Key]Value
}

func NewStore() *Store {
	return &Store{
		registers: make(map[Key]Value),
	}
}

func (s *Store) ApplyTxn(txn []Operation) []Operation {
	result := make([]Operation, 0, len(txn))
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, op := range txn {
		result = append(result, s.applyLocked(op))
	}

	return result
}

func (s *Store) applyLocked(op Operation) Operation {
	result := Operation{
		Type: op.Type,
		Key:  op.Key,
	}
	switch op.Type {
	case OperationTypeWrite:
		s.registers[op.Key] = *op.Value
		result.Value = new(*op.Value)
	case OperationTypeRead:
		if value, ok := s.registers[op.Key]; ok {
			result.Value = new(value)
		}
	}
	return result
}
