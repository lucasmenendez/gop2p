// message package provide an abstraction of the data that is transfered by
// peers in the network. Currently three types of messages available:
// connection, disconnection and plain message. The package provide a group of
// functions associated to the messages for common tasks as message creation,
// serialization and deserialization.
package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/lucasmenendez/gop2p/pkg/peer"
)

const (
	ConnectType    = iota
	DisconnectType = iota
	PlainType      = iota
)

// addressHeader contains default http header key that contains address.
const addresHeader string = "PEER_ADDRESS"

// portHeader contains default http header key that contains port.
const portHeader string = "PEER_PORT"

// Message struct includes the content of a Message and it is transferred
// between peers. It contains its type as integer (checout the default types),
// the information about the message sender and the content of the message.
type Message struct {
	Type int
	Data []byte
	From *peer.Peer
}

// SetType function sets the type of the current message to the provided one,
// and returns this current message. By default, the type of a message will be
// PlainType, unless other valid type has been provided by argument.
func (m *Message) SetType(t int) *Message {
	m.Type = PlainType
	if t == ConnectType || t == DisconnectType {
		m.Type = t
	}

	return m
}

// SetFrom function sets the provided peer as the peer associated to the current
// message and returns it as result.
func (m *Message) SetFrom(peer *peer.Peer) *Message {
	m.From = peer
	return m
}

// SetData function sets the provided data as the data of the current message
// and returns it as result.
func (m *Message) SetData(data []byte) *Message {
	m.Type = PlainType
	m.Data = data
	return m
}

// String function returns a human-readable version of Message struct following
// the format: '[from.address:from.port] data'.
func (m *Message) String() string {
	return fmt.Sprintf("[%s] %s", m.From.String(), string(m.Data))
}

// GetRequest function generates a http.Request to the provided uri endpoint
// with the current message information, setting correct http.Method based on
// the message type, the message peer information as http.Header and message
// data as request body, then return it if everthing was ok, unless returns an
// error.
func (m *Message) GetRequest(uri string) (*http.Request, error) {
	// Set the correct method based on message type. The connection message
	// will be "GET" method, the disconnection message will be the "DELETE"
	// method and the plain message will be "POST".
	var method = http.MethodPost
	if m.Type == ConnectType {
		method = http.MethodGet
	} else if m.Type == DisconnectType {
		method = http.MethodDelete
	}

	// Create a buffer with the message data and creates the request with the
	// defined method.
	var body *bytes.Buffer = bytes.NewBuffer(m.Data)
	var request, err = http.NewRequest(method, uri, body)
	if err != nil {
		var msg = fmt.Sprintf("error generating the request %v", err)
		return nil, errors.New(msg)
	}

	// Set the message peer information as request headers.
	request.Header.Add(addresHeader, m.From.Address)
	request.Header.Add(portHeader, m.From.Port)
	return request, nil
}

// FromRequest function parses a http.Request provided and sets the information
// that it contains to the current message, then return the modified message
// too.
func (m *Message) FromRequest(req *http.Request) *Message {
	// Decodes the message by the method of the request, by default
	// PlainMessage.
	if m.Type = PlainType; req.Method == http.MethodGet {
		m.Type = ConnectType
	} else if req.Method == http.MethodDelete {
		m.Type = DisconnectType
	}

	// Decodes the peer information from the http.Header's.
	m.From = &peer.Peer{
		Address: req.Header.Get(addresHeader),
		Port:    req.Header.Get(portHeader),
	}

	// If the message type is PlainMessage, read the request body as message
	// data
	if m.Type == PlainType {
		m.Data, _ = io.ReadAll(req.Body)
	}

	return m
}
