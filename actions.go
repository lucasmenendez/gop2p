package gop2p

import (
	"io"
	"net/http"
)

func (node *Node) Connect(peer Peer) {
	// Create the connect request using a message
	var msg = new(Message).SetType(CONNECT).SetFrom(node.Self)
	var req, err = msg.Request(peer.URI())
	if err != nil {
		// handle error
		node.Logger.Fatalf("[%s] error creating the connection request: %v\n",
			node.Self.String(), err)
	}

	// Try to join into the network through the provided peer
	var res *http.Response
	if res, err = node.client.Do(req); err != nil {
		// handle error
		node.Logger.Fatalf("[%s] error sending connection request: %v\n",
			node.Self.String(), err)
	}

	// Parsing the list of current members of the network from the peer response
	var body []byte
	defer res.Body.Close()
	if body, err = io.ReadAll(res.Body); err != nil {
		// handle error
		node.Logger.Fatalf("[%s] error parsing connection response: %v\n",
			node.Self.String(), err)
	}

	var membersData = Peers{}
	if membersData, err = PeersFromJSON(body); err != nil {
		node.Logger.Fatalf("[%s] error decoding member list: %v\n",
			node.Self.String(), err)
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
	node.membersMtx.Lock()
	var currentMembers = append(Peers{}, node.members...)
	node.membersMtx.Unlock()

	for _, peer := range currentMembers {
		var req, err = msg.Request(peer.URI())
		if err != nil {
			// handle error
			node.Logger.Fatalf("[%s] error creating the message request: %v\n",
				node.Self.String(), err)
		}

		if _, err = node.client.Do(req); err != nil {
			// handle error
			node.Logger.Fatalf("[%s] error sending the message request: %v\n",
				node.Self.String(), err)
		}
	}
}
