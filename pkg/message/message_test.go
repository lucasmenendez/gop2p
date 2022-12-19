package message

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func TestMessageSetType(t *testing.T) {
	c := qt.New(t)

	msg := new(Message)
	msg.SetType(ConnectType)
	c.Assert(msg.Type, qt.Equals, ConnectType)

	msg.SetType(DisconnectType)
	c.Assert(msg.Type, qt.Equals, DisconnectType)

	msg.SetType(PlainType)
	c.Assert(msg.Type, qt.Equals, PlainType)

	msg.SetType(-1)
	c.Assert(msg.Type, qt.Equals, PlainType)
}

func TestMessageSetFrom(t *testing.T) {
	c := qt.New(t)

	expected := &peer.Peer{Address: "localhost", Port: 8080}
	msg := new(Message).SetFrom(expected)
	c.Assert(msg.From, qt.DeepEquals, expected)

	expected.Address = "0.0.0.0"
	expected.Port = 8081
	msg.SetFrom(expected)
	c.Assert(msg.From, qt.DeepEquals, expected)
}

func TestMessageSetData(t *testing.T) {
	c := qt.New(t)

	msg := &Message{Type: ConnectType}
	data := []byte("test data")
	msg.SetData(data)

	c.Assert(msg.Data, qt.DeepEquals, data)
	c.Assert(msg.Type, qt.Equals, PlainType)
}

func TestMessageGetRequest(t *testing.T) {
	c := qt.New(t)

	from, _ := peer.Me(5000, false)
	msg := new(Message).SetData([]byte("EY")).SetFrom(from)

	buff := bytes.NewBuffer(msg.Data)
	expected, _ := http.NewRequest(http.MethodPost, from.Hostname(), buff)
	expected.Header.Add(addressHeader, from.Address)
	expected.Header.Add(portHeader, fmt.Sprint(from.Port))

	result, err := msg.GetRequest(from.Hostname())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Method, qt.Equals, expected.Method)
	c.Assert(result.Header.Get(addressHeader), qt.Equals, from.Address)
	c.Assert(result.Header.Get(portHeader), qt.Equals, fmt.Sprint(from.Port))

	resBody, expBody := []byte{}, []byte{}
	resLen, err := result.Body.Read(resBody)
	c.Assert(err, qt.IsNil)
	expLen, err := result.Body.Read(expBody)
	c.Assert(err, qt.IsNil)
	c.Assert(resLen, qt.Equals, expLen)
	c.Assert(resBody, qt.DeepEquals, expBody)

	msg = new(Message).SetData([]byte("EY"))
	_, err = msg.GetRequest(from.Hostname())
	c.Assert(err, qt.IsNotNil)

	msg = new(Message).SetFrom(from).SetType(ConnectType)
	expected, _ = http.NewRequest(http.MethodGet, from.Hostname(), nil)
	expected.Header.Add(addressHeader, from.Address)
	expected.Header.Add(portHeader, fmt.Sprint(from.Port))

	result, err = msg.GetRequest(from.Hostname())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Method, qt.Equals, expected.Method)
	c.Assert(result.Header.Get(addressHeader), qt.Equals, from.Address)
	c.Assert(result.Header.Get(portHeader), qt.Equals, fmt.Sprint(from.Port))
	resLen, err = result.Body.Read([]byte{})
	c.Assert(err, qt.IsNotNil)
	c.Assert(resLen, qt.DeepEquals, len(msg.Data))
}

func TestMessageFromRequest(t *testing.T) {
	c := qt.New(t)

	from, _ := peer.Me(5000, false)
	data := []byte("ey")
	expected := new(Message).SetData(data).SetFrom(from)

	buff := bytes.NewBuffer(data)
	req, _ := http.NewRequest(http.MethodPost, from.Hostname(), buff)
	req.Header.Add(addressHeader, from.Address)
	req.Header.Add(portHeader, fmt.Sprint(from.Port))

	result := new(Message).FromRequest(req)
	c.Assert(result.Type, qt.Equals, expected.Type)
	c.Assert(expected.From.Equal(result.From), qt.IsTrue)
	c.Assert(result.Data, qt.DeepEquals, expected.Data)

	req, _ = http.NewRequest(http.MethodGet, from.Hostname(), nil)
	c.Assert(new(Message).FromRequest(req), qt.IsNil)
}
