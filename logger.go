package gop2p

import (
	"fmt"
	"log"
)

// log function logs message provided formated and adding seld peer information
// trace.
func (n *Node) log(m string, args ...interface{}) {
	if n.debug {
		m = fmt.Sprintf(m, args...)
		log.Printf("[%s:%s] %s\n", n.Self.Address, n.Self.Port, m)
	}
}
