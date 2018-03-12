package main

import "github.com/lucasmenendez/gop2p"

func main() {
	entryNode := p2p.New("entryPeer", 5001)
	entryNode.Init()

	childNode := p2p.New("peer", 5002)
	childNode.Init()

	entryNode.Join(childNode)

	entryNode.Broadcast("HELLO NETWORK!")
}
