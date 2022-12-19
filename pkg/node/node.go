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
	"context"
	"net/http"
	"sync"

	"github.com/lucasmenendez/gop2p/v2/pkg/message"
	"github.com/lucasmenendez/gop2p/v2/pkg/peer"
)

// Node struct contains the information about the current peer associated to
// this node, the current network peers, some channels to errors, messages
// (send/receive) and network (connect/leave) management. Also contains some
// hidden parameters such as the http.Client associated to the node or a
// WaitGroup to keep the node working.
type Node struct {
	Self    *peer.Peer    // information about current node
	Members *peer.Members // thread-safe list of peers on the network

	Inbox chan *message.Message // readable channels to receive messages
	Error chan *NodeErr         // readable channels to receive errors

	Connection chan *peer.Peer       // writtable channel to connect to a Peer
	Outbox     chan *message.Message // writtable channel to send messages

	connected bool
	connMtx   *sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	client *http.Client
	server *http.Server
	waiter *sync.WaitGroup
}

// New function create a Node associated to the peer provided as argument.
func New(self *peer.Peer) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		Self:       self,
		Members:    peer.NewMembers(),
		Connection: make(chan *peer.Peer),
		Inbox:      make(chan *message.Message),
		Outbox:     make(chan *message.Message),
		Error:      make(chan *NodeErr),

		connected: false,
		connMtx:   &sync.Mutex{},

		ctx:    ctx,
		cancel: cancel,
		server: nil, // Initialize as nil to know if the the node is started
		client: &http.Client{},
		waiter: &sync.WaitGroup{},
	}
}

// Start function starts two goroutines, the first one to handle incoming
// requests and the second one to handle user actions listening to defined
// channels.
func (n *Node) Start() {
	// Initialize the current node server
	n.server = &http.Server{Addr: n.Self.String()}

	// Start HTTP server to listen to other network peers requests.
	go n.startListening()

	// Increase the counter of the current node WaitGroup to wait for the
	// following goroutine.
	n.waiter.Add(1)
	go func() {
		defer n.waiter.Done()
		// For loop handling the node chanlles looking for new connection,
		// disconection or send message requests, until the context will be
		// canceled.
		for {
			select {
			case p, connect := <-n.Connection:
				if connect {
					// If the channel is still opened, try to connect to the peer
					// provided.
					if err := n.connect(p); err != nil {
						n.Error <- err
					}
				} else {
					// But if it was closed, disconnect from the network and
					// reinitialize the channel.
					if err := n.disconnect(); err != nil {
						n.Error <- err
					}
					n.Connection = make(chan *peer.Peer)
				}

			case msg := <-n.Outbox:
				// If the channel receives a message, broadcast to the network
				if err := n.broadcast(msg); err != nil {
					n.Error <- err
				}
			case <-n.ctx.Done():
				// If the context is cancelled exit from the loop
				return
			}
		}
	}()
}

// IsConnected function returns the status of the current Node. It returns
// `true` if the node is already connected to a network, or `false` if it is
// not connected to any network yet.
func (n *Node) IsConnected() bool {
	n.connMtx.Lock()
	defer n.connMtx.Unlock()
	return n.connected
}

// Stop function disconnect the node from the network, stop other goroutines
// and close the node channels.
func (n *Node) Stop() error {
	// If the current node is not started return error
	if n.server == nil {
		return InternalErr("current node not started", nil)
	}

	// If the node is connected, disconnect from the network
	if n.IsConnected() {
		if err := n.disconnect(); err != nil {
			return err
		}
	}

	// Shutdown the HTTP server
	if err := n.server.Shutdown(n.ctx); err != nil {
		return InternalErr("error shutting down the HTTP server", err)
	}

	// Stop channels for-loop and close the channels
	n.cancel()
	n.waiter.Wait()

	safeClose(n.Inbox)
	safeClose(n.Outbox)
	safeClose(n.Connection)
	safeClose(n.Error)
	n.server = nil
	return nil
}

func safeClose[C *message.Message | *peer.Peer | *NodeErr](ch chan C) {
	select {
	case <-ch:
		close(ch)
		return
	default:
	}
}
