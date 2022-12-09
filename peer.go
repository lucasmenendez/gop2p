package gop2p

import (
	"encoding/json"
	"fmt"
	"net"
)

// baseHostname contains node address template
const baseHostname string = "http://%s:%s"

// baseString contains node address template
const baseString string = "http://%s:%s"

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

// isMe function compares current peer with other to check if both peers are
// equal.
func (me *Peer) IsMe(peer *Peer) bool {
	return peer.Address == me.Address && peer.Port == me.Port
}

func (peer *Peer) String() string {
	return fmt.Sprintf(baseString, peer.Address, peer.Port)
}

func (peer *Peer) Hostname() string {
	return fmt.Sprintf(baseHostname, peer.Address, peer.Port)
}

// Peers involves list of peer
type Peers []*Peer

// contains function returns if current list of peer contains other provided.
func (peers Peers) Contains(p *Peer) bool {
	for _, pn := range peers {
		if pn.Address == p.Address && pn.Port == p.Port {
			return true
		}
	}

	return false
}

// delete function returns a copy of current list of peer removing peer provided
// previously.
func (ps Peers) Delete(p *Peer) (r Peers) {
	for _, pn := range ps {
		if pn.Address != p.Address || pn.Port != p.Port {
			r = append(r, pn)
		}
	}

	return
}

func (peers Peers) ToJSON() ([]byte, error) {
	return json.Marshal(peers)
}

func PeersFromJSON(input []byte) (Peers, error) {
	var peers = Peers{}
	if err := json.Unmarshal(input, &peers); err != nil {
		return nil, err
	}

	return peers, nil
}
