package node

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

var httpClient = &http.Client{}

func testServer(h func(http.ResponseWriter, *http.Request)) (*httptest.Server, int) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", h)
	srv := httptest.NewServer(handler)

	srvData, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(srvData.Port())
	return srv, port
}

func initNode(t *testing.T, port int) *Node {
	c := qt.New(t)

	// Init current node
	p, err := peer.Me(port, false)
	c.Assert(err, qt.IsNil)
	n := New(p)

	return n
}

func prepareRequest(t *testing.T, reqType, to, from int, data []byte) *http.Request {
	c := qt.New(t)

	toPeer, err := peer.Me(to, false)
	c.Assert(err, qt.IsNil)
	fromPeer, err := peer.Me(from, false)
	c.Assert(err, qt.IsNil)

	msg := new(message.Message).SetFrom(fromPeer).SetType(reqType)
	if reqType == message.BroadcastType {
		msg.SetData(data)
	} else if reqType == message.DirectType {
		msg.SetData(data)
		msg.SetTo(toPeer)
	}

	req, err := msg.GetRequest(toPeer.Hostname())
	c.Assert(err, qt.IsNil)

	return req
}

func getRandomPort() int {
	minSafePort, maxSafePort := 49152, 65535
	limit := new(big.Int).SetInt64(int64(maxSafePort - minSafePort))
	r, _ := rand.Int(rand.Reader, limit)
	return int(r.Int64()) + minSafePort
}
