package main

import (
	"context"
	"hash/fnv"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type ShardRouter struct {
	n *maelstrom.Node
}

func NewShardRouter(n *maelstrom.Node) *ShardRouter {
	return &ShardRouter{
		n: n,
	}
}

// Shard consistently maps a key to its owner node
func (s *ShardRouter) Shard(key LogKey) (string, bool) {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))

	ids := s.n.NodeIDs()
	shardID := ids[int(hasher.Sum32()%uint32(len(ids)))]
	return shardID, shardID == s.n.ID()
}

func (s *ShardRouter) SplitOffsets(requested map[LogKey]LogOffset) (mine map[LogKey]LogOffset, remote map[string]map[LogKey]LogOffset) {
	mine = make(map[LogKey]LogOffset)
	remote = make(map[string]map[LogKey]LogOffset)

	for key, offset := range requested {
		shardID, isMine := s.Shard(key)
		if isMine {
			mine[key] = offset
			continue
		}
		if remote[shardID] == nil {
			remote[shardID] = make(map[LogKey]LogOffset)
		}
		remote[shardID][key] = offset
	}

	return mine, remote
}

func (s *ShardRouter) SplitKeys(requested []LogKey) (mine []LogKey, remote map[string][]LogKey) {
	remote = make(map[string][]LogKey)

	for _, key := range requested {
		shardID, isMine := s.Shard(key)
		if isMine {
			mine = append(mine, key)
		} else {
			remote[shardID] = append(remote[shardID], key)
		}
	}

	return mine, remote
}

// relay forwards a sub-request to another shard and parses its typed response
func relay[Resp any](ctx context.Context, n *maelstrom.Node, dest string, req any) (Resp, error) {
	msg, err := n.SyncRPC(ctx, dest, req)
	if err != nil {
		var zero Resp
		return zero, err
	}
	return ParseRequest[Resp](msg)
}
