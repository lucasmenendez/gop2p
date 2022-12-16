// node package contains the logic to keep listening to incoming messages while
// cocurrently allows to the user to perform connect, disconnect and broadcast
// actions.
// The package implements simple peer-to-peer network node in pure Go. Uses
// HTTP client and server to communicate over internet to knowed network
// members. gop2p implements the following functional workflow:
//  1. Connect to the network: The client gop2p.Node know a entry point of the
//     desired network (other gop2p.Node that is already connected). The entry
//     point response with the current network gop2p.Node's and updates its
//     members gop2p.Node list. The client gop2p.Node broadcast a connection
//     request to every gop2p.Node received from entry point.
//  2. Broadcasting a message: The client gop2p.Node prepares and broadcast a
//     gop2p.Message to every network gop2p.Node.
//  3. Disconnect from the network: The client gop2p.Node broadcast a
//     disconnection request to every network gop2p.Node. This gop2p.Node's
//     updates its current network members list unregistering the client
//     gop2p.Node.
package node

import (
	"net/http"
	"sync"

	"github.com/lucasmenendez/gop2p/message"
	"github.com/lucasmenendez/gop2p/peer"
)

// Node struct contains the information about the current peer associated to
// this node, the current network peers, some channels to errors, messages
// (send/receive) and network (connect/leave) managment. Also contains some
// hidden parameters such as the http.Client associated to the node or a
// WaitGroup to keep the node working.
type Node struct {
	Self    *peer.Peer    // information about current node
	Members *peer.Members // thread-safe list of peers on the network

	Inbox chan *message.Message // readable channels to receive messages
	Error chan error            // readable channels to receive errors

	Outbox  chan *message.Message // writtable channel to send messages
	Connect chan *peer.Peer       // writtable channel to connect to a Peer
	Leave   chan struct{}         // writtable channel to leave the network

	connected bool
	connMtx   *sync.Mutex

	waiter sync.WaitGroup
	client *http.Client
	server *http.Server
}

// NewNode function create a Node associated to the peer provided as argument.
func New(self *peer.Peer) (n *Node) {
	n = &Node{
		Self:    self,
		Members: peer.NewMembers(),
		Inbox:   make(chan *message.Message),
		Error:   make(chan error),
		Outbox:  make(chan *message.Message),
		Connect: make(chan *peer.Peer),
		Leave:   make(chan struct{}),

		connected: false,
		connMtx:   &sync.Mutex{},

		waiter: sync.WaitGroup{},
		client: &http.Client{},
		server: &http.Server{Addr: self.String()},
	}

	n.start()
	return
}

// Start function starts two goroutines, the first one to handle incoming
// requests and the second one to handle user actions listening to defined
// channels.
func (node *Node) start() {
	// Start HTTP server to listen to other network peers requests.
	go node.startListening()

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
				if node.IsConnected() {
					go node.broadcast(msg)
				} else {
					node.Error <- ConnErr("node not connected", nil, nil)
				}
			case <-node.Leave:
				if node.IsConnected() {
					node.disconnect()
				} else {
					node.Error <- ConnErr("node not connected", nil, nil)
				}

				// Reinitialize the channel when be closed
				node.Leave = make(chan struct{})
			}
		}
	}()
}

// IsConnected function returns the status of the current Node. It returns
// `true` if the node is already connected to a network, or `false` if it is
// not connected to any network yet.
func (node *Node) IsConnected() bool {
	node.connMtx.Lock()
	defer node.connMtx.Unlock()
	return node.connected
}

// Wait function waits until the WaitGroup of the current node has finished.
func (node *Node) Wait() {
	node.waiter.Wait()
}

// Stop function disconnect the node from the network, stop other goroutines
// and close the node channels.
func (node *Node) Stop() {
	// Stop goroutines
	node.waiter.Done()

	// If the node is connected, disconnect ir
	if node.IsConnected() {
		node.disconnect()
	}

	// Shutdown the HTTP server
	if err := node.server.Close(); err != nil {
		node.Error <- InternalErr("error shuting down the HTTP server", err, nil)
	}

	// Close other channels
	close(node.Inbox)
	close(node.Error)
	close(node.Outbox)
	close(node.Connect)
	close(node.Leave)
}
