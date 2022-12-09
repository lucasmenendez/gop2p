package node

import (
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/message"
)

// handle function
func (node *Node) handle() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		var msg = new(message.Message).FromRequest(r)

		if r.Method == http.MethodGet {
			var result = node.connected(msg)

			w.Header().Set("Content-Type", "text/plain")
			w.Write(result)
		} else if r.Method == http.MethodPost {
			node.Inbox <- msg
		} else if r.Method == http.MethodDelete {
			node.disconnected(msg)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 - Method not allowed!"))
		}
	}
}

// connected function
func (node *Node) connected(msg *message.Message) (members []byte) {
	// Encoding current list of members to a JSON to send it
	var err error
	if members, err = node.Members.ToJSON(); err != nil {
		// TODO: handle error
		node.Error <- ParseErr("error encoding members to JSON", err, msg)
	}

	// Update the current member list safely appending the Message.From Peer
	node.Members.Append(msg.From)
	return
}

// disconnected function
func (node *Node) disconnected(msg *message.Message) {
	// Delete the Message.From Peer from the current member list safely
	node.Members.Delete(msg.From)
}
