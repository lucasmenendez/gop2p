package p2p

import (
	"fmt"
	"log"
	"net"
)

type Peer struct {
	Alias   string
	Port    int
	Address string
}

func CreatePeer(a, n string, p int) (i Peer) {
	i.Address = a
	i.Port = p
	i.Alias = n

	i.log("Peer created.")
	return i
}

func Me(n string, p int) (me Peer) {
	var e error
	var addrs []net.Addr
	if addrs, e = net.InterfaceAddrs(); e != nil {
		return CreatePeer("localhost", n, p)
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

	if a == "" {
		return CreatePeer("localhost", n, p)
	}

	return CreatePeer(a, n, p)
}

func (p Peer) Info() {
	p.log("Here I am!")
}

func (p Peer) isMe(c Peer) bool {
	return p.Alias == c.Alias && p.Address == c.Alias
}

func (p Peer) log(m string, args ...interface{}) {
	m = fmt.Sprintf(m, args...)
	log.Printf("[%s:%d](%s) - %s\n", p.Address, p.Port, p.Alias, m)
}

type msg struct {
	from    Peer
	content string
}

type peers []Peer

func (ps peers) contains(p Peer) bool {
	for _, pn := range ps {
		if pn.Address == p.Address && pn.Alias == p.Alias {
			return true
		}
	}

	return false
}

func (ps peers) delete(p Peer) (r peers) {
	for _, pn := range ps {
		if pn.Address != p.Address || pn.Alias != p.Alias {
			r = append(r, pn)
		}
	}

	return r
}
