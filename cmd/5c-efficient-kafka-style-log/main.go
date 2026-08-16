// 5c: efficient Kafka-style log
// CAS contention was not a problem in 5b. tho 7 server msgs per op was! solution:
// 1) take 5a: data live in the nodes again
// 2) deterministic sharding between nodes
// 3) give answer if node owns data. relay to another node if not (see point #2)
// 5a already basically linearizable by mutex. multi-key fetch is a pain but straightforward.
// 5c doesn't have a partition nemesis and no dying nodes, which GREATLY simplifies the implementation
package main

import (
	"context"
	"log"
	"maps"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const requestTimeout = 500 * time.Millisecond

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), requestTimeout)
}

func main() {
	n := maelstrom.NewNode()
	l := NewLogStore()
	s := NewShardRouter(n)

	n.Handle(CommandTypeSend.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[SendRequest](msg)
		if err != nil {
			return err
		}

		shardID, isMine := s.Shard(req.Key)
		if isMine {
			offset := l.Append(req.Key, req.Msg)
			return n.Reply(msg, NewSendResponse(offset))
		}

		ctx, cancel := requestContext()
		defer cancel()

		resp, err := relay[SendResponse](ctx, n, shardID, req)
		if err != nil {
			return err
		}

		return n.Reply(msg, resp)
	})

	n.Handle(CommandTypePoll.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[PollRequest](msg)
		if err != nil {
			return err
		}

		mine, remote := s.SplitOffsets(req.Offsets)
		entryMap := l.PollFromOffsets(mine)

		ctx, cancel := requestContext()
		defer cancel()

		for shardID, offsets := range remote {
			resp, err := relay[PollResponse](ctx, n, shardID, PollRequest{Type: CommandTypePoll, Offsets: offsets})
			if err != nil {
				return err
			}
			maps.Copy(entryMap, resp.Messages)
		}

		return n.Reply(msg, NewPollResponse(entryMap))
	})

	n.Handle(CommandTypeCommitOffsets.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[CommitOffsetsRequest](msg)
		if err != nil {
			return err
		}

		mine, remote := s.SplitOffsets(req.Offsets)
		l.Commit(mine)

		ctx, cancel := requestContext()
		defer cancel()

		for shardID, offsets := range remote {
			if _, err := relay[CommitOffsetsResponse](ctx, n, shardID, CommitOffsetsRequest{Type: CommandTypeCommitOffsets, Offsets: offsets}); err != nil {
				return err
			}
		}

		return n.Reply(msg, NewCommitOffsetsResponse())
	})

	n.Handle(CommandTypeListCommittedOffsets.String(), func(msg maelstrom.Message) error {
		req, err := ParseRequest[ListCommittedOffsetsRequest](msg)
		if err != nil {
			return err
		}

		mine, remote := s.SplitKeys(req.Keys)
		committedMap := l.ListCommitted(mine)

		ctx, cancel := requestContext()
		defer cancel()

		for shardID, keys := range remote {
			resp, err := relay[ListCommittedOffsetsResponse](
				ctx, n, shardID, ListCommittedOffsetsRequest{Type: CommandTypeListCommittedOffsets, Keys: keys},
			)
			if err != nil {
				return err
			}
			maps.Copy(committedMap, resp.Offsets)
		}

		return n.Reply(msg, NewListCommittedOffsetsResponse(committedMap))
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
