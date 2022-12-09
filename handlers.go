package gop2p

import "net/http"

func (node *Node) handle() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		var msg = new(Message).FromRequest(r)

		if r.Method == http.MethodGet {
			var result = node.connectHandler(msg)

			w.Header().Set("Content-Type", "text/plain")
			w.Write(result)
		} else if r.Method == http.MethodPost {
			node.Inbox <- msg
		} else if r.Method == http.MethodDelete {
			node.disconnectHandler(msg)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 - Method not allowed!"))
		}
	}
}

func (node *Node) connectHandler(msg *Message) (members []byte) {
	// Get members safely to send them to the node that is trying to join
	node.membersMtx.Lock()
	var currentMembers = append(Peers{}, node.members...)
	node.membersMtx.Unlock()

	// Encoding current list of members to a JSON to send it
	var err error
	if members, err = currentMembers.ToJSON(); err != nil {
		// TODO: handle error
		node.Error <- ParseErr("error encoding members to JSON", err, msg)
	}

	// Update the current member list safely appending the Message.From Peer
	node.membersMtx.Lock()
	defer node.membersMtx.Unlock()
	if !node.members.Contains(msg.From) {
		node.members = append(node.members, msg.From)
	}

	return
}

func (node *Node) disconnectHandler(msg *Message) {
	// Delete the Message.From Peer from the current member list safely
	node.membersMtx.Lock()
	defer node.membersMtx.Unlock()
	node.members = node.members.Delete(msg.From)
}
