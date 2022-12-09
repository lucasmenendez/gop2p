// Package gop2p implements simple peer-to-peer network node in pure Go.
package gop2p

import (
	"log"
	"net/http"
	"os"
	"sync"
)

// Node struct contains self peer reference and list of network members. Also
// contains reference to HTTP server node instance and all channels to
// goroutines communication and their sync structures.
type Node struct {
	Self   Peer
	Inbox  chan *Message
	Outbox chan *Message
	Leave  chan struct{}

	members    Peers
	membersMtx *sync.Mutex

	client *http.Client
	waiter sync.WaitGroup
	Logger *log.Logger
}

// InitNode function initializes a peer with current host information and
// creates required channels and structs. Then starts services (listeners and
// HTTP Server) and return node reference.
func NewNode(p int) (n *Node) {
	n = &Node{
		Self:   Me(p),
		Inbox:  make(chan *Message),
		Outbox: make(chan *Message),
		Leave:  make(chan struct{}),

		members:    Peers{},
		membersMtx: &sync.Mutex{},

		client: &http.Client{},
		waiter: sync.WaitGroup{},
		Logger: log.New(os.Stdout, "", 0),
	}

	n.start()
	return n
}

func (node *Node) start() {
	go func() {
		http.HandleFunc("/", node.handle())
		if err := http.ListenAndServe(node.Self.String(), nil); err != nil {
			log.Fatalln(err)
		}
	}()

	node.waiter.Add(1)
	go func() {
		for {
			select {
			// Broadcast Msg
			case msg := <-node.Outbox:
				go node.Broadcast(msg)
			case <-node.Leave:
				defer node.waiter.Done()
				node.Disconnect()
				close(node.Inbox)
				close(node.Outbox)
				return
			}
		}
	}()
}

func (n *Node) Wait() {
	defer n.waiter.Wait()
}
