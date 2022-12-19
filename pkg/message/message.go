// message package provide an abstraction of the data that is transferred by
// peers in the network. Currently three types of messages available:
// connection, disconnection and plain message. The package provide a group of
// functions associated to the messages for common tasks as message creation,
// serialization and deserialization.
package message

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/lucasmenendez/gop2p/pkg/peer"
)

const (
	ConnectType    = iota
	DisconnectType = iota
	PlainType      = iota
)

// addressHeader contains default http header key that contains address.
const addressHeader string = "PEER_ADDRESS"

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
func (msg *Message) SetType(t int) *Message {
	msg.Type = PlainType
	if t == ConnectType || t == DisconnectType {
		msg.Type = t
	}

	return msg
}

// SetFrom function sets the provided peer as the peer associated to the current
// message and returns it as result.
func (msg *Message) SetFrom(from *peer.Peer) *Message {
	msg.From = from
	return msg
}

// SetData function sets the provided data as the data of the current message
// and returns it as result.
func (msg *Message) SetData(data []byte) *Message {
	msg.Type = PlainType
	msg.Data = data
	return msg
}

// String function returns a human-readable version of Message struct following
// the format: '[from.address:from.port] data'.
func (msg *Message) String() string {
	return fmt.Sprintf("[%s] %s", msg.From.String(), string(msg.Data))
}

// GetRequest function generates a http.Request to the provided uri endpoint
// with the current message information, setting correct http.Method based on
// the message type, the message peer information as http.Header and message
// data as request body, then return it if everything was ok, unless returns an
// error.
func (msg *Message) GetRequest(uri string) (*http.Request, error) {
	if msg.From == nil || msg.From.Address == "" || msg.From.Port == 0 {
		return nil, fmt.Errorf("current message have not peer associated")
	}

	// Set the correct method based on message type. The connection message
	// will be "GET" method, the disconnection message will be the "DELETE"
	// method and the plain message will be "POST".
	method := http.MethodPost
	if msg.Type == ConnectType {
		method = http.MethodGet
	} else if msg.Type == DisconnectType {
		method = http.MethodDelete
	}

	// Create a buffer with the message data and creates the request with the
	// defined method.
	body := bytes.NewBuffer(msg.Data)
	request, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, fmt.Errorf("error generating the request %v", err)
	}

	// Set the message peer information as request headers.
	request.Header.Add(addressHeader, msg.From.Address)
	request.Header.Add(portHeader, fmt.Sprint(msg.From.Port))
	return request, nil
}

// FromRequest function parses a http.Request provided and sets the information
// that it contains to the current message, then return the modified message
// too.
func (msg *Message) FromRequest(req *http.Request) *Message {
	// Decodes the message by the method of the request, by default
	// PlainMessage.
	if msg.Type = PlainType; req.Method == http.MethodGet {
		msg.Type = ConnectType
	} else if req.Method == http.MethodDelete {
		msg.Type = DisconnectType
	}

	// Decodes the peer information from the http.Header's.
	msg.From = &peer.Peer{}
	if msg.From.Address = req.Header.Get(addressHeader); msg.From.Address == "" {
		return nil
	} else if portValue := req.Header.Get(portHeader); portValue != "" {
		var err error
		if msg.From.Port, err = strconv.Atoi(portValue); err != nil {
			return nil
		}
	}

	// If the message type is PlainMessage, read the request body as message
	// data
	if msg.Type == PlainType {
		msg.Data, _ = io.ReadAll(req.Body)
	}

	return msg
}
