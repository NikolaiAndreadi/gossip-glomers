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
	CommandTypeTxn    CommandType = "txn"
	CommandTypeTxnOk  CommandType = "txn_ok"
	CommandTypeGossip CommandType = "gossip"
	CommandTypeInit   CommandType = "init"
)

func ParseRequest[T any](msg maelstrom.Message) (T, error) {
	var payload T
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

type OperationType string

const (
	OperationTypeRead  OperationType = "r"
	OperationTypeWrite OperationType = "w"
)

type Key int

type Value int

type Operation struct {
	Type  OperationType
	Key   Key
	Value *Value
}

func (op Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal([3]any{op.Type, op.Key, op.Value})
}

func (op *Operation) UnmarshalJSON(data []byte) error {
	var tmp [3]json.RawMessage
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[0], &op.Type); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[1], &op.Key); err != nil {
		return err
	}
	if err := json.Unmarshal(tmp[2], &op.Value); err != nil {
		return err
	}
	return nil
}

type TxnRequest struct {
	Type CommandType `json:"type"`
	Txn  []Operation `json:"txn"`
}

type TxnResponse struct {
	Type CommandType `json:"type"`
	Txn  []Operation `json:"txn"`
}

func NewTxnResponse(txn []Operation) *TxnResponse {
	return &TxnResponse{
		Type: CommandTypeTxnOk,
		Txn:  txn,
	}
}

type GossipRequest struct {
	Type CommandType     `json:"type"`
	Data map[Key]KeyData `json:"data"`
}

func NewGossipRequest(data map[Key]KeyData) *GossipRequest {
	return &GossipRequest{
		Type: CommandTypeGossip,
		Data: data,
	}
}
