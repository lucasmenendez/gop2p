package node

import (
	"io"
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

// setConnected function updates the current node status safely using a mutex.
func (node *Node) setConnected(connected bool) {
	node.connMtx.Lock()
	defer node.connMtx.Unlock()
	node.connected = connected
}

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
		return
	}

	// Try to join into the network through the provided peer
	var res *http.Response
	if res, err = node.client.Do(req); err != nil {
		node.Error <- ConnErr("error trying to connect to a peer", err, msg)
		return
	}

	// Reading the list of current members of the network from the peer
	// response.
	var body []byte
	defer res.Body.Close()
	if body, err = io.ReadAll(res.Body); err != nil {
		node.Error <- ParseErr("error reading peer response body", err, msg)
		return
	}

	// Parsing the received list
	var receivedMembers = peer.NewMembers()
	if receivedMembers, err = receivedMembers.FromJSON(body); err != nil {
		node.Error <- ParseErr("error parsing incoming member list", err, msg)
		return
	}

	// Update current members
	for _, receivedPeer := range receivedMembers.Peers() {
		if !node.Self.Equal(receivedPeer) {
			node.Members.Append(receivedPeer)
		}
	}

	// Init Leave channel, this channel remain opened while the node is
	// connected to a network.
	node.setConnected(true)

	// Send the same message to each member to greet them and append to the
	// registered members the entrypoint (after send the broadcast the greet to
	// avoid unnecesary calls).
	node.broadcast(msg)
	node.Members.Append(entryPoint)

}

// disconnect function perform a graceful disconnection, warning to other
// network members about the disconnection and deleting registered peers from
// current member list.
func (node *Node) disconnect() {
	// Warn to other network peers about the disconnection
	var msg = new(message.Message).SetType(message.DisconnectType).SetFrom(node.Self)
	node.broadcast(msg)

	// Clean current member list
	node.Members = peer.NewMembers()
	node.setConnected(false)
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
