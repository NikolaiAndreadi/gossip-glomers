// challenge #3a: Single-Node Broadcast
package main

import (
	"encoding/json"
	"log"
	"maps"
	"slices"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	memory := make(map[int]struct{})
	mu := sync.RWMutex{}

	type BroadcastPayload struct {
		Message int `json:"message"`
	}
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var payload BroadcastPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			return err
		}

		mu.Lock()
		memory[payload.Message] = struct{}{}
		mu.Unlock()

		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		mu.RLock()
		messages := make([]int, 0, len(memory))
		messages = slices.AppendSeq(messages, maps.Keys(memory))
		mu.RUnlock()

		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": messages,
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
