package gop2p

import "fmt"

// message struct includes the content of a message and its sender peer.
type message struct {
	data []byte
	from Peer
}

// String function returns a human-readable version of message struct.
func (m message) String() string {
	return fmt.Sprintf("'%s' {from %s:%s}", m.data, m.from.Address, m.from.Port)
}
