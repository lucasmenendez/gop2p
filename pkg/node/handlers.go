package node

import (
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/message"
)

// handle function manages every request received by the current network and
// performs the correct acttion to this requests. The function selects the
// correct handler based on the parsed message from the request. The connection
// message will come from "GET" requests, the disconnection message from
// "DELETE" requests and the plain message from "POST" requests.
func (node *Node) handle() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Omit OPTION requests.
		if r.Method == http.MethodOptions {
			return
		}

		// Set default headers and parse current request into a message.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		var msg = new(message.Message).FromRequest(r)

		// Select the handler based on the current request http.Method, GET
		// method is for connection requests, POST method for the plain message
		// request and  DELETE method for the disconnection requests. Otherwise
		// response with 405 HTTP Status Code.
		switch msg.Type {
		case message.ConnectType:
			var result = node.connected(msg)

			w.Header().Set("Content-Type", "text/plain")
			w.Write(result)
		case message.PlainType:
			// When plain message is received it will be redirected to the inbox
			// messages channel where the user will be waiting for read it.
			node.Inbox <- msg
		case message.DisconnectType:
			node.disconnected(msg)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 - Method not allowed!"))
		}
	}
}

// connected function handle a new connection to the network, appending the
// peer of the request to the current network members and response with that
// list encoding to JSON.
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

// disconnected function deletes the message peer from the current network
// members.
func (node *Node) disconnected(msg *message.Message) {
	node.Members.Delete(msg.From)
}
