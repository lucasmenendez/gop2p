package node

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/message"
)

func Test_startListening(t *testing.T) {
	c := qt.New(t)

	srv := initNode(t, getRandomPort())

	t.Run("request to non existing server", func(t *testing.T) {
		req := prepareRequest(t, message.ConnectType, srv.Self.Port, getRandomPort(), nil)
		_, err := httpClient.Do(req)
		c.Assert(err, qt.IsNotNil)
		c.Assert(err, qt.ErrorMatches, "(.*)connection refused(.*)")
	})

	t.Run("request to started server", func(t *testing.T) {
		srv.server = &http.Server{Addr: srv.Self.String()}
		go srv.startListening()

		req := prepareRequest(t, message.ConnectType, srv.Self.Port, getRandomPort(), nil)

		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		_, err = io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)

		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		_, err = io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)
	})

	t.Run("start a server with already started server info", func(t *testing.T) {
		newSrv := initNode(t, srv.Self.Port)

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for err := range newSrv.Error {
				c.Assert(err, qt.IsNotNil)
				c.Assert(err, qt.ErrorAs, new(*NodeErr))
				c.Assert(err.ErrCode, qt.Equals, INTERNAL_ERR)
				wg.Done()
			}
		}()
		newSrv.server = &http.Server{Addr: newSrv.Self.String()}
		go newSrv.startListening()
		wg.Wait()
	})
}

func Test_handleRequest(t *testing.T) {
	c := qt.New(t)

	t.Run("invalid request method", func(t *testing.T) {
		srv := initNode(t, getRandomPort())
		srv.Start()

		req, _ := http.NewRequest(http.MethodPut, srv.Self.Hostname(), nil)
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusBadRequest)
	})
	// invalid connection request (invalid json)

	t.Run("invalid requests", func(t *testing.T) {
		srv := initNode(t, getRandomPort())
		srv.Start()

		req, _ := http.NewRequest(http.MethodGet, srv.Self.Hostname(), nil)
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusBadRequest)

		req, _ = http.NewRequest(http.MethodDelete, srv.Self.Hostname(), nil)
		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusBadRequest)
	})

	// valid requests (connect, disconnect, message)
	t.Run("valid requests", func(t *testing.T) {
		srv := initNode(t, getRandomPort())
		srv.Start()

		firstPort := getRandomPort()
		req := prepareRequest(t, message.ConnectType, srv.Self.Port, firstPort, nil)
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusOK)

		body, err := io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)
		c.Assert(body, qt.DeepEquals, []byte("[]"))

		req = prepareRequest(t, message.ConnectType, srv.Self.Port, getRandomPort(), nil)
		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusOK)

		body, err = io.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)
		c.Assert(body, qt.DeepEquals, []byte("[{\"port\":"+fmt.Sprint(firstPort)+",\"address\":\"localhost\"}]"))
	})

	t.Run("plain message from external peer", func(t *testing.T) {
		srv := initNode(t, getRandomPort())
		srv.Start()
		peerPort := getRandomPort()

		req := prepareRequest(t, message.PlainType, srv.Self.Port, peerPort, nil)
		res, err := httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusForbidden)

		req = prepareRequest(t, message.DisconnectType, srv.Self.Port, peerPort, nil)
		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusForbidden)
	})
}
