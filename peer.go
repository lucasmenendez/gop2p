package gop2p

import (
	"fmt"
	"log"
	"net"
)

type msg struct {
	From    peer   `json:"From"`
	Content string `json:"Content"`
}

type peer struct {
	Alias   string `json:"alias"`
	Port    int    `json:"port"`
	Address string `json:"address"`
}

type peers []peer

func CreatePeer(a, n string, p int) (i peer) {
	i.Address = a
	i.Port = p
	i.Alias = n
	return i
}

func Me(n string, p int) (me peer) {
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

func (p peer) isMe(c peer) bool {
	return p.Alias == c.Alias && p.Address == c.Address && p.Port == c.Port
}

func (p peer) log(m string, args ...interface{}) {
	m = fmt.Sprintf(m, args...)
	log.Printf("[%s:%d](%s) - %s\n", p.Address, p.Port, p.Alias, m)
}

func (ps peers) contains(p peer) bool {
	for _, pn := range ps {
		if pn.Address == p.Address && pn.Alias == p.Alias {
			return true
		}
	}

	return false
}

func (ps peers) delete(p peer) (r peers) {
	for _, pn := range ps {
		if pn.Address != p.Address || pn.Alias != p.Alias {
			r = append(r, pn)
		}
	}

	return r
}
