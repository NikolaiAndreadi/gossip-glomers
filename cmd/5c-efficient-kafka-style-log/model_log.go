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

func (e *LogEntry) UnmarshalJSON(data []byte) error {
	var tmp [2]int
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	e.Offset = LogOffset(tmp[0])
	e.Value = LogValue(tmp[1])
	return nil
}

type LogEntries []LogEntry
