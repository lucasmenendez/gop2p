package gop2p

import (
	"net"
	"strconv"
)

// peer struct contains peer alias, ip address and port to communicate with.
type peer struct {
	port    string
	address string
}

// Createpeer function returns defined peer based on peer alias, ip address and
// port provided.
func Createpeer(a string, p int) (i peer) {
	i.address = a
	i.port = strconv.Itoa(p)
	return
}

// Me function involves Createpeer function getting current host ip address
// previously.
func Me(p int) (me peer) {
	me = Createpeer("localhost", p)

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
		me.address = a
	}

	return
}

// isMe function compare current peer with other to check if both peers are
// equal.
func (p peer) isMe(c peer) bool {
	return p.address == c.address && p.port == c.port
}

// peers involves list of peer
type peers []peer

// contains function return if current list of peer contains other provided.
func (ps peers) contains(p peer) bool {
	for _, pn := range ps {
		if pn.address == p.address && pn.port == p.port {
			return true
		}
	}

	return false
}

// delete function returns a copy of current list of peer removing peer provided
// previously.
func (ps peers) delete(p peer) (r peers) {
	for _, pn := range ps {
		if pn.address != p.address || pn.port != p.port {
			r = append(r, pn)
		}
	}

	return
}
