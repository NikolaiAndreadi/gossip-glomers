// Challenges #6b & #6c: Totally-Available Transactions
package main

import (
	"context"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := maelstrom.NewNode()
	store := NewStore(n)

	n.Handle(CommandTypeTxn.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[TxnRequest](msg)
		if err != nil {
			return err
		}

		resp := NewTxnResponse(store.ApplyTxn(req.Txn))

		return n.Reply(msg, resp)
	})

	n.Handle(CommandTypeGossip.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[GossipRequest](msg)
		if err != nil {
			return err
		}

		store.Merge(req.Data)

		return nil
	})

	n.Handle(CommandTypeInit.String(), func(msg maelstrom.Message) error {
		store.Start(ctx)
		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
