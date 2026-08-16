package main

import "encoding/json"

type LogKey string

type LogValue int

type LogOffset int

type LogEntry struct {
	Offset LogOffset
	Value  LogValue
}

func (e LogEntry) MarshalJSON() ([]byte, error) {
	tmp := []int{int(e.Offset), int(e.Value)}
	return json.Marshal(tmp)
}

type LogEntries []LogEntry
