package gop2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	CONNECT    = iota
	DISCONNECT = iota
	PLAIN      = iota
)

// addressHeader contains default environment variable that contains address.
const addresHeader string = "PEER_ADDRESS"

// portHeader contains default environment variable that contains port.
const portHeader string = "PEER_PORT"

// Message struct includes the content of a Message and its sender peer.
type Message struct {
	Type int
	Data []byte
	From *Peer
}

// SetType function
func (m *Message) SetType(t int) *Message {
	if t == CONNECT || t == DISCONNECT {
		m.Type = t
	} else {
		m.Type = PLAIN
	}

	return m
}

// SetFrom function
func (m *Message) SetFrom(peer *Peer) *Message {
	m.From = peer
	return m
}

// SetData function
func (m *Message) SetData(data []byte) *Message {
	m.Type = PLAIN
	m.Data = data
	return m
}

// String function returns a human-readable version of Message struct.
func (m *Message) String() string {
	return fmt.Sprintf("'%s' {from %s:%s}", string(m.Data), m.From.Address, m.From.Port)
}

// GetRequest function
func (m *Message) GetRequest(uri string) (*http.Request, error) {
	var method = http.MethodPost
	if m.Type == CONNECT {
		method = http.MethodGet
	} else if m.Type == DISCONNECT {
		method = http.MethodDelete
	}

	var body *bytes.Buffer = bytes.NewBuffer(m.Data)
	var request, err = http.NewRequest(method, uri, body)
	if err != nil {
		var msg = fmt.Sprintf("error generating the request %v", err)
		return nil, errors.New(msg)
	}

	request.Header.Add(addresHeader, m.From.Address)
	request.Header.Add(portHeader, m.From.Port)
	return request, nil
}

// FromRequest function
func (m *Message) FromRequest(req *http.Request) *Message {
	if m.Type = PLAIN; req.Method == http.MethodGet {
		m.Type = CONNECT
	} else if req.Method == http.MethodDelete {
		m.Type = DISCONNECT
	}

	m.From = &Peer{
		Address: req.Header.Get(addresHeader),
		Port:    req.Header.Get(portHeader),
	}
	if m.Type == PLAIN {
		m.Data, _ = io.ReadAll(req.Body)
	}

	return m
}
