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

// panicIfNotInitialized function calls panic if the provided Members is not
// initialized with a empty slice of *Peer's and an initialized mutex to protect
// the access to it.
func panicIfNotInitialized(members *Members) {
	if members.mutex == nil || members.peers == nil {
		panic("current Members struct instance not initialized, use NewMembers() function")
	}
}

// NewMembers function intializes a new Members struct and return it.
func NewMembers() *Members {
	return &Members{
		peers: []*Peer{},
		mutex: &sync.Mutex{},
	}
}

// Peers function returns a copy of the list of peers safely ot the current
// members, using the mutex associated to it.
func (members *Members) Peers() []*Peer {
	panicIfNotInitialized(members)

	members.mutex.Lock()
	defer members.mutex.Unlock()

	var peers []*Peer
	peers = append(peers, members.peers...)
	return peers
}

// Len function returns the number of peers that current members contains
// safely (using the current members mutex).
func (members *Members) Len() int {
	panicIfNotInitialized(members)

	members.mutex.Lock()
	defer members.mutex.Unlock()
	return len(members.peers)
}

// Append function checks if the provided peer is already into the current
// members. If so, it returns the current members without any change. Unless,
// it append safely the provided peer to the current list and returns the
// modified current members, using its own mutex.
func (members *Members) Append(peer *Peer) *Members {
	panicIfNotInitialized(members)

	members.mutex.Lock()
	defer members.mutex.Unlock()

	included := false
	for _, member := range members.peers {
		if member.Equal(peer) {
			included = true
			break
		}
	}

	if !included {
		members.peers = append(members.peers, peer)
	}

	return members
}

// Delete function removes the provided peer from the current members safely.
func (members *Members) Delete(peer *Peer) *Members {
	panicIfNotInitialized(members)

	members.mutex.Lock()
	defer members.mutex.Unlock()

	for i, member := range members.peers {
		if member.Equal(peer) {
			members.peers = append(members.peers[:i], members.peers[i+1:]...)
			return members
		}
	}

	return members
}

// Contains function checks if the provided peer is registered into the current
// members peer list safely.
func (members *Members) Contains(peer *Peer) bool {
	panicIfNotInitialized(members)

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
	panicIfNotInitialized(members)

	members.mutex.Lock()
	defer members.mutex.Unlock()
	return json.Marshal(members.peers)
}

// FromJSON function parses the provided input as peer list and sets it to the
// current members peer list safely.
func (members *Members) FromJSON(input []byte) (*Members, error) {
	panicIfNotInitialized(members)

	peers := []*Peer{}
	if err := json.Unmarshal(input, &peers); err != nil {
		return nil, err
	}

	members.mutex.Lock()
	defer members.mutex.Unlock()
	members.peers = peers

	return members, nil
}
