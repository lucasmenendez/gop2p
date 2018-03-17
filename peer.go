package gop2p

import (
	"net"
	"fmt"
)

// msg struct contains message information: emitter peer and content.
type msg struct {
	From    peer   `json:"from"`
	Content string `json:"content"`
}

// toMap function returns a structured map with message information formatted.
func (m msg) toMap() map[string]interface{} {
	return map[string]interface{}{
		"from": m.From.toMap(),
		"content": m.Content,
	}
}

// peer struct contains peer alias, ip address and port to communicate with.
type peer struct {
	Alias   string `json:"alias"`
	Port    int    `json:"port"`
	Address string `json:"address"`
}

// CreatePeer function returns defined peer based on peer alias, ip address and
// port provided.
func CreatePeer(a, n string, p int) (i peer) {
	i.Address = a
	i.Port = p
	i.Alias = n
	return
}

// Me function involves CreatePeer function getting current host ip address
// previously.
func Me(n string, p int) (me peer) {
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

// isMe function compare current peer with other to check if both peers are
// equal.
func (p peer) isMe(c peer) bool {
	return p.Alias == c.Alias && p.Address == c.Address && p.Port == c.Port
}

// toMap function returns a structured map with peer information formatted.
func (p peer) toMap() map[string]interface{} {
	return map[string]interface{}{
		"address": p.Address,
		"alias":   p.Alias,
		"port":    fmt.Sprintf("%d", p.Port),
	}
}

// peers involves list of peer
type peers []peer

// contains function return if current list of peer contains other provided.
func (ps peers) contains(p peer) bool {
	for _, pn := range ps {
		if pn.Address == p.Address && pn.Alias == p.Alias {
			return true
		}
	}

	return false
}

// delete function returns a copy of current list of peer removing peer provided
// previously.
func (ps peers) delete(p peer) (r peers) {
	for _, pn := range ps {
		if pn.Address != p.Address || pn.Alias != p.Alias {
			r = append(r, pn)
		}
	}

	return
}
