package node

import (
	"fmt"
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
func (n *Node) connect(entryPoint *peer.Peer) *NodeErr {
	// Create the request using a connection message.
	msg := new(message.Message).SetType(message.ConnectType).SetFrom(n.Self)
	req, err := composeRequest(msg, entryPoint)
	if err != nil {
		return ParseErr("error encoding message to request", err)
	}

	// Try to join into the network through the provided peer
	res, err := n.client.Do(req)
	if err != nil {
		return ConnErr("error trying to connect to a peer", err)
	}

	if code := res.StatusCode; code != http.StatusOK {
		err := fmt.Errorf("%d http status received from %s", code, entryPoint)
		return ConnErr("error making the request to a peer", err)
	}

	// Reading the list of current members of the network from the peer
	// response.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ParseErr("error reading peer response body", err)
	}
	res.Body.Close()

	// Parsing the received list
	receivedMembers := peer.NewMembers()
	if receivedMembers, err = receivedMembers.FromJSON(body); err != nil {
		return ParseErr("error parsing incoming member list", err)
	}

	// Update current members and send a connection request to all of them,
	// discarting the response received (the list of current members).
	for _, member := range receivedMembers.Peers() {
		// If a received peer is not the same that contains the current node try
		// to connect directly.
		if !n.Self.Equal(member) {
			if req, err := composeRequest(msg, member); err != nil {
				return ParseErr("error decoding request to message", err)
			} else if _, err := n.client.Do(req); err != nil {
				return ConnErr("error trying to perform the request", err)
			}
			n.Members.Append(member)
		}
	}

	// Set node status as connected.
	n.setConnected(true)
	// Append the entrypoint to the current members.
	n.Members.Append(entryPoint)
	return nil
}

// disconnect function perform a graceful disconnection, warning to other
// network members about the disconnection and deleting registered peers from
// current member list.
func (n *Node) disconnect() *NodeErr {
	// Send an error to Node.Error channel if the node is not connected
	if !n.IsConnected() {
		return ConnErr("node not connected", nil)
	}

	// Warn to other network peers about the disconnection
	msg := new(message.Message).SetType(message.DisconnectType).SetFrom(n.Self)
	if err := n.broadcast(msg); err != nil {
		return err
	}

	// Clean current member list
	n.Members = peer.NewMembers()
	n.setConnected(false)
	return nil
}

// broadcast function sends the message provided to every peer registered on the
// node network. It iterates over current network peers performing a
// http.Request from the message provided.
func (n *Node) broadcast(msg *message.Message) *NodeErr {
	// Send an error to Node.Error channel if the node is not connected
	if !n.IsConnected() {
		return ConnErr("node not connected", nil)
	}

	// Iterate over each member encoding as a request and performing it with
	// the provided Message.
	encMsg := msg.JSON()
	if encMsg == nil {
		return ParseErr("error encoding message to JSON", nil)
	}
	for _, member := range n.Members.Peers() {
		req, err := composeRequest(msg, member)
		if err != nil {
			return ParseErr("error decoding request to message", err)
		}

		if _, err := n.client.Do(req); err != nil {
			return ConnErr("error trying to perform the request", err)
		}
	}
	return nil
}

// send function sends the message provided to a single peer registered from the
// current node network.
func (n *Node) send(msg *message.Message) *NodeErr {
	if !n.IsConnected() {
		// Return an error if the current node is not connected
		return ConnErr("node not connected", nil)
	} else if msg.To == nil {
		// Return an error if no Message.To parameter is initialized
		return InternalErr("no intended peer defined at provided message", nil)
	}

	encMsg := msg.JSON()
	if encMsg == nil {
		return ParseErr("error encoding message to JSON", nil)
	}
	for _, to := range msg.To {
		if !n.Members.Contains(to) {
			// Return an error if the current network does not contains the
			// Message.To peer provided
			return ConnErr("target peer is not into the network", nil)
		}

		// Encode message as a request and send it
		req, err := composeRequest(msg, to)
		if err != nil {
			return ParseErr("error decoding request to message", err)
		}
		if _, err := n.client.Do(req); err != nil {
			return ConnErr("error trying to perform the request", err)
		}
	}
	return nil
}
