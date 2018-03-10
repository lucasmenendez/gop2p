package p2p

type handler func(t interface{})

type Node struct {
	Self    Peer
	Members peers

	inbox chan msg
	outbox chan msg
	join chan Peer
	leave chan Peer

	sync chan peers
	update chan bool

	callback handler
}

func New(p Peer) *Node {
	return &Node{
		p,
		peers{},
		make(chan msg),
		make(chan msg),
		make(chan Peer),
		make(chan Peer),
		make(chan peers),
		make(chan bool),
		nil,
	}
}

func (n *Node) SetCallback(c handler) {
	n.callback = c
}

func (n *Node) Init() {
	go n.eventLoop()
	go n.eventListeners()
}

func (n *Node) Join(p Peer) {
	n.join <- p
}

func (n *Node) eventLoop() {
	n.Self.log("Start event loop...")

	for {
		select {
		case m := <-n.inbox:
			n.handler(m)
		case m := <-n.outbox:
			outboxEmitter(m, n.Members)
		case p := <-n.join:
			if !n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = append(n.Members, p)
				n.Self.log(" 🔌 Connected to [%s](%s)", p.Address, p.Alias)
				joinEmitter(n, p)
			}
		case p := <-n.leave:
			n.Members = n.Members.delete(p)
		case <-n.update:
			n.sync <-n.Members
		}
	}
}

func (n *Node) eventListeners() {
	var ls listeners = listeners{
		broadcastPath: inboxListener(n),
		joinPath:      joinListener(n),
	}

	ls.startListen(n.Self.Address)
	n.Self.log("Start listeners...")
}

func (n *Node) Broadcast(p Peer, c string) {
	n.outbox <- msg{p, c}
	p.log("📨 Message has been sent: '%s'", c)
}

func (n *Node) handler(m msg) {
	if !n.Self.isMe(m.from) {
		if n.callback != nil {
			var info map[string]string = map[string]string {
				"Address": m.from.Address,
				"Alias": m.from.Alias,
				"content": m.content,
			}
			n.callback(info)
		} else {
			n.Self.log("\t📩 Message received from [%s](%s): '%s'",
				m.from.Address, m.from.Alias, m.content)
		}
	}
}