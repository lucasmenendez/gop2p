package gop2p

type handler func(t interface{})

type Node struct {
	Self    Peer
	Members peers

	inbox  chan Msg
	outbox chan Msg
	join   chan Peer
	leave  chan Peer

	sync   chan peers
	update chan bool

	callback handler
}

func InitNode(a string, p int) (n *Node) {
	n = &Node{
		Me(a, p),
		peers{},
		make(chan Msg),
		make(chan Msg),
		make(chan Peer),
		make(chan Peer),
		make(chan peers),
		make(chan bool),
		nil,
	}

	n.startService()
	return n
}

func (n *Node) SetCallback(c handler) {
	n.callback = c
}

func (n *Node) startService() {
	go n.eventLoop()
	go n.eventListeners()
}

func (n *Node) Join(p *Node) {
	n.join <- p.Self
}

func (n *Node) eventLoop() {
	n.Self.log("Start event loop...")

	for {
		select {
		case m := <-n.inbox:
			if !n.Self.isMe(m.From) {
				go n.handler(m)
			}

		case m := <-n.outbox:
			go outboxEmitter(m, n.Members)

		case p := <-n.join:
			if !n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = append(n.Members, p)
				n.Self.log(" 🔌 Connected to [%s:%d](%s)", p.Address, p.Port, p.Alias)
				go joinEmitter(n, p)
			}

		case p := <-n.leave:
			n.Members = n.Members.delete(p)

		case <-n.update:
			n.sync <- n.Members
		}
	}
}

func (n *Node) eventListeners() {
	var ls listeners = listeners{
		broadcastPath: inboxListener(n),
		joinPath:      joinListener(n),
	}

	n.Self.log("Start listeners...")
	ls.startListen(n.Self.Address, n.Self.Port)
	n.Self.log("Listen at %s:%d", n.Self.Address, n.Self.Port)
}

func (n *Node) Broadcast(c string) {
	n.outbox <- Msg{n.Self, c}
	n.Self.log("📨 Message has been sent: '%s'", c)
}

func (n *Node) handler(m Msg) {
	if !n.Self.isMe(m.From) {
		if n.callback != nil {
			var info map[string]string = map[string]string{
				"Address": m.From.Address,
				"Alias":   m.From.Alias,
				"Content": m.Content,
			}
			n.callback(info)
		} else {
			n.Self.log("\t📩 Message received From [%s:%d](%s): '%s'",
				m.From.Address, m.From.Port, m.From.Alias, m.Content)
		}
	}
}
