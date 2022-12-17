package node

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

var httpClient = &http.Client{}

func Test_startListening(t *testing.T) {
	c := qt.New(t)

	me, _ := peer.Me(5001, false)
	msg := new(message.Message).SetType(message.ConnectType).SetFrom(me)
	entryPoint, _ := peer.Me(5002, false)
	req, _ := msg.GetRequest(entryPoint.Hostname())

	t.Run("request to non existing server", func(t *testing.T) {
		_, err := httpClient.Do(req)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err, qt.ErrorMatches, "(.*)connection refused(.*)")
	})

	srv := New(entryPoint)
	go srv.startListening()

	t.Run("request to started server", func(t *testing.T) {
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		_, err = io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)

		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		_, err = io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)
	})

	newSrv := New(entryPoint)
	t.Run("start a server with already started server info", func(t *testing.T) {
		go func() {
			err := <-newSrv.Error
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.ErrorAs, new(*NodeErr))
			c.Assert((err.(*NodeErr)).ErrCode, qt.Equals, INTERNAL_ERR)
		}()
		newSrv.startListening()
	})
}

func Test_handleRequest(t *testing.T) {
	c := qt.New(t)

	me, _ := peer.Me(5001, false)
	srv := New(me)
	srv.Start()

	t.Run("invalid request method", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, me.Hostname(), nil)
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusBadRequest)
	})
	// invalid connection request (invalid json)

	t.Run("invalid requests", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, me.Hostname(), nil)
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusBadRequest)

		req, _ = http.NewRequest(http.MethodDelete, me.Hostname(), nil)
		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusBadRequest)
	})

	// valid requests (connect, disconnect, message)
	t.Run("valid requests", func(t *testing.T) {
		p, _ := peer.Me(5002, false)

		msg := new(message.Message).SetFrom(p).SetType(message.ConnectType)
		req, _ := msg.GetRequest(me.Hostname())
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusOK)

		body, err := io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)
		res.Body.Close()
		c.Assert(body, qt.DeepEquals, []byte("[]"))

		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
		body, err = io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)
		res.Body.Close()
		c.Assert(body, qt.DeepEquals, []byte("[{\"port\":"+fmt.Sprint(p.Port)+",\"address\":\""+p.Address+"\"}]"))
	})
}
