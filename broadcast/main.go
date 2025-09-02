package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Data struct {
	messages map[int]struct{}
	mu       sync.RWMutex
}

func (d *Data) Init(msgs []int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, msg := range msgs {
		d.messages[msg] = struct{}{}
	}
}

func (d *Data) Write(msg int) {
	d.mu.Lock()
	d.messages[msg] = struct{}{}
	d.mu.Unlock()
}

func (d *Data) Has(msg int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if _, ok := d.messages[msg]; ok {
		return true
	}
	return false
}

func (d *Data) Read() []int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	data := make([]int, 0, len(d.messages))
	for msg := range d.messages {
		data = append(data, msg)
	}
	return data
}

type Topology struct {
	// graph is a directed graph. map of node mapping to set of neighbors
	graph map[string]map[string]struct{}
	mu    sync.RWMutex
}

// Build constructs the directed graph topology
func (t *Topology) Build(nodes map[string][]string) {
	// skip node if neighbor is already present and has recorded current node
	t.mu.Lock()
	for node, neighbors := range nodes {
		t.graph[node] = map[string]struct{}{}

		for _, neighbor := range neighbors {
			t.graph[node][neighbor] = struct{}{}
		}
	}
}

func main() {
	node := maelstrom.NewNode()
	data := &Data{messages: make(map[int]struct{})}

	// init handler: if existing nodes exists, we request for data from them on initialization
	node.Handle("init", func(req maelstrom.Message) error {
		var body struct {
			Type      string   `json:"type"`
			MessageID int      `json:"msg_id"`
			NodeID    string   `json:"node_id"`
			NodeIDs   []string `json:"node_ids"`
		}
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return err
		}

		// request for data from any random node. node should be alive to receive request
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			// prepare read request body
			readReqBody := map[string]any{
				"type": "read",
			}

			for _, neighbor := range body.NodeIDs {
				if neighbor == body.NodeID {
					continue
				}

				msg, err := node.SyncRPC(ctx, neighbor, readReqBody)
				// skip node on error
				if err != nil {
					continue
				}

				var reply struct {
					Type     string `json:"type"`
					Messages []int  `json:"messages"`
				}
				if err := json.Unmarshal(msg.Body, &reply); err != nil {
					continue
				}

				data.Init(reply.Messages)
				break
			}
		}()

		return nil
	})

	// broadcast rpc. we receive the message, store it and broadcast to neighbor nodes
	node.Handle("broadcast", func(req maelstrom.Message) error {
		var body struct {
			Type      string `json:"type"`
			MessageID *int   `json:"msg_id,omitempty"`
			Message   int    `json:"message"`
		}

		if err := json.Unmarshal(req.Body, &body); err != nil {
			return err
		}

		res := map[string]any{"type": "broadcast_ok"}

		// skip if message has been processed to avoid broadcast loop
		if data.Has(body.Message) {
			return node.Reply(req, res)
		}

		// write message and remove from body
		data.Write(body.Message)

		broadcastBody := map[string]any{
			"type":    "broadcast",
			"message": body.Message,
		}

		// TODO: use a gossip approach rather than broadcasting to all nodes in the cluster
		// response
		for _, neighbor := range node.NodeIDs() {
			if neighbor == node.ID() {
				continue
			}

			go Broadcast(node, neighbor, broadcastBody)
		}

		// inter-server messages sent with RPC call will typically have no message id so we skip response
		if body.MessageID == nil {
			return nil
		}

		return node.Reply(req, res)
	})

	// read rpc
	node.Handle("read", func(req maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return err
		}

		// reply with saved messages
		body["type"] = "read_ok"
		body["messages"] = data.Read()
		return node.Reply(req, body)
	})

	// topology rpc to get neighboring nodes
	node.Handle("topology", func(req maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return err
		}

		// remove data from response
		delete(body, "topology")
		body["type"] = "topology_ok"
		return node.Reply(req, body)
	})

	// run node
	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

// broadcast attempts to continuously send a message to the specified neighbor until its online.
// the timeout is 30 seconds with exponential backoff
func Broadcast(node *maelstrom.Node, neighbor string, body any) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	backOff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second
	isSuccess := false

	for {
		select {
		// stop after timeout
		case <-ctx.Done():
			return
		default:
			// an error can occur within the rpc when it is sent to the destination so a boolean flag is used as the indicator
			err := node.RPC(neighbor, body, func(msg maelstrom.Message) error {
				isSuccess = true
				return nil
			})
			if err == nil && isSuccess {
				return
			}

			// exponential backoff here
			time.Sleep(backOff)
			backOff = min(backOff*2, maxBackoff)
		}
	}
}
