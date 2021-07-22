// Package gop2p implements simple peer-to-peer network node in pure Go.
package gop2p

import "sync"

// Node struct contains self peer reference and list of network members. Also
// contains reference to HTTP server node instance and all channels to
// goroutines communication and their sync structures.
type Node struct {
	Self    Peer
	Members Peers

	network *network

	inbox      chan message
	outbox     chan message
	connect    chan Peer
	disconnect chan bool
	join       chan Peer
	leave      chan Peer

	waiter *sync.WaitGroup
	events *eventBus

	debug     bool
	connected bool
}

// InitNode function initializes a peer with current host information and
// creates required channels and structs. Then starts services (listeners and
// HTTP Server) and return node reference.
func InitNode(p int, d bool) (n *Node) {
	n = &Node{
		Self:       Me(p),
		Members:    Peers{},
		inbox:      make(chan message),
		outbox:     make(chan message),
		connect:    make(chan Peer),
		join:       make(chan Peer),
		disconnect: make(chan bool),
		leave:      make(chan Peer),
		waiter:     &sync.WaitGroup{},
		events:     newEventBus(),
		debug:      d,
		connected:  true,
	}

	n.startService()
	return
}

// Wait function keeps node alive.
func (n *Node) Wait() {
	defer n.waiter.Wait()

	for {
		select {}
	}
}

// Connect function allows node to connect to a network via entry peer
// reference, that contains its information.
func (n *Node) Connect(p Peer) {
	n.connect <- p
}

// Disconnect function communicates to whole services (goroutines) that they
// must end.
func (n *Node) Disconnect() {
	n.disconnect <- true
}

// Broadcast function emmits a message to the network passing received Content
// to broadcasting service.
func (n *Node) Broadcast(m []byte) {
	n.outbox <- message{
		data: m,
		from: n.Self,
	}
}

// OnMessage function allows to create a message handler to listen for data from
// other peers.
func (n *Node) OnMessage(f func([]byte, Peer)) {
	n.events.on("message", f)
}

// OnConnection function allows to assign a connection handler, fired when a peer
// join to the network
func (n *Node) OnConnection(f func(Peer)) {
	n.events.on("connection", func(_ []byte, peer Peer) {
		f(peer)
	})
}

// OnDisconnection function allows to assign a disconnection handler, fired when
// a peer leaves the network.
func (n *Node) OnDisconnection(f func(Peer)) {
	n.events.on("disconnection", func(_ []byte, peer Peer) {
		f(peer)
	})
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
			n.events.emit("message", m.data, m.from)
			n.log("Message received: %s", m.String())

		case m := <-n.outbox:
			if len(n.Members) > 0 {
				go n.network.messageEmitter(m.data)
				n.log("Message sended: %s", m.String())
			} else {
				n.log("Broadcasting aborted. Empty network!")
			}

		case p := <-n.connect:
			n.log("Connecting to [%s:%s]", p.Address, p.Port)
			go n.network.connectEmitter(p)

		case p := <-n.join:
			if !n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = append(n.Members, p)
				n.events.emit("connection", nil, p)
				n.log("Connected to [%s:%s]", p.Address, p.Port)
			}

		case <-n.disconnect:
			if n.connected {
				n.log("Disconnecting...")
				n.network.disconnectEmitter()
				n.Members = Peers{}
				n.connected = false
				n.waiter.Done()
			}

		case p := <-n.leave:
			if n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = n.Members.delete(p)
				n.events.emit("disconnection", nil, p)
				n.log("Disconnected from [%s:%s]", p.Address, p.Port)
			}
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
