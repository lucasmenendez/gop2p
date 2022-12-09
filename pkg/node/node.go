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

package node

import (
	"net/http"
	"sync"

	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

// Node struct
type Node struct {
	Self    *peer.Peer
	Members *peer.Members
	Inbox   chan *message.Message
	Outbox  chan *message.Message
	Connect chan *peer.Peer
	Leave   chan struct{}
	Error   chan error

	client *http.Client
	waiter sync.WaitGroup
}

// NewNode function
func New(self *peer.Peer) (n *Node) {
	n = &Node{
		Self:    self,
		Members: peer.EmptyMembers(),
		Inbox:   make(chan *message.Message),
		Outbox:  make(chan *message.Message),
		Connect: make(chan *peer.Peer),
		Leave:   make(chan struct{}),
		Error:   make(chan error),

		client: &http.Client{},
		waiter: sync.WaitGroup{},
	}

	return n
}

// init function
func (node *Node) Start() {
	go func() {
		http.HandleFunc("/", node.handle())
		if err := http.ListenAndServe(node.Self.String(), nil); err != nil {
			node.Error <- err
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

// Wait function
func (node *Node) Wait() {
	defer node.waiter.Wait()
}
