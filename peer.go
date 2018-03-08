package p2p

import (
	"os"
	"net"
	"log"
	"fmt"
)

type Peer struct {
	Alias   string
	Address string
}

func CreatePeer(a, n string) (p Peer) {
	p.Address = a
	p.Alias = n

	p.log("Peer created.")
	return p
}

func Me(n string) (p Peer) {
	var e error
	var host string
	if host, e = os.Hostname(); e != nil {
		return CreatePeer("localhost", n)
	}

	var addrs []net.IP
	if addrs, e = net.LookupIP(host); e != nil {
		return CreatePeer("localhost", n)
	}

	var a string
	for _, an := range addrs {
		if ip := an.To4(); ip != nil {
			a = string(ip)
			break
		}
	}

	if a == "" {
		return CreatePeer("localhost", n)
	}

	return CreatePeer(a, n)
}

func (p Peer) Info() {
	p.log("Here I am!")
}

func (p Peer) isMe(c Peer) bool {
	return p.Alias == c.Alias && p.Address == c.Alias
}

func (p Peer) log(m string, args ...interface{}) {
	m = fmt.Sprintf(m, args...)
	log.Printf("[%s](%s) - %s\n", p.Address, p.Alias, m)
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