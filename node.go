// gop2p package implements simple Peer-to-Peer network node in pure Go.
package gop2p

import (
	"net/http"
	"sync"
)

// Handler type involves function to new messages handling.
type Handler func(t interface{})

// Node struct contains self peer reference and list of network members. Also
// contains reference to HTTP server node instance and all channels to
// goroutines communication and their sync structures.
type Node struct {
	Self    peer
	Members peers

	server *http.Server

	inbox  chan msg
	outbox chan msg
	join   chan peer
	leave  chan peer

	sync   chan peers
	update chan bool

	exit   chan bool
	waiter *sync.WaitGroup

	callback Handler
}

// InitNode function initializes a peer with current host information and
// creates required channels and structs. Then starts services (listeners and
// HTTP Server) and return node reference.
func InitNode(a string, p int) (n *Node) {
	n = &Node{
		Self:    Me(a, p),
		Members: peers{},
		inbox:   make(chan msg),
		outbox:  make(chan msg),
		join:    make(chan peer),
		leave:   make(chan peer),
		sync:    make(chan peers),
		update:  make(chan bool),
		exit:    make(chan bool),
		waiter:  &sync.WaitGroup{},
	}

	n.startService()
	return n
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

// Connect function allows node to connect to a network via entry peer
// reference, that contains its information.
func (n *Node) Connect(p peer) {
	n.join <- p
}

// Leave function communicate to whole services (goroutines) that they must end.
func (n *Node) Leave() {
	n.exit <- true
}

// Broadcast function emmit message to the network passing received content to
// broadcasting service.
func (n *Node) Broadcast(c string) {
	n.outbox <- msg{n.Self, c}
	n.Self.log("💬 Message has been sent: '%s'", c)
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
	n.Self.log("⌛️ Start event loop...")

	for {
		select {
		case m := <-n.inbox:
			if !n.Self.isMe(m.From) {
				if n.callback != nil {
					var info map[string]interface{} = m.toMap()
					n.callback(info)
				} else {
					n.Self.log("✉️ Message received From [%s:%d](%s): '%s'",
						m.From.Address, m.From.Port, m.From.Alias, m.Content)
				}
			}

		case m := <-n.outbox:
			if len(n.Members) > 0 {
				go outboxEmitter(n, m)
			} else {
				n.Self.log("⚠️ Broadcasting aborted. Empty network!")
			}

		case p := <-n.join:
			if !n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = append(n.Members, p)
				n.Self.log("🔵 Connected to [%s:%d](%s)", p.Address, p.Port,
					p.Alias)
				go joinEmitter(n, p)
			}

		case p := <-n.leave:
			if n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = n.Members.delete(p)
				n.Self.log("❌ Disconnected from [%s:%d](%s)", p.Address,
					p.Port, p.Alias)
			}

		case <-n.update:
			n.sync <- n.Members

		case <-n.exit:
			go leaveEmitter(n)
			n.Self.log("❌ Disconnecting...")
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

	n.Self.log("⌛️ Start listeners...")
	n.Self.log("👂 Listen at %s:%d", n.Self.Address, n.Self.Port)
	go ls.startListen(n)
}