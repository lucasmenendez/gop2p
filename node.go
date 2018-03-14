package gop2p

import (
	"net/http"
	"sync"
)

type Handler func(t interface{})

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

func (n *Node) SetCallback(c Handler) {
	n.callback = c
}

func (n *Node) Wait() {
	n.waiter.Wait()
}

func (n *Node) Connect(p peer) {
	n.join <- p
}

func (n *Node) Leave() {
	n.exit <- true
}

func (n *Node) Broadcast(c string) {
	n.outbox <- msg{n.Self, c}
	n.Self.log("💬 Message has been sent: '%s'", c)
}

func (n *Node) startService() {
	n.waiter.Add(1)
	go n.eventLoop()
	go n.eventListeners()
}

func (n *Node) eventLoop() {
	n.Self.log("⌛️ Start event loop...")

	for {
		select {
		case m := <-n.inbox:
			if !n.Self.isMe(m.From) {
				n.handler(m)
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
				n.Self.log("🔵 Connected to [%s:%d](%s)", p.Address, p.Port, p.Alias)
				go joinEmitter(n, p)
			}

		case p := <-n.leave:
			if n.Members.contains(p) && !n.Self.isMe(p) {
				n.Members = n.Members.delete(p)
				n.Self.log("❌ Disconnected from [%s:%d](%s)", p.Address, p.Port, p.Alias)
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

func (n *Node) handler(m msg) {
	if !n.Self.isMe(m.From) {
		if n.callback != nil {
			var info map[string]interface{} = map[string]interface{}{
				"from": map[string]interface{}{
					"adress": m.From.Address,
					"alias":  m.From.Alias,
					"port":   m.From.Port,
				},
				"content": m.Content,
			}
			n.callback(info)
		} else {
			n.Self.log("✉️ Message received From [%s:%d](%s): '%s'",
				m.From.Address, m.From.Port, m.From.Alias, m.Content)
		}
	}
}
