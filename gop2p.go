// Package gop2p implements simple peer-to-peer network node in pure Go. Uses
// HTTP client and server to communicate over internet to knowed network
// members. gop2p implements the following functional workflow:
// 	1. Connect to the network: The client gop2p.Node know a entry point of the
// 		desired network (other gop2p.Node that is already connected). The entry
//		point response with the current network gop2p.Node's and updates its
//		members gop2p.Node list. The client gop2p.Node broadcast a connection
//		request to every gop2p.Node received from entry point.
// 	2. Broadcasting a message: The client gop2p.Node prepares and broadcast a
//		gop2p.Message to every network gop2p.Node.
// 	3. Disconnect from the network: The client gop2p.Node broadcast a
//		disconnection request to every network gop2p.Node. This gop2p.Node's
//		updates its current network members list unregistering the client
//		gop2p.Node.

package gop2p

import (
	"fmt"

	"github.com/lucasmenendez/gop2p/pkg/node"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func StartLocalNode(port int) *node.Node {
	var peer = peer.Me(port)
	var n = node.New(peer)
	n.Start()
	return n
}

func StartNode(address string, port int) *node.Node {
	var peer = &peer.Peer{Address: address, Port: fmt.Sprint(port)}
	var n = node.New(peer)
	n.Start()
	return n
}
