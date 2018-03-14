package gop2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const baseUri string = "http://%s:%d%s"
const joinPath string = "/join"
const leavePath string = "/leave"
const broadcastPath string = "/broadcast"
const contentType string = "application/json"

type listener func(http.ResponseWriter, *http.Request)
type listeners map[string]listener

func (ls listeners) startListen(n *Node) {
	var s *http.ServeMux = http.NewServeMux()
	for r, l := range ls {
		s.HandleFunc(r, l)
	}

	var h string = fmt.Sprintf("%s:%d", n.Self.Address, n.Self.Port)
	n.server  = &http.Server{ Addr: h, Handler: s }
	go func() {
		if e := http.ListenAndServe(h, s); e != nil {
			n.Self.log("‼️ Error sending join: %s", e.Error())
			n.exit <- true
		}
	}()
}

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
