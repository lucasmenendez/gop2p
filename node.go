// Package gop2p implements simple Peer-to-Peer network node in pure Go.
package gop2p

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type event struct {
	etype string
	data  map[string]interface{}
}

func (e event) toMap() map[string]interface{} {
	return map[string]interface{}{
		"type": e.etype,
		"data": e.data,
	}
}

// Handler type involves function to new messages handling.
type Handler func(d map[string]interface{})

// Node struct contains self Peer reference and list of network members. Also
// contains reference to HTTP server node instance and all channels to
// goroutines communication and their sync structures.
type Node struct {
	Self    Peer
	Members peers

	server *http.Server

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
func InitNode(a string, p int, d bool) (n *Node) {
	n = &Node{
		Self:    Me(a, p),
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
// a message. If node doesn'etype have associated Handler, incoming messages will be
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

// Leave function communicate to whole services (goroutines) that they must end.
func (n *Node) Leave() {
	n.exit <- true
}

// Broadcast function emmit message to the network passing received Content to
// broadcasting service.
func (n *Node) Broadcast(c string) {
	n.outbox <- Message{n.Self, c}
	n.log("💬 Message has been sent: '%s'", c)
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
	n.log("⌛️ Start event loop...")

	for {
		select {
		case m := <-n.inbox:
			if !n.Self.isMe(m.From) {
				n.handler(event{"inbox", m.toMap()})
				n.log("✉️ Message received From [%s:%d](%s): '%s'",
					m.From.Address, m.From.Port, m.From.Alias, m.Content)
			}

		case m := <-n.outbox:
			n.handler(event{"outbox", m.toMap()})
			if len(n.Members) > 0 {
				go outboxEmitter(n, m)
			} else {
				n.log("⚠️ Broadcasting aborted. Empty network!")
			}

		case p := <-n.join:
			if !n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = append(n.Members, p)

				n.handler(event{"join", p.toMap()})
				n.log("🔵 Connected to [%s:%d](%s)", p.Address, p.Port,
					p.Alias)

				go joinEmitter(n, p)
			}

		case p := <-n.leave:
			if n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = n.Members.delete(p)

				n.handler(event{"leave", p.toMap()})
				n.log("❌ Disconnected From [%s:%d](%s)", p.Address,
					p.Port, p.Alias)
			}

		case <-n.update:
			n.sync <- n.Members

		case <-n.exit:
			go leaveEmitter(n)

			n.handler(event{"disconnect", n.Self.toMap()})
			n.log("❌ Disconnecting...")

			n.server.Shutdown(nil)
			n.waiter.Done()
		}
	}
}

// eventListeners function defines listeners functions and each routes. Then
// starts HTTP Server goroutine.
func (n *Node) eventListeners() {
	var ls listeners = listeners{
		joinPath:      joinListener(n),
		leavePath:     leaveListener(n),
		broadcastPath: inboxListener(n),
	}

	n.log("⌛️ Start listeners...")
	n.log("👂 Listen at %s:%d", n.Self.Address, n.Self.Port)
	ls.startListen(n)
}

// handler function checks if node have a defined callback and execute it
// passing event information has map.
func (n *Node) handler(e event) {
	if n.callback != nil {
		n.callback(e.toMap())
	}
}

// log function logs message provided formated and adding seld Peer information
// trace.
func (n *Node) log(m string, args ...interface{}) {
	if n.debug {
		m = fmt.Sprintf(m, args...)
		log.Printf("[%s:%d](%s) - %s\n", n.Self.Address, n.Self.Port, n.Self.Alias, m)
	}
}
