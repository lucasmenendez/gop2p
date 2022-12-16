package message

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/lucasmenendez/gop2p/peer"
)

func TestMessageSetType(t *testing.T) {
	var msg = new(Message)
	msg.SetType(ConnectType)
	if msg.Type != ConnectType {
		t.Errorf("expected message with ConnectType, got with other type (%d)", msg.Type)
	}

	msg.SetType(DisconnectType)
	if msg.Type != DisconnectType {
		t.Errorf("expected message with DisconnectType, got with other type (%d)", msg.Type)
	}

	msg.SetType(PlainType)
	if msg.Type != PlainType {
		t.Errorf("expected message with PlainType, got with other type (%d)", msg.Type)
	}
}

func TestMessageSetFrom(t *testing.T) {
	var expected = &peer.Peer{Address: "localhost", Port: 8080}
	var msg = new(Message).SetFrom(expected)
	if !expected.Equal(msg.From) {
		t.Errorf("expected %s, got %s", expected.String(), msg.From.String())
	}

	expected.Address = "0.0.0.0"
	expected.Port = 8081
	msg.SetFrom(expected)
	if !expected.Equal(msg.From) {
		t.Errorf("expected %s, got %s", expected.String(), msg.From.String())
	}
}

func TestMessageSetData(t *testing.T) {
	var msg = &Message{Type: ConnectType}
	var data = []byte("test data")
	msg.SetData(data)
	if !reflect.DeepEqual(msg.Data, data) {
		t.Errorf("expected %s, got %s", data, msg.Data)
	} else if msg.Type != PlainType {
		t.Errorf("expected message with PlainType, got with other type (%d)", msg.Type)
	}
}

func TestMessageGetRequest(t *testing.T) {
	var from = peer.Me(5000)
	var msg = new(Message).SetData([]byte("EY")).SetFrom(from)

	var buff = bytes.NewBuffer(msg.Data)
	var expected, _ = http.NewRequest(http.MethodPost, from.Hostname(), buff)
	expected.Header.Add(addressHeader, from.Address)
	expected.Header.Add(portHeader, fmt.Sprint(from.Port))

	var result, err = msg.GetRequest(from.Hostname())
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	} else if result.Method != expected.Method {
		t.Errorf("expected %s, got %s", expected.Method, result.Method)
	} else if result.Header.Get(addressHeader) != from.Address {
		t.Errorf("expected %s, got %s", from.Address, result.Header.Get(addressHeader))
	} else if result.Header.Get(portHeader) != fmt.Sprint(from.Port) {
		t.Errorf("expected %d, got %s", from.Port, result.Header.Get(portHeader))
	} else if body, _ := io.ReadAll(result.Body); !reflect.DeepEqual(msg.Data, body) {
		t.Errorf("expected %s, get %s", string(msg.Data), string(body))
	}

	msg = new(Message).SetData([]byte("EY"))
	_, err = msg.GetRequest(from.Hostname())
	if err == nil {
		t.Error("expected error, got nil")
	}

	msg = new(Message).SetFrom(from).SetType(ConnectType)
	expected, _ = http.NewRequest(http.MethodGet, from.Hostname(), nil)
	expected.Header.Add(addressHeader, from.Address)
	expected.Header.Add(portHeader, fmt.Sprint(from.Port))

	result, err = msg.GetRequest(from.Hostname())
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	} else if result.Method != expected.Method {
		t.Errorf("expected %s, got %s", expected.Method, result.Method)
	} else if result.Header.Get(addressHeader) != from.Address {
		t.Errorf("expected %s, got %s", from.Address, result.Header.Get(addressHeader))
	} else if result.Header.Get(portHeader) != fmt.Sprint(from.Port) {
		t.Errorf("expected %d, got %s", from.Port, result.Header.Get(portHeader))
	}
}

func TestMessageFromRequest(t *testing.T) {
	var from = peer.Me(5000)
	var data = []byte("ey")
	var expected = new(Message).SetData(data).SetFrom(from)

	var buff = bytes.NewBuffer(data)
	var req, _ = http.NewRequest(http.MethodPost, from.Hostname(), buff)
	req.Header.Add(addressHeader, from.Address)
	req.Header.Add(portHeader, fmt.Sprint(from.Port))

	var result = new(Message).FromRequest(req)
	if expected.Type != result.Type {
		t.Errorf("expected %d, got %d", expected.Type, result.Type)
	} else if !expected.From.Equal(result.From) {
		t.Errorf("expected %s, got %s", expected.From.String(), result.From.String())
	} else if !reflect.DeepEqual(expected.Data, result.Data) {
		t.Errorf("expected %s, got %s", fmt.Sprint(expected.Data), fmt.Sprint(result.Data))
	}

	req, _ = http.NewRequest(http.MethodGet, from.Hostname(), nil)
	if result = new(Message).FromRequest(req); result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}
