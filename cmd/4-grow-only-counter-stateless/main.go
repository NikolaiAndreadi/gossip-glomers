// challenge #4: grow-only counter, stateless variant (as intended by challenge)
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type AddPayload struct {
	Delta int `json:"delta"`
}

const keyName = "counter"

func add(ctx context.Context, kv *maelstrom.KV, delta int) error {
	for {
		current, err := kv.ReadInt(ctx, keyName)
		if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
			return err
		}
		err = kv.CompareAndSwap(ctx, keyName, current, current+delta, true)
		if err == nil {
			return nil
		}
		if maelstrom.ErrorCode(err) != maelstrom.PreconditionFailed {
			return err
		}
	}
}

func read(ctx context.Context, kv *maelstrom.KV, nodeID string) (int, error) {
	// seq-kv is sequentially consistent, so a plain read may observe an old state.
	// a no-op CAS(v, v) may be serialized in the past, at a point where the value was still v!
	// a special write of a value that never was written before cannot be ordered anywhere but the present
	// so we use it to drag this node's session to the current state
	if err := kv.Write(ctx, "sync-"+nodeID, time.Now().UnixNano()); err != nil {
		return 0, err
	}

	value, err := kv.ReadInt(ctx, keyName)
	if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return value, nil
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n.Handle("add", func(msg maelstrom.Message) error {
		var payload AddPayload
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			return err
		}

		if err := add(ctx, kv, payload.Delta); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{"type": "add_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		sum, err := read(ctx, kv, n.ID())
		if err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": sum,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
