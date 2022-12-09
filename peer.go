package gop2p

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

// baseURI contains node address template
const baseURI string = "http://%s:%s"

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

// isMe function compares current peer with other to check if both peers are
// equal.
func (p Peer) IsMe(c Peer) bool {
	return p.Address == c.Address && p.Port == c.Port
}

// toByte function returns serialized json with peer information.
func (p Peer) ToBytes() (d []byte) {
	d, _ = json.Marshal(&p)
	return
}

func (p Peer) String() string {
	return p.Address + ":" + p.Port
}

func (p Peer) URI() string {
	return fmt.Sprintf(baseURI, p.Address, p.Port)
}

// FromBytes function returns deserialized peer from json.
func FromBytes(d []byte) (p Peer) {
	_ = json.Unmarshal(d, &p)
	return
}

// Peers involves list of peer
type Peers []Peer

// contains function returns if current list of peer contains other provided.
func (ps Peers) Contains(p Peer) bool {
	for _, pn := range ps {
		if pn.Address == p.Address && pn.Port == p.Port {
			return true
		}
	}

	return false
}

// delete function returns a copy of current list of peer removing peer provided
// previously.
func (ps Peers) Delete(p Peer) (r Peers) {
	for _, pn := range ps {
		if pn.Address != p.Address || pn.Port != p.Port {
			r = append(r, pn)
		}
	}

	return
}

func (ps Peers) ToJSON() ([]byte, error) {
	return json.Marshal(ps)
}

func PeersFromJSON(input []byte) (Peers, error) {
	var records = []map[string]string{}
	if err := json.Unmarshal(input, &records); err != nil {
		return nil, err
	}

	var peers = Peers{}
	for _, record := range records {
		var peer = Peer{
			Address: record["address"],
			Port:    record["port"],
		}
		peers = append(peers, peer)
	}
	return peers, nil
}
