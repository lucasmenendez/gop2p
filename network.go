// gop2p package implements simple Peer-to-Peer network node in pure Go.
package gop2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// baseUri contains node address template
const baseUri string = "http://%s:%d%s"

// joinPath contains node route where listen for node join.
const joinPath string = "/join"

// leavePath contains node route where listen for node leave.
const leavePath string = "/leave"

// broadcastPath contains node route where listen for messages.
const broadcastPath string = "/broadcast"

// contentType contains HTTP Header Content-Type default option.
const contentType string = "application/json"

// listener type involves http handler.
type listener func(http.ResponseWriter, *http.Request)

// listeners type involves map struct to store each listener with its route.
type listeners map[string]listener

// startListen function initializes HTTP Server node assigning to each route
// its listener function.
func (ls listeners) startListen(n *Node) {
	var s *http.ServeMux = http.NewServeMux()
	for r, l := range ls {
		s.HandleFunc(r, l)
	}

	var h string = fmt.Sprintf("%s:%d", n.Self.Address, n.Self.Port)
	n.server = &http.Server{Addr: h, Handler: s}
	go func() {
		if e := http.ListenAndServe(h, s); e != nil {
			n.Self.log("‼️ Error sending join: %s", e.Error())
			n.exit <- true
		}
	}()
}

// joinEmitter function make a POST HTTP request to external peer to connect
// with their network. If request is successfully, response contains a list of
// network nodes, function iterate over this nodes and communicate node join.
func joinEmitter(n *Node, p peer) {
	if j, e := json.Marshal(n.Self); e == nil {
		var u string = fmt.Sprintf(baseUri, p.Address, p.Port, joinPath)

		var r *http.Response
		var b *bytes.Buffer = bytes.NewBuffer(j)
		if r, e = http.Post(u, contentType, b); e != nil {
			n.Self.log("‼️ Error sending join: %s", e.Error())
			n.leave <- p
		}

		defer r.Body.Close()
		var ps peers = peers{}

		d := json.NewDecoder(r.Body)
		if e := d.Decode(&ps); e != nil {
			n.Self.log("‼️ Error decoding new peers: %s", e.Error())
			n.leave <- p
		}

		for _, p := range ps {
			n.join <- p
		}
	}
}

// leaveEmitter function send to all node members list a message that contains
// peer information to their leavePath of each node. Then, communicates to the
// eventLoop that peer leave has been broadcasting.
func leaveEmitter(n *Node) {
	if j, e := json.Marshal(n.Self); e == nil {
		for _, p := range n.Members {
			var u string = fmt.Sprintf(baseUri, p.Address, p.Port, leavePath)

			var r *http.Response
			var b *bytes.Buffer = bytes.NewBuffer(j)
			if r, e = http.Post(u, contentType, b); e != nil {
				n.Self.log("‼️ Error sending join: %s", e.Error())
				n.leave <- p
			}

			r.Body.Close()
		}
	}
}

// outboxEmitter function send message struct to all node members list to node
// broadcast path.
func outboxEmitter(n *Node, m msg) {
	for _, p := range n.Members {
		var u string = fmt.Sprintf(baseUri, p.Address, p.Port, broadcastPath)

		if d, e := json.Marshal(m); e != nil {
			n.Self.log("‼️ Error decoding message output: %s", e.Error())
		} else if _, e = http.Post(u, contentType, bytes.NewBuffer(d)); e != nil {
			n.Self.log("‼️ Error connecting: %s", e.Error())
			n.leave <- p
		}
	}

}

// joinListener function listens for new nodes in the network and communicate it
// to the eventLoop.
func joinListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var p peer = peer{}

			d := json.NewDecoder(r.Body)
			if e := d.Decode(&p); e != nil {
				n.Self.log("‼️ Error decoding peer joins: %s", e.Error())
				return
			}

			n.join <- p
			n.update <- true
			e := json.NewEncoder(w)
			e.Encode(<-n.sync)
		}
	}
}

// leaveListener function listens for peer network leaves and communicate it to
// the eventLoop.
func leaveListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var p peer = peer{}

			d := json.NewDecoder(r.Body)
			if e := d.Decode(&p); e != nil {
				n.Self.log("‼️ Error decoding peer leaves: %s", e.Error())
				return
			}

			n.leave <- p
		}
	}
}

// inbocListener function listens for new message broadcasted and communicate it
// to the eventLoop.
func inboxListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg msg = msg{}

		d := json.NewDecoder(r.Body)
		if e := d.Decode(&msg); e != nil {
			n.Self.log("‼️ Error decoding incoming message: %s", e.Error())
			return
		}

		n.inbox <- msg
	}
}
