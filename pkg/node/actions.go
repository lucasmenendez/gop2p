package node

import (
	"io"
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

// connect function allows to a node to join to a network using a knowed a peer
// that is already into that network. The function request a connection to that
// peer and it response with the current network members. To complete the
// joining, the current node send the same request to ever member received to
// populate its information.
func (node *Node) connect(entryPoint *peer.Peer) {
	// Create the request using a connection message.
	var msg = new(message.Message).SetType(message.ConnectType).SetFrom(node.Self)
	var req, err = msg.GetRequest(entryPoint.Hostname())
	if err != nil {
		node.Error <- ParseErr("error encoding message to request", err, msg)
	}

	// Try to join into the network through the provided peer
	var res *http.Response
	if res, err = node.client.Do(req); err != nil {
		// TODO: handle error
		node.Error <- ConnErr("error trying to connect to a peer", err, msg)
	}

	// Reading the list of current members of the network from the peer
	// response.
	var body []byte
	defer res.Body.Close()
	if body, err = io.ReadAll(res.Body); err != nil {
		node.Error <- ParseErr("error reading peer response body", err, msg)
	}

	// Parsing the received list
	var receivedMembers = peer.EmptyMembers()
	if receivedMembers, err = receivedMembers.FromJSON(body); err != nil {
		node.Error <- ParseErr("error parsing incoming member list", err, msg)
	}

	// Update current members
	for _, receivedPeer := range receivedMembers.Peers() {
		node.Members.Append(receivedPeer)
	}

	// Send the same message to each member to greet them and append to the
	// registered members the entrypoint (after send the broadcast the greet to
	// avoid unnecesary calls).
	node.broadcast(msg)
	node.Members.Append(entryPoint)
}

// disconnect function perform a graceful disconnection, stopping other routines
// warning to other network members about the disconnection and closing
// remaining channels.
func (node *Node) disconnect() {
	// Stop routines at the end
	defer node.waiter.Done()

	// Warn to other network peers about the disconnection
	var msg = new(message.Message).SetType(message.DisconnectType).SetFrom(node.Self)
	node.broadcast(msg)

	// Closing channels
	close(node.Inbox)
	close(node.Outbox)
	close(node.Connect)
}

// broadcast function sends the message provided to every peer registered on the
// node network. It iterates over current network peers performing a
// http.Request from the message provided.
func (node *Node) broadcast(msg *message.Message) {
	// Iterate over each member encoding as a request and performing it with
	// the provided Message.
	for _, peer := range node.Members.Peers() {
		var req, err = msg.GetRequest(peer.Hostname())
		if err != nil {
			node.Error <- ParseErr("error decoding request to message", err, msg)
		}

		if _, err = node.client.Do(req); err != nil {
			node.Error <- ConnErr("error trying to perform the request", err, msg)
		}
	}
}
