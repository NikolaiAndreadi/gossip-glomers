// challenge #3d: Efficient Broadcast. #3e also passes! check 'init' handler
package main

import (
	"context"
	"encoding/json"
	"log"
	"slices"

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
		nodes := slices.Clone(n.NodeIDs())
		L := len(nodes)
		slices.Sort(nodes)

		selfIdx := slices.Index(nodes, n.ID())

		var neighbors []PeerID

		add := func(idx int) {
			if idx < 0 || idx >= L || idx == selfIdx {
				return
			}
			peer := PeerID(nodes[idx])
			if !slices.Contains(neighbors, peer) {
				neighbors = append(neighbors, peer)
			}
		}

		const hubCount = 2
		if selfIdx < hubCount {
			for i := 0; i < L; i++ {
				add(i)
			}
		} else {
			for i := 0; i < hubCount; i++ {
				add(i)
			}
		}

		g.Start(ctx, neighbors)
		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
