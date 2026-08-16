package main

import (
	"encoding/json"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type CommandType string

func (c CommandType) String() string {
	return string(c)
}

const (
	CommandTypeSend                   CommandType = "send"
	CommandTypeSendOk                 CommandType = "send_ok"
	CommandTypePoll                   CommandType = "poll"
	CommandTypePollOk                 CommandType = "poll_ok"
	CommandTypeCommitOffsets          CommandType = "commit_offsets"
	CommandTypeCommitOffsetsOk        CommandType = "commit_offsets_ok"
	CommandTypeListCommittedOffsets   CommandType = "list_committed_offsets"
	CommandTypeListCommittedOffsetsOk CommandType = "list_committed_offsets_ok"
)

func ParseRequest[T any](msg maelstrom.Message) (T, error) {
	var payload T
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

type SendRequest struct {
	Type CommandType `json:"type"`
	Key  LogKey      `json:"key"`
	Msg  LogValue    `json:"msg"`
}

type SendResponse struct {
	Type   CommandType `json:"type"`
	Offset LogOffset   `json:"offset"`
}

func NewSendResponse(offset LogOffset) *SendResponse {
	return &SendResponse{
		Type:   CommandTypeSendOk,
		Offset: offset,
	}
}

type PollRequest struct {
	Type    CommandType          `json:"type"`
	Offsets map[LogKey]LogOffset `json:"offsets"`
}

type PollResponse struct {
	Type     CommandType           `json:"type"`
	Messages map[LogKey]LogEntries `json:"msgs"`
}

func NewPollResponse(messages map[LogKey]LogEntries) *PollResponse {
	return &PollResponse{
		Type:     CommandTypePollOk,
		Messages: messages,
	}
}

type CommitOffsetsRequest struct {
	Type    CommandType          `json:"type"`
	Offsets map[LogKey]LogOffset `json:"offsets"`
}

type CommitOffsetsResponse struct {
	Type CommandType `json:"type"`
}

func NewCommitOffsetsResponse() *CommitOffsetsResponse {
	return &CommitOffsetsResponse{
		Type: CommandTypeCommitOffsetsOk,
	}
}

type ListCommittedOffsetsRequest struct {
	Type CommandType `json:"type"`
	Keys []LogKey    `json:"keys"`
}

type ListCommittedOffsetsResponse struct {
	Type    CommandType          `json:"type"`
	Offsets map[LogKey]LogOffset `json:"offsets"`
}

func NewListCommittedOffsetsResponse(offsets map[LogKey]LogOffset) *ListCommittedOffsetsResponse {
	return &ListCommittedOffsetsResponse{
		Type:    CommandTypeListCommittedOffsetsOk,
		Offsets: offsets,
	}
}
