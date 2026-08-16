package main

import (
	"context"
	"slices"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	logKeyPrefix       = "log_"
	committedKeyPrefix = "committed_"
)

type Log struct {
	kv *maelstrom.KV
}

func NewLog(kv *maelstrom.KV) *Log {
	return &Log{
		kv: kv,
	}
}

func (l *Log) Append(ctx context.Context, key LogKey, val LogValue) (LogOffset, error) {
	kvKey := logKeyPrefix + key.String()
	for {
		var current LogEntries
		err := l.kv.ReadInto(ctx, kvKey, &current)
		if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
			return 0, err
		}

		offset := LogOffset(0)
		if len(current) > 0 {
			offset = current[len(current)-1].Offset + 1
		}

		next := append(slices.Clip(current), LogEntry{Offset: offset, Value: val})

		err = l.kv.CompareAndSwap(ctx, kvKey, current, next, true)
		if err == nil {
			return offset, nil
		}
		if maelstrom.ErrorCode(err) != maelstrom.PreconditionFailed {
			return 0, err
		}
	}
}

func (l *Log) PollFromOffset(ctx context.Context, key LogKey, offset LogOffset) (LogEntries, error) {
	var current LogEntries
	err := l.kv.ReadInto(ctx, logKeyPrefix+key.String(), &current)
	if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
		return nil, err
	}

	result := make(LogEntries, 0, len(current))
	for _, entry := range current {
		if entry.Offset >= offset {
			result = append(result, entry)
		}
	}

	return result, nil
}

func (l *Log) Commit(ctx context.Context, key LogKey, offset LogOffset) error {
	kvKey := committedKeyPrefix + key.String()
	for {
		current, hasCommitted, err := l.readCommitted(ctx, kvKey)
		if err != nil {
			return err
		}
		if hasCommitted && current >= offset {
			return nil
		}

		err = l.kv.CompareAndSwap(ctx, kvKey, current, offset, !hasCommitted)
		if err == nil {
			return nil
		}
		if maelstrom.ErrorCode(err) != maelstrom.PreconditionFailed {
			return err
		}
	}
}

func (l *Log) Committed(ctx context.Context, key LogKey) (LogOffset, bool, error) {
	return l.readCommitted(ctx, committedKeyPrefix+key.String())
}

func (l *Log) readCommitted(ctx context.Context, kvKey string) (LogOffset, bool, error) {
	var offset LogOffset
	if err := l.kv.ReadInto(ctx, kvKey, &offset); err != nil {
		if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
			return 0, false, nil
		}
		return 0, false, err
	}
	return offset, true, nil
}
