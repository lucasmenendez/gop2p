package main

import (
	"github.com/lucasmenendez/gop2p"
	"time"
)

func main() {
	entryNode := gop2p.InitNode("entryPeer", 5001)
	childNode := gop2p.InitNode("peer", 5002)

	time.Sleep(time.Second)
	entryNode.Join(childNode)
	time.Sleep(time.Second)

	childNode.Broadcast("HELLO NETWORK!")
	time.Sleep(time.Second)
}
