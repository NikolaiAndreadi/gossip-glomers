// challenge #4: grow-only counter
package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	flushInterval = 100 * time.Millisecond
	kvTimeout     = 500 * time.Millisecond
)

type AddPayload struct {
	Delta int `json:"delta"`
}

type Counter struct {
	mu    sync.Mutex
	total int
}

func (c *Counter) Add(delta int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total += delta
}

func (c *Counter) Total() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	counter := &Counter{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var once sync.Once
	n.Handle("init", func(msg maelstrom.Message) error {
		nodeID := n.ID()
		once.Do(func() {
			go flushLoop(ctx, kv, nodeID, counter)
		})
		return nil
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var payload AddPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			return err
		}

		counter.Add(payload.Delta)

		return n.Reply(msg, map[string]any{"type": "add_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		sum := 0
		for _, nodeID := range n.NodeIDs() {
			val, err := readCounter(ctx, kv, nodeID)
			if err != nil {
				return err
			}
			sum += val
		}

		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": sum,
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

func flushLoop(ctx context.Context, kv *maelstrom.KV, nodeID string, counter *Counter) {
	ticker := time.NewTicker(flushInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			total := counter.Total()

			writeCtx, cancel := context.WithTimeout(ctx, kvTimeout)
			err := kv.Write(writeCtx, nodeID, total)
			cancel()

			if err != nil {
				log.Printf("kv write of %d failed: %v", total, err)
			}
		}
	}
}

func readCounter(ctx context.Context, kv *maelstrom.KV, nodeID string) (int, error) {
	readCtx, cancel := context.WithTimeout(ctx, kvTimeout)
	defer cancel()

	val, err := kv.ReadInt(readCtx, nodeID)
	if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
		return 0, nil
	}
	return val, err
}
