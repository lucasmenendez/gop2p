// peer package abstracts the logic of two structs: the Peer, that
// contains peer address and port, information that identifies any node and
// allows to others to communicate with it; and the Members, a thread-safe list
// of network peers.
package peer

import (
	"fmt"
	"net"
)

// baseHostname contains node address template
const baseHostname string = "http://%s:%s"

// baseString contains node address template
const baseString string = "%s:%s"

// Peer struct contains peer address and port, information that identifies any
// node and allows to others to communicate with it.
type Peer struct {
	Port    string `json:"port"`
	Address string `json:"address"`
}

// Me function creates and returns a new peer with the current host address and
// the port provided as input.
func Me(port int) (me *Peer) {
	me = &Peer{
		Address: "localhost",
		Port:    fmt.Sprint(port),
	}

	var addresses, err = net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, address := range addresses {
		if ip, ok := address.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				me.Address = ip.IP.String()
				break
			}
		}
	}

	return
}

// Equal function returns if the current peer is the same that the provided one.
// It seems that both has the same address and port.
func (peer *Peer) Equal(to *Peer) bool {
	return peer.Address == to.Address && peer.Port == to.Port
}

// String function returns a human-readable format of the current peer.
func (peer *Peer) String() string {
	return fmt.Sprintf(baseString, peer.Address, peer.Port)
}

// Hostname function returns the current peer information as URL form.
func (peer *Peer) Hostname() string {
	return fmt.Sprintf(baseHostname, peer.Address, peer.Port)
}
