// Package gop2p implements simple peer-to-peer network node in pure Go.
package gop2p

import (
	"fmt"
	"log"
	"sync"
)

// Handler type involves function to new messages handling.
type Handler func(m []byte)

// Node struct contains self peer reference and list of network members. Also
// contains reference to HTTP server node instance and all channels to
// goroutines communication and their sync structures.
type Node struct {
	Self    peer
	Members peers

	network *network

	inbox      chan []byte
	outbox     chan []byte
	connect    chan peer
	join       chan peer
	disconnect chan bool
	leave      chan peer

	waiter *sync.WaitGroup

	callback Handler
	debug    bool
}

// InitNode function initializes a peer with current host information and
// creates required channels and structs. Then starts services (listeners and
// HTTP Server) and return node reference.
func InitNode(p int, d bool) (n *Node) {
	n = &Node{
		Self:       Me(p),
		Members:    peers{},
		inbox:      make(chan []byte),
		outbox:     make(chan []byte),
		connect:    make(chan peer),
		join:       make(chan peer),
		disconnect: make(chan bool),
		leave:      make(chan peer),
		waiter:     &sync.WaitGroup{},
		debug:      d,
	}

	n.startService()
	return
}

// OnMessage function receives a Handler function to call when node receives
// a message. If node doesn't have associated Handler, incoming messages will be
// logged with standard library.
func (n *Node) OnMessage(c Handler) {
	n.callback = c
}

// Wait function keeps node alive.
func (n *Node) Wait() {
	n.waiter.Wait()
}

// Connect function allows node to connect to a network via entry peer
// reference, that contains its information.
func (n *Node) Connect(p peer) {
	n.connect <- p
}

// Disconnect function communicate to whole services (goroutines) that they
// must end.
func (n *Node) Disconnect() {
	n.disconnect <- true
}

// Broadcast function emmit message to the network passing received Content to
// broadcasting service.
func (n *Node) Broadcast(m []byte) {
	n.outbox <- m
}

// startService function adds 1 to node WaitGroup and starts eventLoop and
// eventListener goroutines.
func (n *Node) startService() {
	n.waiter.Add(1)
	go n.eventLoop()
	go n.eventListeners()
}

// eventLoop function contains main loop. Into main loop, each channel is
// checked and executes the corresponding functions.
func (n *Node) eventLoop() {
	n.log("Start event loop...")

	for {
		select {
		case m := <-n.inbox:
			if n.callback != nil {
				n.callback(m)
			}
			n.log("Message received: '%s'", m)

		case m := <-n.outbox:
			if len(n.Members) > 0 {
				go n.network.messageEmitter(m)
				n.log("Message sended: '%s'", m)
			} else {
				n.log("Broadcasting aborted. Empty network!")
			}

		case p := <-n.connect:
			n.log("Connecting to [%s:%s]", p.address, p.port)
			go n.network.connectEmitter(p)

		case p := <-n.join:
			if !n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = append(n.Members, p)
				n.log("Connected to [%s:%s]", p.address, p.port)
			}

		case <-n.disconnect:
			n.log("Disconnecting...")
			n.network.disconnectEmitter()
			n.waiter.Done()

		case p := <-n.leave:
			if n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = n.Members.delete(p)
				n.log("Disconnected From [%s:%s]", p.address, p.port)
			}
		}
	}
}

// eventListeners function defines listeners functions and each routes. Then
// starts HTTP Server goroutine.
func (n *Node) eventListeners() {
	n.network = newNetwork(n)
	n.log("Listen at %s:%s", n.Self.address, n.Self.port)
	n.network.start()
}

// log function logs message provided formated and adding seld peer information
// trace.
func (n *Node) log(m string, args ...interface{}) {
	if n.debug {
		m = fmt.Sprintf(m, args...)
		log.Printf("[%s:%s] %s\n", n.Self.address, n.Self.port, m)
	}
}
