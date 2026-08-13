// challenge #3b: Multi-Node Broadcast
package main

import (
	"context"
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastPayload struct {
	Message StoredItem `json:"message"`
}

func main() {
	n := maelstrom.NewNode()

	s := NewStore()
	g := NewGossip(n, s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var payload BroadcastPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			return err
		}
		s.Add(payload.Message)
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("gossip", func(msg maelstrom.Message) error {
		var payload GossipMessage
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			return err
		}
		s.Add(payload.Messages...)
		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": s.All(),
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	n.Handle("init", func(msg maelstrom.Message) error {
		var peers []PeerID
		for _, id := range n.NodeIDs() {
			if id != n.ID() {
				peers = append(peers, PeerID(id))
			}
		}
		g.Start(ctx, peers)
		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
