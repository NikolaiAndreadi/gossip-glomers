package main

import (
	"slices"
	"sync"
)

type Log struct {
	mu sync.RWMutex

	committed    LogOffset
	hasCommitted bool

	entries LogEntries
}

func NewLog() *Log {
	return &Log{
		entries: make(LogEntries, 0),
	}
}

func (l *Log) Append(val LogValue) LogOffset {
	l.mu.Lock()
	defer l.mu.Unlock()

	offset := l.currOffsetLocked()
	l.entries = append(l.entries, LogEntry{Offset: offset, Value: val})
	return offset
}

func (l *Log) PollFromOffset(offset LogOffset) LogEntries {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if offset > l.currOffsetLocked() {
		return LogEntries{}
	}

	return slices.Clone(l.entries[offset:])
}

func (l *Log) Commit(offset LogOffset) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.hasCommitted = true
	l.committed = max(offset, l.committed)
}

func (l *Log) Committed() (LogOffset, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.committed, l.hasCommitted
}

func (l *Log) currOffsetLocked() LogOffset {
	return LogOffset(len(l.entries))
}
