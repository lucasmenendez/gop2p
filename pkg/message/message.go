// message package provide an abstraction of the data that is transferred by
// peers in the network. Currently three types of messages available:
// connection, disconnection, broadcas and direct message. The package provide a
// group of functions associated to the messages for common tasks as message
// creation, serialization and deserialization.
package message

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/lucasmenendez/gop2p/pkg/peer"
)

const (
	// ConnectType identifies a connection message that allows to connect to a
	// network
	ConnectType = iota
	// DisconnectType identifies a disconnection message that report a peer
	// disconnection to the current network peers
	DisconnectType = iota
	// BroadcastType identifies a message for the entire network
	BroadcastType = iota
	// DirectType identifies a message that is intended for a single network
	// peer (such as direct message).
	DirectType = iota
)

const (
	// addressHeader contains the default http header key to send the peer
	// address through a network request.
	addressHeader string = "X-PEER_ADDRESS"
	// portHeader contains the default http header key to send the peer port
	// through a network request.
	portHeader string = "X-PEER_PORT"
	// typeHeader contains the default http header key to send the peer type
	// through a network request.
	typeHeader    string = "X-PEER_TYPE"
	fromParameter string = "from"
)

// Message struct includes the content of a Message and it is transferred
// between peers. It contains its type as integer (checkout defined types),
// the information about the message sender and the content of the message.
type Message struct {
	Type int
	Data []byte
	From *peer.Peer
	To   *peer.Peer
}

// SetType function sets the type of the current message to the provided one,
// and returns this current message. By default, the type of a message will be
// BroadcastType, unless other valid type has been provided by argument.
func (msg *Message) SetType(t int) *Message {
	msg.Type = BroadcastType
	if t == ConnectType || t == DisconnectType || t == DirectType {
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

// SetTo function sets the provided peer as the peer for which the current
// message is intended and returns it as result.
func (msg *Message) SetTo(to *peer.Peer) *Message {
	msg.Type = DirectType
	msg.To = to
	return msg
}

// SetData function sets the provided data as the data of the current message
// and returns it as result.
func (msg *Message) SetData(data []byte) *Message {
	if msg.Type != BroadcastType && msg.Type != DirectType {
		msg.Type = BroadcastType
	}
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
	// method, the broadcast message will be "POST" and the direct message will
	// be "PUT".
	method := http.MethodPost
	if msg.Type == ConnectType {
		method = http.MethodGet
	} else if msg.Type == DisconnectType {
		method = http.MethodDelete
	} else if msg.Type == DirectType {
		method = http.MethodPut
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
	request.Header.Add(typeHeader, msg.From.Type)
	request.Host = msg.From.String()
	return request, nil
}

// FromRequest function parses a http.Request provided and sets the information
// that it contains to the current message, then return the modified message
// too.
func (msg *Message) FromRequest(req *http.Request) *Message {
	// Decodes the message by the method of the request, by default
	// BroadcastType.
	if msg.Type = BroadcastType; req.Method == http.MethodGet {
		msg.Type = ConnectType
	} else if req.Method == http.MethodDelete {
		msg.Type = DisconnectType
	} else if req.Method == http.MethodPut {
		msg.Type = DirectType
	}

	// Decodes the peer port and address information from the http.Header's or
	// the request.Host information.
	fromAddress := req.Header.Get(addressHeader)
	portValue := req.Header.Get(portHeader)
	if fromAddress == "" || portValue == "" {
		host := req.URL.Query().Get(fromParameter)
		if host == "" && req.Host == "" {
			return nil
		} else if host == "" {
			host = req.Host
		}

		var err error
		fromAddress, portValue, err = net.SplitHostPort(host)
		if err != nil {
			return nil
		}
	}

	// Try to create a valid peer with the decoded info and set the peer type.
	if fromPort, err := strconv.Atoi(portValue); err != nil {
		return nil
	} else if msg.From, err = peer.New(fromAddress, fromPort); err != nil {
		return nil
	} else if fromType := req.Header.Get(typeHeader); fromType == peer.TypeWeb {
		msg.From.Type = peer.TypeWeb
	} else if req.Header.Get("Connection") == "keep-alive" {
		msg.From.Type = peer.TypeWeb
	}

	// If the message type is BroadcastType or DirectType, read the request body
	// as message data
	if msg.Type == BroadcastType || msg.Type == DirectType {
		msg.Data, _ = io.ReadAll(req.Body)
	}

	return msg
}
