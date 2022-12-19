package node

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func TestNodeStart(t *testing.T) {
	c := qt.New(t)

	p1, _ := peer.Me(5001, false)
	c.Assert(New(p1), qt.Not(qt.DeepEquals), new(Node))
}

func TestNodeIsConnected(t *testing.T) {
	c := qt.New(t)
	c.Assert(true, qt.IsTrue)
}

func TestNodeWait(t *testing.T) {
	c := qt.New(t)
	c.Assert(true, qt.IsTrue)
}

func TestNodeStop(t *testing.T) {
	c := qt.New(t)

	p1, _ := peer.Me(5001, false)

	n1 := New(p1)
	err := n1.Stop()
	c.Assert(err, qt.IsNil)

	n1 = New(p1)
	n1.Start()

	err = n1.Stop()
	c.Assert(err, qt.IsNil)
}
