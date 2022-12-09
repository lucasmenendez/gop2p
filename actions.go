package gop2p

import (
	"io"
	"net/http"
)

func (node *Node) Connect(peer *Peer) {
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

	var membersData = Peers{}
	if membersData, err = PeersFromJSON(body); err != nil {
		// TODO: handle error
		node.Error <- ParseErr("error parsing incoming member list", err, msg)
	}

	// Update current members
	node.membersMtx.Lock()
	for _, member := range membersData {
		if !node.members.Contains(member) {
			node.members = append(node.members, member)
		}
	}
	node.membersMtx.Unlock()

	// Send the same message to each member to greet them
	node.Broadcast(msg)

	// Safe append the connected peer
	node.membersMtx.Lock()
	defer node.membersMtx.Unlock()
	node.members = append(node.members, peer)
}

func (node *Node) Disconnect() {
	var msg = new(Message).SetType(DISCONNECT).SetFrom(node.Self)
	node.Broadcast(msg)
}

func (node *Node) Broadcast(msg *Message) {
	// Get current members safely
	node.membersMtx.Lock()
	var currentMembers = append(Peers{}, node.members...)
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
