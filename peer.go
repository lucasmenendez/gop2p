package gop2p

import (
	"fmt"
	"net"
)

// Message struct contains message information: emitter Peer and Content.
type Message struct {
	From    Peer   `json:"From"`
	Content string `json:"Content"`
}

// toMap function returns a structured map with message information formatted.
func (m Message) toMap() map[string]interface{} {
	return map[string]interface{}{
		"From":    m.From.toMap(),
		"Content": m.Content,
	}
}

// Peer struct contains Peer alias, ip address and port to communicate with.
type Peer struct {
	Alias   string `json:"alias"`
	Port    int    `json:"port"`
	Address string `json:"address"`
}

// CreatePeer function returns defined Peer based on Peer alias, ip address and
// port provided.
func CreatePeer(a, n string, p int) (i Peer) {
	i.Address = a
	i.Port = p
	i.Alias = n
	return
}

// Me function involves CreatePeer function getting current host ip address
// previously.
func Me(n string, p int) (me Peer) {
	me = CreatePeer("localhost", n, p)

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

// isMe function compare current Peer with other to check if both peers are
// equal.
func (p Peer) isMe(c Peer) bool {
	return p.Alias == c.Alias && p.Address == c.Address && p.Port == c.Port
}

// toMap function returns a structured map with Peer information formatted.
func (p Peer) toMap() map[string]interface{} {
	return map[string]interface{}{
		"address": p.Address,
		"alias":   p.Alias,
		"port":    fmt.Sprintf("%data", p.Port),
	}
}

// peers involves list of Peer
type peers []Peer

// contains function return if current list of Peer contains other provided.
func (ps peers) contains(p Peer) bool {
	for _, pn := range ps {
		if pn.Address == p.Address && pn.Alias == p.Alias {
			return true
		}
	}

	return false
}

// delete function returns a copy of current list of Peer removing Peer provided
// previously.
func (ps peers) delete(p Peer) (r peers) {
	for _, pn := range ps {
		if pn.Address != p.Address || pn.Alias != p.Alias {
			r = append(r, pn)
		}
	}

	return
}
