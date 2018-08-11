package gop2p

import (
	"encoding/json"
	"net"
	"strconv"
)

// Peer struct contains peer ip address and port to communicate with ir.
type Peer struct {
	Port    string `json:"port"`
	Address string `json:"address"`
}

// CreatePeer function returns manually created peer based on ip address and
// port provided.
func CreatePeer(a string, p int) (i Peer) {
	i.Address = a
	i.Port = strconv.Itoa(p)
	return
}

// Me function involves Createpeer function getting current host ip address
// previously.
func Me(p int) (me Peer) {
	me = CreatePeer("localhost", p)

	var e error
	var addrs []net.Addr
	if addrs, e = net.InterfaceAddrs(); e != nil {
		return
	}

	var a string
	for _, an := range addrs {
		if ip, ok := an.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				a = ip.IP.String()
				break
			}
		}
	}

	if a != "" {
		me.Address = a
	}

	return
}

// isMe function compare current peer with other to check if both peers are
// equal.
func (p Peer) isMe(c Peer) bool {
	return p.Address == c.Address && p.Port == c.Port
}

func (p Peer) toBytes() (d []byte) {
	d, _ = json.Marshal(&p)
	return
}

// Peers involves list of peer
type Peers []Peer

// contains function return if current list of peer contains other provided.
func (ps Peers) contains(p Peer) bool {
	for _, pn := range ps {
		if pn.Address == p.Address && pn.Port == p.Port {
			return true
		}
	}

	return false
}

// delete function returns a copy of current list of peer removing peer provided
// previously.
func (ps Peers) delete(p Peer) (r Peers) {
	for _, pn := range ps {
		if pn.Address != p.Address || pn.Port != p.Port {
			r = append(r, pn)
		}
	}

	return
}
