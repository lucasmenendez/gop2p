package node

import (
	"io"
	"net/http"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func Test_connect(t *testing.T) {
	c := qt.New(t)

	t.Run("success connection", func(t *testing.T) {
		me, _ := peer.Me(5001, false)
		srv, port := testServer(func(w http.ResponseWriter, r *http.Request) {
			c.Assert(r.Host, qt.Equals, me.String())

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
			c.Assert(err.ErrCode, qt.Equals, CONNECTION_ERR)
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
			c.Assert(err.ErrCode, qt.Equals, CONNECTION_ERR)
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
			c.Assert(r.Host, qt.Equals, me.String())
		}
	})
	defer srv.Close()

	t.Run("disconnection fails", func(t *testing.T) {
		client := New(me)
		go func() {
			err := <-client.Error
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.ErrorAs, new(*NodeErr))
			c.Assert(err.ErrCode, qt.Equals, CONNECTION_ERR)
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
		resMsg, err := io.ReadAll(r.Body)
		c.Assert(err, qt.IsNil)
		msg := new(message.Message).SetJSON(resMsg)

		if msg.Type == message.ConnectType {
			res, _ := peer.NewMembers().ToJSON()
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		} else if msg.Type == message.BroadcastType {
			c.Assert(r.Host, qt.Equals, me.String())

			c.Assert(err, qt.IsNil)
			c.Assert(msg.Data, qt.DeepEquals, expMsg)
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
			c.Assert(err.ErrCode, qt.Equals, CONNECTION_ERR)
		}()

		msg := new(message.Message).SetFrom(me).SetData(expMsg)
		client.broadcast(msg)
	})
}

func Test_send(t *testing.T) {
	c := qt.New(t)

	directData := []byte("private")
	srv, port := testServer(func(w http.ResponseWriter, r *http.Request) {
		resMsg, err := io.ReadAll(r.Body)
		c.Assert(err, qt.IsNil)
		msg := new(message.Message).SetJSON(resMsg)

		if msg.Type == message.ConnectType {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
			return
		} else if msg.Type == message.DirectType {
			c.Assert(msg.Data, qt.DeepEquals, directData)
		}
	})
	defer srv.Close()

	entryPoint, _ := peer.Me(port, false)
	me, _ := peer.Me(getRandomPort(), false)
	client := New(me)
	client.connect(entryPoint)

	t.Run("success send", func(t *testing.T) {
		msg := new(message.Message).SetFrom(me).SetData(directData).SetTo(entryPoint)
		err := client.send(msg)
		c.Assert(err, qt.DeepEquals, (*NodeErr)(nil))
	})

	t.Run("broadcast fails", func(t *testing.T) {
		client.disconnect()
		go func() {
			err := <-client.Error
			c.Assert(err, qt.IsNotNil)
			c.Assert(err, qt.ErrorAs, new(*NodeErr))
			c.Assert(err.ErrCode, qt.Equals, CONNECTION_ERR)
		}()

		msg := new(message.Message).SetFrom(me).SetType(message.DirectType)
		client.send(msg)
	})
}
