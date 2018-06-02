// Package gop2p implements simple Peer-to-Peer network node in pure Go.
package gop2p

import (
	"fmt"
	"log"
	"sync"
)

// Handler type involves function to new messages handling.
type Handler func(m []byte)

// Node struct contains self Peer reference and list of network members. Also
// contains reference to HTTP server node instance and all channels to
// goroutines communication and their sync structures.
type Node struct {
	Self    Peer
	Members peers

	network *network

	inbox  chan Message
	outbox chan Message
	join   chan Peer
	leave  chan Peer

	sync   chan peers
	update chan bool

	exit   chan bool
	waiter *sync.WaitGroup

	callback Handler
	debug    bool
}

// InitNode function initializes a Peer with current host information and
// creates required channels and structs. Then starts services (listeners and
// HTTP Server) and return node reference.
func InitNode(p int, d bool) (n *Node) {
	n = &Node{
		Self:    Me(p),
		Members: peers{},
		inbox:   make(chan Message),
		outbox:  make(chan Message),
		join:    make(chan Peer),
		leave:   make(chan Peer),
		sync:    make(chan peers),
		update:  make(chan bool),
		exit:    make(chan bool),
		waiter:  &sync.WaitGroup{},
		debug:   d,
	}

	n.startService()
	return
}

// SetCallback function receives a Handler function to call when node receives
// a message. If node doesn't have associated Handler, incoming messages will be
// logged with standard library.
func (n *Node) SetCallback(c Handler) {
	n.callback = c
}

// Wait function keeps node alive.
func (n *Node) Wait() {
	n.waiter.Wait()
}

// Join function allows node to connect to a network via entry Peer
// reference, that contains its information.
func (n *Node) Join(p Peer) {
	n.join <- p
}

func (n *Node)

// Leave function communicate to whole services (goroutines) that they must end.
func (n *Node) Leave() {
	n.exit <- true
}

// Broadcast function emmit message to the network passing received Content to
// broadcasting service.
func (n *Node) Broadcast(c []byte) {
	n.outbox <- Message{n.Self, c}
	n.log("Message has been sent: '%s'", c)
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
			if !n.Self.isMe(m.From) {
				n.handler(m)
				n.log("Message received From [%s:%s]: '%s'",
					m.From.Address, m.From.Port, m.Content)
			}

		case _ = <-n.outbox:
			if len(n.Members) > 0 {
				//go n.network.outboxEmitter(m)
			} else {
				n.log("Broadcasting aborted. Empty network!")
			}

		case p := <-n.join:
			if !n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = append(n.Members, p)

				n.log("Connected to [%s:%s]", p.Address, p.Port)
				//go joinEmitter(n, p)
			}

		case p := <-n.leave:
			if n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = n.Members.delete(p)
				n.log("Disconnected From [%s:%s]", p.Address, p.Port)
			}

		case <-n.update:
			n.sync <- n.Members

		case <-n.exit:
			//go leaveEmitter(n)

			n.log("Disconnecting...")
			n.waiter.Done()
		}
	}
}

// eventListeners function defines listeners functions and each routes. Then
// starts HTTP Server goroutine.
func (n *Node) eventListeners() {
	n.network = newNetwork(n)
	n.log("Listen at %s:%s", n.Self.Address, n.Self.Port)
	n.network.start()
}

// handler function checks if node have a defined callback and execute it
// passing event information has map.
func (n *Node) handler(m Message) {
	if n.callback != nil {
		n.callback(m.Content)
	}
}

// log function logs message provided formated and adding seld Peer information
// trace.
func (n *Node) log(m string, args ...interface{}) {
	if n.debug {
		m = fmt.Sprintf(m, args...)
		log.Printf("[%s:%s] %s\n", n.Self.Address, n.Self.Port, m)
	}
}
