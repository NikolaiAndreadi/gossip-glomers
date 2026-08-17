// Challenge #6a: Single-Node, Totally-Available Transactions
package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	store := NewStore()

	n.Handle(CommandTypeTxn.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[TxnRequest](msg)
		if err != nil {
			return err
		}

		resp := NewTxnResponse(store.ApplyTxn(req.Txn))

		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
