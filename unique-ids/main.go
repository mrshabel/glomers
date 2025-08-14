// the generated id here is 12-byte. [ 4-byte epoch time, length = 8 | 5-byte random value (node id is chosen here), length = 10 | 3-byte counter value, length = 6 ]
package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

// byte order for numeric fields: MSB to LSB
var ByteOrder = binary.BigEndian

// atomic counter
var counter atomic.Uint32

func main() {
	node := maelstrom.NewNode()

	// we receive an rpc request with type='generate'. we reply with type='generate_ok' and id
	node.Handle("generate", func(req maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return err
		}

		id := getCurrentTime() + getNodeID(node.ID()) + getNextCounter()

		// update body with id and message type then send response
		body["type"] = "generate_ok"
		body["id"] = id

		return node.Reply(req, body)
	})

	// start node
	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

// getCurrentTime returns the current unix timestamp of length 8 as hex string
func getCurrentTime() string {
	buf := make([]byte, 4)
	// get the year, month, day first
	ByteOrder.PutUint32(buf, uint32(time.Now().Unix()))
	return hex.EncodeToString(buf)
}

// getNodeID returns the id of the node with length 10
func getNodeID(id string) string {
	// strip n from id if present and pad with zeroes
	id = strings.TrimPrefix(id, "n")
	if len(id) > 10 {
		id = id[:10]
	}
	return strings.Repeat("0", 10-len(id)) + id
}

// getNextCounter retrieves the next value of the counter. the length of the returned value is 6
func getNextCounter() string {
	// increase value atomically and pad with zero's up to 6 digits
	next := counter.Add(1)
	return fmt.Sprintf("%06d", next)
}
