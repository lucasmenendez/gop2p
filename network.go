package gop2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const baseUri = "http://%s:%d%s"
const broadcastPath = "/broadcast"
const joinPath = "/join"
const contentType = "application/json"

type listener func(http.ResponseWriter, *http.Request)
type listeners map[string]listener

func (ls listeners) startListen(a string, p int) {
	var s *http.ServeMux = http.NewServeMux()
	for r, l := range ls {
		s.HandleFunc(r, l)
	}

	var h string = fmt.Sprintf("%s:%d", a, p)
	http.ListenAndServe(h, s)
}

func joinEmitter(n *Node, p peer) {
	if j, e := json.Marshal(n.Self); e == nil {
		var uri string = fmt.Sprintf(baseUri, p.Address, p.Port, joinPath)

		var r *http.Response
		var b *bytes.Buffer = bytes.NewBuffer(j)
		if r, e = http.Post(uri, contentType, b); e != nil {
			n.Self.log("\t‼️ Error sending join: %s", e.Error())
			n.leave <- p
		}

		defer r.Body.Close()
		var ps peers = peers{}

		d := json.NewDecoder(r.Body)
		if e := d.Decode(&ps); e != nil {
			n.Self.log("\t‼️ Error decoding new peers: %s", e.Error())
			n.leave <- p
		}

		for _, p := range ps {
			n.join <- p
		}
	}
}

func joinListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var p peer = peer{}

			d := json.NewDecoder(r.Body)
			if e := d.Decode(&p); e != nil {
				n.Self.log("\t‼️ Error decoding peer joins: %s", e.Error())
				return
			}

			n.join <- p
			n.update <- true
			e := json.NewEncoder(w)
			e.Encode(<-n.sync)
		}
	}
}

func outboxEmitter(m msg, ls peers) {
	for _, p := range ls {
		var u string = fmt.Sprintf(baseUri, p.Address, p.Port, broadcastPath)

		// TODO: Do something with marshal and network errors
		// ¿We can supose that peer is disconnected and alert peer left?
		if d, e := json.Marshal(m); e != nil {
			fmt.Println(e)
		} else if _, e = http.Post(u, contentType, bytes.NewBuffer(d)); e != nil {
			fmt.Println(e)
		}
	}

}

func inboxListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg msg = msg{}

		d := json.NewDecoder(r.Body)
		if e := d.Decode(&msg); e != nil {
			n.Self.log("\t‼️ Error decoding incoming message: %s", e.Error())
			return
		}

		n.inbox <- msg
	}
}
