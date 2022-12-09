// Package gop2p implements simple peer-to-peer network node in pure Go. Uses
// HTTP client and server to communicate over internet to knowed network
// members. gop2p implements the following functional workflow:
// 	1. Connect to the network: The client gop2p.Node know a entry point of the
// 		desired network (other gop2p.Node that is already connected). The entry
//		point response with the current network gop2p.Node's and updates its
//		members gop2p.Node list. The client gop2p.Node broadcast a connection
//		request to every gop2p.Node received from entry point.
// 	2. Broadcasting a message: The client gop2p.Node prepares and broadcast a
//		gop2p.Message to every network gop2p.Node.
// 	3. Disconnect from the network: The client gop2p.Node broadcast a
//		disconnection request to every network gop2p.Node. This gop2p.Node's
//		updates its current network members list unregistering the client
//		gop2p.Node.

package gop2p

import (
	"log"
	"net/http"
	"sync"
)

// Node struct
type Node struct {
	Self    *Peer
	Inbox   chan *Message
	Outbox  chan *Message
	Connect chan *Peer
	Leave   chan struct{}
	Error   chan error

	members    *Peers
	membersMtx *sync.Mutex

	client *http.Client
	waiter sync.WaitGroup
}

// NewNode function
func NewNode(p int) (n *Node) {
	n = &Node{
		Self:    Me(p),
		Inbox:   make(chan *Message),
		Outbox:  make(chan *Message),
		Connect: make(chan *Peer),
		Leave:   make(chan struct{}),
		Error:   make(chan error),

		members:    &Peers{},
		membersMtx: &sync.Mutex{},

		client: &http.Client{},
		waiter: sync.WaitGroup{},
	}

	n.init()
	return n
}

// Wait function
func (node *Node) Wait() {
	defer node.waiter.Wait()
}

// init function
func (node *Node) init() {
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
			case peer := <-node.Connect:
				node.connect(peer)
			case msg := <-node.Outbox:
				go node.broadcast(msg)
			case <-node.Leave:
				node.disconnect()
				return
			}
		}
	}()
}
