// challenge #2: Unique ID Generation
package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	sf := NewSnowflake()

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"
		body["id"] = strconv.FormatInt(sf.Generate(n.ID()), 10) // precision loss inside n.Reply if sent as int
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

// Snowflake generates 64-bit identifiers with the following layout:
// [1b:sign][41b:timestamp][10b:machine_id][12b:seq]
// The node ID bits are initialized lazily because n.ID() is unavailable until Maelstrom handles the init message.
// Not exactly required, just wanted to implement it!
type Snowflake struct {
	mu            sync.Mutex
	machineIdBits *int64
	counter       int64
	lastTs        int64
}

func NewSnowflake() *Snowflake {
	return &Snowflake{
		machineIdBits: nil,
		counter:       0,
		lastTs:        -1,
	}
}

func (s *Snowflake) SetID(machineID string) {
	var padded [8]byte
	copy(padded[:], machineID)
	machineIdBytes := binary.LittleEndian.Uint64(padded[:])
	shiftedMachineId := int64(machineIdBytes&0x3FF) << 12 // 0x3FF - 10 bits
	s.machineIdBits = &shiftedMachineId
}

const epochShift = 1675728000000 // skibidi toilet released. Shift from unix so that numbers are not too big

func (s *Snowflake) Generate(machineID string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.machineIdBits == nil {
		s.SetID(machineID)
	}

	now := time.Now().UTC().UnixMilli() - epochShift

	if s.lastTs == now {
		// no overflow protection, request rate well below limit
		s.counter = (s.counter + 1) & 0xFFF // 12 bits
	} else {
		s.counter = 0
		s.lastTs = now
	}

	return (now << 22) | *s.machineIdBits | s.counter
}
