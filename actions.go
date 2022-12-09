package gop2p

import (
	"io"
	"net/http"
)

// connect function
func (node *Node) connect(peer *Peer) {
	// Create the connect request using a message
	var msg = new(Message).SetType(CONNECT).SetFrom(node.Self)
	var req, err = msg.GetRequest(peer.Hostname())
	if err != nil {
		// TODO: handle error
		node.Error <- ParseErr("error encoding message to request", err, msg)
	}

	// Try to join into the network through the provided peer
	var res *http.Response
	if res, err = node.client.Do(req); err != nil {
		// TODO: handle error
		node.Error <- ConnErr("error trying to connect to a peer", err, msg)
	}

	// Parsing the list of current members of the network from the peer response
	var body []byte
	defer res.Body.Close()
	if body, err = io.ReadAll(res.Body); err != nil {
		// TODO: handle error
		node.Error <- ParseErr("error reading peer response body", err, msg)
	}

	var membersData = new(Peers)
	if membersData, err = membersData.FromJSON(body); err != nil {
		// TODO: handle error
		node.Error <- ParseErr("error parsing incoming member list", err, msg)
	}

	// Update current members
	node.membersMtx.Lock()
	for _, member := range *membersData {
		if !node.members.Contains(member) {
			var members = append(*node.members, member)
			node.members = &members
		}
	}
	node.membersMtx.Unlock()

	// Send the same message to each member to greet them
	node.broadcast(msg)

	// Safe append the connected peer
	node.membersMtx.Lock()
	defer node.membersMtx.Unlock()
	var members = append(*node.members, peer)
	node.members = &members
}

// disconnect function
func (node *Node) disconnect() {
	defer node.waiter.Done()

	var msg = new(Message).SetType(DISCONNECT).SetFrom(node.Self)
	node.broadcast(msg)

	close(node.Inbox)
	close(node.Outbox)
}

// broadcast function
func (node *Node) broadcast(msg *Message) {
	// Get current members safely
	node.membersMtx.Lock()
	var currentMembers = append(Peers{}, *node.members...)
	node.membersMtx.Unlock()

	// Iterate over each member encoding as a request and performing it with
	// the provided Message.
	for _, peer := range currentMembers {
		var req, err = msg.GetRequest(peer.Hostname())
		if err != nil {
			// TODO: handle error
			node.Error <- ParseErr("error decoding request to message", err, msg)
		}

		if _, err = node.client.Do(req); err != nil {
			// TODO: handle error
			node.Error <- ConnErr("error trying to perform the request", err, msg)
		}
	}
}
