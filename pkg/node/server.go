package node

import (
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/message"
)

// startListening function creates a HTTP request multiplexer to assing the root
// path to the Node.handleRequest function, assign it to the current Node.server
// and tries to start the HTTP server.
func (node *Node) startListening() {
	// Only listen on root and send every request to node handler.
	mux := http.NewServeMux()
	mux.HandleFunc("/", node.handleRequest())

	// Create the node HTTP server to listen to other peers requests.
	node.server.Handler = mux

	// If some error occurs it will be writted into Error channel and try to
	// disconnect.
	err := node.server.ListenAndServe()

	// If something was wrong, except the server is cosed, handle the error.
	if err != nil && err != http.ErrServerClosed {
		node.Error <- InternalErr("error listening for HTTP requests", err, nil)

		// If the current node was connected, update status to disconnected and
		// close the channel.
		if node.IsConnected() {
			node.setConnected(false)
		}
	}
}

// handleRequest function manages every request received by the current network
// and performs the correct acttion to this requests. The function selects the
// correct handler based on the parsed message from the request. The connection
// message will come from "GET" requests, the disconnection message from
// "DELETE" requests and the plain message from "POST" requests.
func (node *Node) handleRequest() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Omit OPTION requests.
		if r.Method == http.MethodOptions {
			return
		}

		// Set default headers and parse current request into a message.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		msg := new(message.Message).FromRequest(r)
		if msg == nil {
			http.Error(w, "No valid Message provided", http.StatusBadRequest)
			return
		}

		// Select the handler based on the current request http.Method, GET
		// method is for connection requests, POST method for the plain message
		// request and  DELETE method for the disconnection requests. Otherwise
		// response with 405 HTTP Status Code.
		switch msg.Type {
		case message.ConnectType:
			// Handle a new connection to the network, appending the peer of the
			// request to the current network members and response with that
			// list encoding to JSON.

			// Encode current list of members to a JSON to send it
			responseBody, err := node.Members.ToJSON()
			if err != nil {
				errMsg := "error encoding members to JSON"
				node.Error <- ParseErr(errMsg, err, msg)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}

			// Update the current member list safely appending the Message.From
			// Peer and if the current node was not connected update its status.
			node.Members.Append(msg.From)
			node.setConnected(true)

			// Send the current member list JSON to the connected peer
			w.Header().Set("Content-Type", "text/plain")
			w.Write(responseBody)
		case message.PlainType:
			// When plain message is received it will be redirected to the inbox
			// messages channel where the user will be waiting for read it.
			node.Inbox <- msg
		case message.DisconnectType:
			// disconnected function deletes the message peer from the current
			// network members.
			node.Members.Delete(msg.From)
			if node.Members.Len() == 0 {
				node.setConnected(false)
			}
		default:
			// By default response with 405 HTTP Status Code.
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
