// 5b: distributed Kafka-style log
package main

import (
	"context"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const requestTimeout = 500 * time.Millisecond

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), requestTimeout)
}

func main() {
	n := maelstrom.NewNode()
	l := NewLogStore(maelstrom.NewLinKV(n))

	n.Handle(CommandTypeSend.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[SendRequest](msg)
		if err != nil {
			return err
		}

		ctx, cancel := requestContext()
		defer cancel()

		offset, err := l.Append(ctx, req.Key, req.Msg)
		if err != nil {
			return err
		}

		return n.Reply(msg, NewSendResponse(offset))
	})

	n.Handle(CommandTypePoll.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[PollRequest](msg)
		if err != nil {
			return err
		}

		ctx, cancel := requestContext()
		defer cancel()

		entryMap, err := l.PollFromOffsets(ctx, req.Offsets)
		if err != nil {
			return err
		}

		return n.Reply(msg, NewPollResponse(entryMap))
	})

	n.Handle(CommandTypeCommitOffsets.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[CommitOffsetsRequest](msg)
		if err != nil {
			return err
		}

		ctx, cancel := requestContext()
		defer cancel()

		if err := l.Commit(ctx, req.Offsets); err != nil {
			return err
		}

		return n.Reply(msg, NewCommitOffsetsResponse())
	})

	n.Handle(CommandTypeListCommittedOffsets.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[ListCommittedOffsetsRequest](msg)
		if err != nil {
			return err
		}

		ctx, cancel := requestContext()
		defer cancel()

		committedMap, err := l.ListCommitted(ctx, req.Keys)
		if err != nil {
			return err
		}

		return n.Reply(msg, NewListCommittedOffsetsResponse(committedMap))
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
