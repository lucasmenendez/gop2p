// node package contains the logic to keep listening to incoming messages while
// cocurrently allows to the user to perform connect, disconnect and broadcast
// actions.
package node

import (
	"net/http"
	"sync"

	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

// Node struct contains the information about the current peer associated to
// this node, the current network peers, some channels to errors, messages
// (send/receive) and network (connect/leave) managment. Also contains some
// hidden parameters such as the http.Client associated to the node or a
// WaitGroup to keep the node working.
type Node struct {
	Self    *peer.Peer
	Members *peer.Members

	// Only write channel
	Inbox chan *message.Message
	Error chan error

	// Only read channels
	Outbox  chan *message.Message
	Connect chan *peer.Peer
	Leave   chan struct{}

	client *http.Client
	waiter sync.WaitGroup
}

// NewNode function create a Node associated to the peer provided as argument.
func New(self *peer.Peer) (n *Node) {
	return &Node{
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
}

// Start function starts two goroutines, the first one to handle incoming
// requests and the second one to handle user actions listening to defined
// channels.
func (node *Node) Start() {
	go func() {
		// Only listen on root and send every request to node handler. If some
		// error is registered it will be writted into Error channel.
		http.HandleFunc("/", node.handle())
		if err := http.ListenAndServe(node.Self.String(), nil); err != nil {
			node.Error <- err
		}
	}()

	// Increase the counter of the current node WaitGroup to wait for the
	// following goroutine.
	node.waiter.Add(1)
	go func() {
		// Run forever unless the Leave channel will be close, handling user
		// actions such as connect to knowed peer or broadcast a message.
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

// Wait function waits until the WaitGroup of the current node has finished.
func (node *Node) Wait() {
	node.waiter.Wait()
}
