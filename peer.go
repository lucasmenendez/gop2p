package gop2p

import (
	"encoding/json"
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

// String function
func (peer *Peer) String() string {
	return fmt.Sprintf(baseString, peer.Address, peer.Port)
}

// Hostname function
func (peer *Peer) Hostname() string {
	return fmt.Sprintf(baseHostname, peer.Address, peer.Port)
}

// Peers involves list of peer
type Peers []*Peer

// Contains function returns if current list of peer contains other provided.
func (peers Peers) Contains(p *Peer) bool {
	for _, pn := range peers {
		if pn.Address == p.Address && pn.Port == p.Port {
			return true
		}
	}

	return false
}

// Delete function returns a copy of current list of peer removing peer provided
// previously.
func (peers *Peers) Delete(p *Peer) *Peers {
	var result = Peers{}
	for _, pn := range *peers {
		if pn.Address != p.Address || pn.Port != p.Port {
			result = append(result, pn)
		}
	}

	peers = &result
	return peers
}

// ToJSON function
func (peers *Peers) ToJSON() ([]byte, error) {
	return json.Marshal(peers)
}

// FromJSON
func (peers *Peers) FromJSON(input []byte) (*Peers, error) {
	peers = &Peers{}
	if err := json.Unmarshal(input, peers); err != nil {
		return nil, err
	}

	return peers, nil
}
