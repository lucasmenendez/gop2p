package peer

import (
	"encoding/json"
	"sync"
)

type Members struct {
	peers []*Peer
	mutex *sync.Mutex
}

func EmptyMembers() *Members {
	return &Members{
		peers: []*Peer{},
		mutex: &sync.Mutex{},
	}
}

func (members *Members) Peers() []*Peer {
	members.mutex.Lock()
	defer members.mutex.Unlock()

	var peers []*Peer
	peers = append(peers, members.peers...)
	return peers
}

func (members *Members) Append(peer *Peer) *Members {
	if !members.Contains(peer) {
		members.mutex.Lock()
		members.peers = append(members.peers, peer)
		members.mutex.Unlock()
	}
	return members
}

func (members *Members) Delete(peer *Peer) *Members {
	members.mutex.Lock()
	defer members.mutex.Unlock()

	var newMembers = []*Peer{}
	for _, member := range members.peers {
		if !member.Equal(peer) {
			newMembers = append(newMembers, member)
		}
	}

	members.peers = newMembers
	return members
}

func (members *Members) Len() int {
	members.mutex.Lock()
	defer members.mutex.Unlock()
	return len(members.peers)
}

func (members *Members) Contains(peer *Peer) bool {
	members.mutex.Lock()
	defer members.mutex.Unlock()

	for _, member := range members.peers {
		if member.Equal(peer) {
			return true
		}
	}

	return false
}

func (members *Members) ToJSON() ([]byte, error) {
	members.mutex.Lock()
	defer members.mutex.Unlock()

	var res, err = json.Marshal(members.peers)
	return res, err
}

func (members *Members) FromJSON(input []byte) (*Members, error) {
	var peers = []*Peer{}
	if err := json.Unmarshal(input, &peers); err != nil {
		return nil, err
	}

	members.mutex.Lock()
	defer members.mutex.Unlock()
	members.peers = peers

	return members, nil
}
