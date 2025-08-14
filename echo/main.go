package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	// instantiate maelstrom node
	node := maelstrom.NewNode()

	// register echo handler by simply unmarshalling the request body and replying with an 'echo_ok'
	node.Handle("echo", func(req maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return err
		}

		body["type"] = "echo_ok"

		// send reply with updated body
		return node.Reply(req, body)
	})

	// finally run node. this reads input from stdin and fires a goroutine for each message read
	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}
