package main

import (
	"context"
	"slices"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type PeerID string

type Gossip struct {
	once  sync.Once
	node  *maelstrom.Node
	store *Store
	peers []PeerID
}

func NewGossip(node *maelstrom.Node, store *Store) *Gossip {
	g := &Gossip{
		node:  node,
		peers: nil,
		store: store,
	}
	return g
}

func (g *Gossip) Start(ctx context.Context, peers []PeerID) {
	if len(peers) == 0 {
		return
	}
	g.once.Do(func() {
		g.peers = slices.Clone(peers)
		go g.run(ctx)
	})
}

func (g *Gossip) run(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.flush()
		}
	}
}

type GossipMessage struct {
	Type     string       `json:"type"`
	Messages []StoredItem `json:"messages"`
}

func (g *Gossip) flush() {
	snapshot := g.store.All()
	for _, peer := range g.peers {
		_ = g.node.Send(string(peer), GossipMessage{
			Type:     "gossip",
			Messages: snapshot,
		})
	}
}
