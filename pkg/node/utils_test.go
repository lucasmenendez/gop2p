package node

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func Test_setConnected(t *testing.T) {
	c := qt.New(t)

	n := initNode(t, getRandomPort())
	c.Assert(n.connected, qt.IsFalse)
	n.setConnected(true)
	c.Assert(n.connected, qt.IsTrue)
	n.setConnected(false)
	c.Assert(n.connected, qt.IsFalse)
}

func Test_composeRequest(t *testing.T) {
	c := qt.New(t)

	to, _ := peer.Me(getRandomPort(), false)
	n := initNode(t, getRandomPort())
	msg := new(message.Message).SetFrom(n.Self).SetData([]byte("test"))

	body := bytes.NewBuffer(msg.JSON())
	expected, err := http.NewRequest(http.MethodPost, to.Hostname(), body)
	expected.Host = msg.From.String()
	c.Assert(err, qt.IsNil)

	result, err := composeRequest(msg, to)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Method, qt.Equals, expected.Method)
	c.Assert(result.Host, qt.Equals, expected.Host)
	expBody, err := io.ReadAll(result.Body)
	c.Assert(err, qt.IsNil)
	c.Assert(expBody, qt.DeepEquals, msg.JSON())
}

func Test_safeClose(t *testing.T) {
	c := qt.New(t)

	defer func() {
		perr := recover()
		c.Assert(perr, qt.IsNil)
	}()

	testChan := make(chan *message.Message)
	safeClose(testChan) // opened channel
	safeClose(testChan) // closed channel, prevent panics

}
