// peer package abstracts the logic of two structs: the Peer, that
// contains peer address and port, information that identifies any node and
// allows to others to communicate with it; and the Members, a thread-safe list
// of network peers.
package peer

import (
	"fmt"
	"net"
	"net/url"
)

const (
	// baseHostname contains node address template
	baseHostname string = "http://%s:%d"
	// baseString contains node address template
	baseString string = "%s:%d"
	// allAddresses contains the wildcard IP as string
	allAddresses string = "0.0.0.0"
)

var (
	ErrBadAddress  = fmt.Errorf("bad peer address provided")
	ErrPortAddress = fmt.Errorf("bad peer port provided")
)

// Peer struct contains peer address and port, information that identifies any
// node and allows to others to communicate with it.
type Peer struct {
	Port    int    `json:"port"`
	Address string `json:"address"`
}

// New function creates a peer with the provided address and port as argument
// and returns it.
func New(address string, port int) (*Peer, error) {
	if address == "" {
		return nil, ErrPortAddress
	} else if port < 1 || port > 65535 {
		return nil, ErrPortAddress
	}

	peer := &Peer{
		Address: address,
		Port:    port,
	}

	if _, err := url.Parse(peer.Hostname()); err != nil {
		return nil, ErrBadAddress
	}
	return peer, nil
}

// Me function creates and returns a new peer with the current host address and
// the port provided as input. The function also receives a boolean as a second
// parameter 'remote' that indicates if the peer is a local peer (with local IP
// or 'localhost' as Address) or a remote one (with 0.0.0.0 IP as Address)
func Me(port int, remote bool) (*Peer, error) {
	if remote {
		return New(allAddresses, port)
	}

	me, err := New("localhost", port)
	if err != nil {
		return nil, err
	}

	if addresses, err := net.InterfaceAddrs(); err != nil {
		for _, address := range addresses {
			ip, ok := address.(*net.IPNet)
			if ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
				me.Address = ip.IP.String()
				break
			}
		}
	}

	return me, nil
}

// Equal function returns if the current peer is the same that the provided one.
// It seems that both has the same address and port.
func (p *Peer) Equal(to *Peer) bool {
	return p.Address == to.Address && p.Port == to.Port
}

// String function returns a human-readable format of the current peer.
func (p *Peer) String() string {
	return fmt.Sprintf(baseString, p.Address, p.Port)
}

// Hostname function returns the current peer information as URL form.
func (p *Peer) Hostname() string {
	return fmt.Sprintf(baseHostname, p.Address, p.Port)
}
