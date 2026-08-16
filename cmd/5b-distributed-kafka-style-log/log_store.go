package main

import (
	"context"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type LogStore struct {
	log *Log
}

func NewLogStore(kv *maelstrom.KV) *LogStore {
	return &LogStore{
		log: NewLog(kv),
	}
}

func (l *LogStore) Append(ctx context.Context, key LogKey, val LogValue) (LogOffset, error) {
	return l.log.Append(ctx, key, val)
}

func (l *LogStore) PollFromOffsets(ctx context.Context, requested map[LogKey]LogOffset) (map[LogKey]LogEntries, error) {
	result := make(map[LogKey]LogEntries, len(requested))

	for key, offset := range requested {
		entries, err := l.log.PollFromOffset(ctx, key, offset)
		if err != nil {
			return nil, err
		}
		result[key] = entries
	}

	return result, nil
}

func (l *LogStore) Commit(ctx context.Context, requested map[LogKey]LogOffset) error {
	for key, offset := range requested {
		if err := l.log.Commit(ctx, key, offset); err != nil {
			return err
		}
	}
	return nil
}

func (l *LogStore) ListCommitted(ctx context.Context, requested []LogKey) (map[LogKey]LogOffset, error) {
	result := make(map[LogKey]LogOffset, len(requested))

	for _, key := range requested {
		offset, hasCommitted, err := l.log.Committed(ctx, key)
		if err != nil {
			return nil, err
		}
		if hasCommitted {
			result[key] = offset
		}
	}

	return result, nil
}
