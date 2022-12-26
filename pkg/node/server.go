package node

import (
	"fmt"
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

// startListening function creates a HTTP request multiplexer to assing the root
// path to the Node.handleRequest function, assign it to the current Node.server
// and tries to start the HTTP server.
func (n *Node) startListening() {
	// Only listen on root and send every request to node handler.
	mux := http.NewServeMux()
	mux.HandleFunc("/", n.handleRequest())
	mux.HandleFunc("/sse", n.handelSSE())

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
			w.WriteHeader(http.StatusOK)
			return
		}
		// Set default headers and parse current request into a message.
		w.Header().Set("Access-Control-Allow-Origin", "*")

		msg := new(message.Message).FromRequest(r)
		if msg == nil {
			// If something fails decoding message from the request, response
			// with a bad request HTTP error.
			http.Error(w, "No valid Message provided", http.StatusBadRequest)
			return
		} else if !n.Members.Contains(msg.From) && msg.Type != message.ConnectType {
			// If the message peer is not a registered member of the current
			// network, return a forbidden HTTP error.
			http.Error(w, "Peer not registered", http.StatusForbidden)
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
			// When broadcast or direct message is received it will be redirected
			// to the inbox messages channel where the user will be waiting for
			// read it.
			n.Inbox <- msg
		case message.DisconnectType:
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

func (n *Node) handelSSE() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the http.Flusher of the current response writer to stream data
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported.", http.StatusInternalServerError)
			return
		}

		// Set Server Send Events compatible headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		// Parse open stream message and
		msg := new(message.Message).FromRequest(r)
		if !n.Members.Contains(msg.From) {
			http.Error(w, "Not connected peer, perform a connect request first.", http.StatusForbidden)
			return
		}

		msgChan := n.Members.WebChan(msg.From)

		// Handling Outbox messages chan to stream it to the client and
		// disconnection events throught request Context Done channel.
		for {
			select {
			case <-r.Context().Done():
				close(msgChan)
				n.Members.Delete(msg.From)
				if n.Members.Len() == 0 {
					n.setConnected(false)
				}
				return
			case data := <-msgChan:
				fmt.Fprintf(w, "data: %s\n\n", data)
				// Flush the data instead of buffering it
				flusher.Flush()
			}
		}
	}
}
