package peer

import (
	"fmt"
	"net"
)

// baseHostname contains node address template
const baseHostname string = "http://%s:%s"

// baseString contains node address template
const baseString string = "%s:%s"

// Peer struct contains peer ip address and port to communicate with ir.
type Peer struct {
	Port    string `json:"port"`
	Address string `json:"address"`
}

// Me function involves Createpeer function getting current host ip address
// previously.
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

// Equal function
func (peer *Peer) Equal(to *Peer) bool {
	return peer.Address == to.Address && peer.Port == to.Port
}

// String function
func (peer *Peer) String() string {
	return fmt.Sprintf(baseString, peer.Address, peer.Port)
}

// Hostname function
func (peer *Peer) Hostname() string {
	return fmt.Sprintf(baseHostname, peer.Address, peer.Port)
}
