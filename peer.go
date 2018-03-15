package gop2p

import (
	"fmt"
	"log"
	"net"
)

// msg struct contains message information: emitter peer and content.
type msg struct {
	From    peer   `json:"from"`
	Content string `json:"content"`
}

// toMap function returns a structured map with message information formatted.
func (m msg) toMap() map[string]interface{} {
	return map[string]interface{}{
		"from": map[string]interface{}{
			"address": m.From.Address,
			"alias":   m.From.Alias,
			"port":    m.From.Port,
		},
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
	return i
}

// Me function involves CreatePeer function getting current host ip address
// previously.
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

// isMe function compare current peer with other to check if both peers are
// equal.
func (p peer) isMe(c peer) bool {
	return p.Alias == c.Alias && p.Address == c.Address && p.Port == c.Port
}

// log function logs message provided formated and adding peer information trace.
func (p peer) log(m string, args ...interface{}) {
	m = fmt.Sprintf(m, args...)
	log.Printf("[%s:%d](%s) - %s\n", p.Address, p.Port, p.Alias, m)
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

	return r
}
