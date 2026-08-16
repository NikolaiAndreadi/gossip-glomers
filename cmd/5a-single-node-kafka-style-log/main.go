// 5a: single-node Kafka-style log
package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	l := NewLogStore()

	n.Handle(CommandTypeSend.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[SendRequest](msg)
		if err != nil {
			return err
		}

		offset := l.Append(req.Key, req.Msg)

		return n.Reply(msg, NewSendResponse(offset))
	})

	n.Handle(CommandTypePoll.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[PollRequest](msg)
		if err != nil {
			return err
		}

		entryMap := l.PollFromOffsets(req.Offsets)

		return n.Reply(msg, NewPollResponse(entryMap))
	})

	n.Handle(CommandTypeCommitOffsets.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[CommitOffsetsRequest](msg)
		if err != nil {
			return err
		}

		l.Commit(req.Offsets)

		return n.Reply(msg, NewCommitOffsetsResponse())
	})

	n.Handle(CommandTypeListCommittedOffsets.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[ListCommittedOffsetsRequest](msg)
		if err != nil {
			return err
		}

		committedMap := l.ListCommitted(req.Keys)

		return n.Reply(msg, NewListCommittedOffsetsResponse(committedMap))
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
