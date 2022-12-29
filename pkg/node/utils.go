package node

import (
	"bytes"
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

func composeRequest(msg *message.Message, to *peer.Peer) (*http.Request, error) {
	encMsg := msg.JSON()
	if encMsg == nil {
		return nil, ParseErr("error encoding message to JSON", nil)
	}
	body := bytes.NewBuffer(encMsg)
	req, err := http.NewRequest(http.MethodPost, to.Hostname(), body)
	if err != nil {
		return nil, ParseErr("error decoding request to message", err)
	}
	req.Host = msg.From.String()

	return req, nil
}

// safeClose function allows closing gracefully any Node channel avoiding
// closing a non-opened channel.
func safeClose[C *message.Message | *peer.Peer | *NodeErr](ch chan C) {
	select {
	case <-ch:
		close(ch)
		return
	default:
	}
}
