package node

import (
	"net/http"
	"sync"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func TestNodeStart(t *testing.T) {
	c := qt.New(t)

	// Init current node and external peer
	n := initNode(t, getRandomPort())
	p, err := peer.Me(getRandomPort(), false)
	c.Assert(err, qt.IsNil)

	// Prepare a connection request
	req := prepareRequest(t, message.ConnectType, n.Self.Port, p.Port, nil)

	// Perfom a connection request that must fail because the current node is
	// not started yet
	_, err = httpClient.Do(req)
	c.Assert(err, qt.IsNotNil)
	c.Assert(n.Members.Len(), qt.Equals, 0)

	// Start current node and repeate the request, now it must success
	n.Start()
	res, err := httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(n.Members.Len(), qt.Equals, 1)
	c.Assert(n.Members.Contains(p), qt.IsTrue)

	// Prepare a plain message to emulate broadcast
	req = prepareRequest(t, message.BroadcastType, n.Self.Port, p.Port, []byte("test"))

	// Start a goroutine to handle new messages with a wait group associated
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for msg := range n.Inbox {
			c.Assert(msg.Data, qt.DeepEquals, []byte("test"))
			wg.Done()
			return
		}
	}()

	// Perform the prepared message broadcast that must success
	res, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	wg.Wait()

	// Prepare a disconnection message to emulate a disconnection request
	req = prepareRequest(t, message.DisconnectType, n.Self.Port, p.Port, nil)

	// Perform the disconnection request
	res, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(n.Members.Len(), qt.Equals, 0)

	// Stop the the current node
	err = n.Stop()
	c.Assert(err, qt.IsNil)
}

func TestNodeIsConnected(t *testing.T) {
	c := qt.New(t)

	// Init the current node, it must not be connected
	n := initNode(t, 5001)
	c.Assert(n.IsConnected(), qt.IsFalse)

	// Prepare a connection request and start the current node and check if the
	// current node still disconnected
	req := prepareRequest(t, message.ConnectType, 5001, 5002, nil)
	n.Start()
	c.Assert(n.IsConnected(), qt.IsFalse)

	// Perform the connection request and assert that it is connected
	res, err := httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(n.IsConnected(), qt.IsTrue)

	// Perform a disconnection request and assert it is disconnected
	req = prepareRequest(t, message.DisconnectType, 5001, 5002, nil)
	res, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	c.Assert(n.IsConnected(), qt.IsFalse)

	// Stop the current node
	err = n.Stop()
	c.Assert(err, qt.IsNil)
}

func TestNodeStop(t *testing.T) {
	c := qt.New(t)

	t.Run("stop not started node", func(t *testing.T) {
		n := initNode(t, getRandomPort())
		err := n.Stop()
		c.Assert(err, qt.IsNotNil)
	})

	t.Run("success start stop node", func(t *testing.T) {
		n := initNode(t, getRandomPort())
		n.Start()
		time.Sleep(time.Second)
		err := n.Stop()
		c.Assert(err, qt.IsNil)
	})

	t.Run("stop node connected a network with ghost perr", func(t *testing.T) {
		n := initNode(t, getRandomPort())
		n.Start()

		req := prepareRequest(t, message.ConnectType, n.Self.Port, getRandomPort(), nil)
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusOK)

		err = n.Stop()
		c.Assert(err, qt.IsNotNil)
	})
}
