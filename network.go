package p2p

import (
	"net/http"
	"encoding/json"
	"fmt"
	"bytes"
)

const broadcastPath = "/broadcast"
const joinPath = "/join"
const contentType = "application/json"
const baseUri = "http://%s%s"
const defaultPort = "5555"

type listener func(http.ResponseWriter, *http.Request)
type listeners map[string]listener

func (ls listeners) startListen(a string) {
	for r, l := range ls {
		http.HandleFunc(r, l)
	}

	var h string = fmt.Sprintf("%s:%s", a, defaultPort)
	http.ListenAndServe(h, nil)
}

func joinEmitter(n *Node, p Peer) {
	if j, e := json.Marshal(n.Self); e != nil {
		var uri string = fmt.Sprintf(baseUri, p.Address, joinPath)

		var r *http.Response
		var b *bytes.Buffer = bytes.NewBuffer(j)
		if r, e = http.Post(uri, contentType, b); e != nil {
			n.Self.log("\t‼️ Error sending join")
			n.leave <- p
		}

		defer r.Body.Close()
		var ps peers = peers{}

		d := json.NewDecoder(r.Body)
		if e := d.Decode(&ps); e != nil {
			n.Self.log("\t‼️ Error decoding new peers")
			n.leave <- p
			return
		}

		for _, p := range ps {
			n.join <- p
		}
	}
}

func joinListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var p Peer = Peer{}

			d := json.NewDecoder(r.Body)
			if e := d.Decode(&p); e != nil {
				n.Self.log("\t‼️ Error decoding Peer joins")
			}

			n.join <- p
			n.update <- true
			e := json.NewEncoder(w)
			e.Encode(<- n.sync)
		}
	}
}

func outboxEmitter(m msg, ls peers) {}

func inboxListener(n *Node) listener {
	return func(http.ResponseWriter, *http.Request) {}
}