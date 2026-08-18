package main

import (
	"context"
	"maps"
	"strconv"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	gossipPeriod = 500 * time.Millisecond

	nodeBits = 10
	nodeMask = 1<<nodeBits - 1
)

type Token uint64

type KeyData struct {
	Value Value `json:"value"`
	Token Token `json:"token"`
}

type Store struct {
	startOnce sync.Once

	node      *maelstrom.Node
	machineID uint64
	peers     []string

	mu        sync.RWMutex
	registers map[Key]KeyData
	counter   Token
}

func NewStore(node *maelstrom.Node) *Store {
	return &Store{
		node:      node,
		registers: make(map[Key]KeyData),
	}
}

func (s *Store) Start(ctx context.Context) {
	s.startOnce.Do(func() {
		numericID, _ := strconv.ParseUint(s.node.ID()[1:], 10, 64)
		s.machineID = numericID & nodeMask

		s.peers = make([]string, 0, len(s.node.NodeIDs())-1)

		for _, id := range s.node.NodeIDs() {
			if id != s.node.ID() {
				s.peers = append(s.peers, id)
			}
		}

		go s.run(ctx)
	})
}

func (s *Store) ApplyTxn(txn []Operation) []Operation {
	result := make([]Operation, 0, len(txn))
	s.mu.Lock()
	defer s.mu.Unlock()

	token := s.genLamportToken()

	for _, op := range txn {
		result = append(result, s.applyLocked(op, token))
	}

	return result
}

func (s *Store) Merge(data map[Key]KeyData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, otherV := range data {
		s.counter = max(s.counter, otherV.Token>>nodeBits)
		myV, ok := s.registers[k]
		if !ok || myV.Token < otherV.Token {
			s.registers[k] = otherV
		}
	}
}

func (s *Store) applyLocked(op Operation, token Token) Operation {
	result := Operation{
		Type: op.Type,
		Key:  op.Key,
	}
	switch op.Type {
	case OperationTypeWrite:
		s.registers[op.Key] = KeyData{
			Value: *op.Value,
			Token: token,
		}
		result.Value = new(*op.Value)
	case OperationTypeRead:
		if valueData, ok := s.registers[op.Key]; ok {
			result.Value = new(valueData.Value)
		}
	}
	return result
}

func (s *Store) genLamportToken() Token {
	s.counter++
	return s.counter<<nodeBits | Token(s.machineID)
}

func (s *Store) run(ctx context.Context) {
	ticker := time.NewTicker(gossipPeriod)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.flush()
		}
	}
}

func (s *Store) flush() {
	s.mu.RLock()
	data := maps.Clone(s.registers)
	s.mu.RUnlock()

	req := NewGossipRequest(data)
	for _, peer := range s.peers {
		_ = s.node.Send(peer, req)
	}
}
