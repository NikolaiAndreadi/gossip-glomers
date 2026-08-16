package main

import "sync"

type LogStore struct {
	mu sync.RWMutex

	logs map[LogKey]*Log
}

func NewLogStore() *LogStore {
	return &LogStore{
		logs: make(map[LogKey]*Log),
	}
}

func (l *LogStore) getOrCreate(key LogKey) *Log {
	if store, exists := l.get(key); exists {
		return store
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	store, exists := l.logs[key]
	if !exists {
		store = NewLog()
		l.logs[key] = store
	}

	return store
}

func (l *LogStore) get(key LogKey) (*Log, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	store, exists := l.logs[key]
	return store, exists
}

func (l *LogStore) Append(key LogKey, val LogValue) LogOffset {
	store := l.getOrCreate(key)
	return store.Append(val)
}

func (l *LogStore) PollFromOffsets(requested map[LogKey]LogOffset) map[LogKey]LogEntries {
	result := make(map[LogKey]LogEntries, len(requested))

	for key, offset := range requested {
		if store, exists := l.get(key); exists {
			result[key] = store.PollFromOffset(offset)
		}
	}

	return result
}

func (l *LogStore) Commit(requested map[LogKey]LogOffset) {
	for key, offset := range requested {
		store := l.getOrCreate(key)
		store.Commit(offset)
	}
}

func (l *LogStore) ListCommitted(requested []LogKey) map[LogKey]LogOffset {
	result := make(map[LogKey]LogOffset, len(requested))

	for _, key := range requested {
		if store, exists := l.get(key); exists {
			if committedOffset, isCommittedBefore := store.Committed(); isCommittedBefore {
				result[key] = committedOffset
			}
		}
	}

	return result
}
