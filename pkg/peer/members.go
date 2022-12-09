package peer

import (
	"encoding/json"
	"sync"
)

// Members struct abstracts a thread-safe list of network peers. It includes
// a slice of peers and a mutex to modify it safely. Both private arguments to
// keep the control of the data isolated on this package.
type Members struct {
	peers []*Peer
	mutex *sync.Mutex
}

// EmptyMembers function intializes a new Members struct and return it.
func EmptyMembers() *Members {
	return &Members{
		peers: []*Peer{},
		mutex: &sync.Mutex{},
	}
}

// Peers function returns a copy of the list of peers safely ot the current
// members, using the mutex associated to it.
func (members *Members) Peers() []*Peer {
	members.mutex.Lock()
	defer members.mutex.Unlock()

	var peers []*Peer
	peers = append(peers, members.peers...)
	return peers
}

// Len function returns the number of peers that current members contains
// safely (using the current members mutex).
func (members *Members) Len() int {
	members.mutex.Lock()
	defer members.mutex.Unlock()
	return len(members.peers)
}

// Append function checks if the provided peer is already into the current
// members. If so, it returns the current members without any change. Unless,
// it append safely the provided peer to the current list and returns the
// modified current members, using its own mutex.
func (members *Members) Append(peer *Peer) *Members {
	if !members.Contains(peer) {
		members.mutex.Lock()
		members.peers = append(members.peers, peer)
		members.mutex.Unlock()
	}
	return members
}

// Delete function removes the provided peer from the current members safely.
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

// Contains function checks if the provided peer is registered into the current
// members peer list safely.
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

// ToJSON function encodes the current list of network members into a JSON
// format and returns it as slice of bytes. If something was wrong, returns an
// error.
func (members *Members) ToJSON() ([]byte, error) {
	members.mutex.Lock()
	defer members.mutex.Unlock()

	var res, err = json.Marshal(members.peers)
	return res, err
}

// FromJSON function parses the provided input as peer list and sets it to the
// current members peer list safely.
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
