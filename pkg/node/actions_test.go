package node

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func testServer(h func(http.ResponseWriter, *http.Request)) (*httptest.Server, int) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", h)
	srv := httptest.NewServer(handler)

	srvData, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(srvData.Port())
	return srv, port
}

func Test_setConnected(t *testing.T) {
	c := qt.New(t)

	p, _ := peer.Me(5000, false)
	n := New(p)
	c.Assert(n.connected, qt.IsFalse)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func(n *Node, c *qt.C, wg *sync.WaitGroup) {
		state := true
		for i := 0; i < 10; i++ {
			n.setConnected(state)
			c.Assert(n.connected, qt.Equals, state)
			state = !state
		}
		wg.Done()
	}(n, c, wg)

	state := true
	for i := 0; i < 10; i++ {
		n.setConnected(state)
		c.Assert(n.connected, qt.Equals, state)
		state = !state
	}
	wg.Wait()
}

func Test_connect(t *testing.T) {
	c := qt.New(t)

	t.Run("success connection", func(t *testing.T) {
		me, _ := peer.Me(5001, false)
		srv, port := testServer(func(w http.ResponseWriter, r *http.Request) {
			c.Assert(r.Method, qt.Equals, http.MethodGet)
			c.Assert(r.Header.Get("PEER_ADDRESS"), qt.Equals, me.Address)
			c.Assert(r.Header.Get("PEER_PORT"), qt.Equals, fmt.Sprint(me.Port))

			res, _ := peer.NewMembers().ToJSON()
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		})
		defer srv.Close()

		entryPoint, _ := peer.Me(port, false)
		client := New(me)
		client.connect(entryPoint)
		c.Assert(client.Members.Contains(entryPoint), qt.IsTrue)
	})

	t.Run("no server listening", func(t *testing.T) {
		me, _ := peer.Me(5001, false)
		entryPoint, _ := peer.Me(5045, false)
		client := New(me)
		go func() {
			err := <-client.Error
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.ErrorAs, new(*NodeErr))
			c.Assert((err.(*NodeErr)).ErrCode, qt.Equals, CONNECTION_ERR)
		}()
		client.connect(entryPoint)
	})

	t.Run("connection fails", func(t *testing.T) {
		me, _ := peer.Me(5001, false)
		srv, port := testServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer srv.Close()

		entryPoint, _ := peer.Me(port, false)
		client := New(me)
		go func() {
			err := <-client.Error
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.ErrorAs, new(*NodeErr))
			c.Assert((err.(*NodeErr)).ErrCode, qt.Equals, CONNECTION_ERR)
		}()
		client.connect(entryPoint)
	})
}

func Test_disconnect(t *testing.T) {
	c := qt.New(t)

	me, _ := peer.Me(5001, false)
	srv, port := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			res, _ := peer.NewMembers().ToJSON()
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		} else {
			c.Assert(r.Method, qt.Equals, http.MethodDelete)
			c.Assert(r.Header.Get("PEER_ADDRESS"), qt.Equals, me.Address)
			c.Assert(r.Header.Get("PEER_PORT"), qt.Equals, fmt.Sprint(me.Port))
		}
	})
	defer srv.Close()

	t.Run("disconnection fails", func(t *testing.T) {
		client := New(me)
		go func() {
			err := <-client.Error
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.ErrorAs, new(*NodeErr))
			c.Assert((err.(*NodeErr)).ErrCode, qt.Equals, CONNECTION_ERR)
		}()
		client.disconnect()
	})

	t.Run("success disconnection", func(t *testing.T) {
		client := New(me)
		entryPoint, _ := peer.Me(port, false)
		client.connect(entryPoint)
		client.disconnect()
		c.Assert(client.Members.Contains(entryPoint), qt.IsFalse)
	})
}

func Test_broadcast(t *testing.T) {
	c := qt.New(t)

	me, _ := peer.Me(5001, false)
	expMsg := []byte("ey")
	srv, port := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			res, _ := peer.NewMembers().ToJSON()
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		} else if r.Method != http.MethodDelete {
			c.Assert(r.Method, qt.Equals, http.MethodPost)
			c.Assert(r.Header.Get("PEER_ADDRESS"), qt.Equals, me.Address)
			c.Assert(r.Header.Get("PEER_PORT"), qt.Equals, fmt.Sprint(me.Port))

			resMsg, err := io.ReadAll(r.Body)
			c.Assert(err, qt.IsNil)
			c.Assert(resMsg, qt.DeepEquals, expMsg)
		}
	})
	defer srv.Close()

	entryPoint, _ := peer.Me(port, false)
	client := New(me)
	client.connect(entryPoint)

	t.Run("success broadcast", func(t *testing.T) {
		time.Sleep(time.Second)
		msg := new(message.Message).SetFrom(me).SetData(expMsg)
		client.broadcast(msg)
	})

	t.Run("broadcast fails", func(t *testing.T) {
		client.disconnect()
		go func() {
			err := <-client.Error
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.ErrorAs, new(*NodeErr))
			c.Assert((err.(*NodeErr)).ErrCode, qt.Equals, CONNECTION_ERR)
		}()

		msg := new(message.Message).SetFrom(me).SetData(expMsg)
		client.broadcast(msg)
	})
}
