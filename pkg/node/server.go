package node

import (
	"io"
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

// startListening function creates a HTTP request multiplexer to assing the root
// path to the Node.handleRequest function, assign it to the current Node.server
// and tries to start the HTTP server.
func (n *Node) startListening() {
	mux := http.NewServeMux()
	// Listen on root every request and handle it with the default node handler.
	mux.HandleFunc("/", n.handleRequest())

	// Create the node HTTP server to listen to other peers requests.
	n.server.Handler = mux

	// If some error occurs it will be writted into Error channel and try to
	// disconnect.
	err := n.server.ListenAndServe()

	// If something was wrong, except the server is cosed, handle the error.
	if err != nil {
		// If the current node was connected, update status to disconnected and
		// close the channel.
		if n.IsConnected() {
			n.Members = peer.NewMembers()
			n.setConnected(false)
		}

		n.Error <- InternalErr("error listening for HTTP requests", err)
	}
}

// handleRequest function manages every request received by the current network
// and performs the correct acttion to this requests. The function selects the
// correct handler based on the parsed message from the request. The connection
// message will come from "GET" requests, the disconnection message from
// "DELETE" requests and the plain message from "POST" requests.
func (n *Node) handleRequest() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == n.Self.String() {
			http.Error(w, "You can not connect with yourself.", http.StatusBadRequest)
			return
		}

		// Set cors compatible headers when the request has OPTION method.
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		// Parse request to a message
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "no valid message provided", http.StatusBadRequest)
			return
		}

		msg := new(message.Message).SetJSON(data)
		if msg == nil {
			// If something fails decoding message from the request, response
			// with a bad request HTTP error.
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
			responseBody, err := n.Members.ToJSON()
			if err != nil {
				errMsg := "error encoding members to JSON"
				n.Error <- ParseErr(errMsg, err)
				http.Error(w, errMsg, http.StatusInternalServerError)
				return
			}

			// Update the current member list safely appending the Message.From
			// Peer and if the current node was not connected update its status.
			n.Members.Append(msg.From)
			n.setConnected(true)

			// Send the current member list JSON to the connected peer
			w.Header().Set("Content-Type", "text/plain")
			w.Write(responseBody)
		case message.BroadcastType, message.DirectType:
			if !n.Members.Contains(msg.From) {
				// If the message peer is not a registered member of the current
				// network, return a forbidden HTTP error.
				http.Error(w, "Peer not registered", http.StatusForbidden)
				return
			}
			// When broadcast or direct message is received it will be redirected
			// to the inbox messages channel where the user will be waiting for
			// read it.
			n.Inbox <- msg
		case message.DisconnectType:
			if !n.Members.Contains(msg.From) {
				// If the message peer is not a registered member of the current
				// network, return a forbidden HTTP error.
				http.Error(w, "Peer not registered", http.StatusForbidden)
				return
			}

			// disconnected function deletes the message peer from the current
			// network members.
			n.Members.Delete(msg.From)
			if n.Members.Len() == 0 {
				n.setConnected(false)
			}
		default:
			// By default response with 405 HTTP Status Code.
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
