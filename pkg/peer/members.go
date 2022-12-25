package peer

import (
	"encoding/json"
	"sync"
)

// Members struct abstracts a thread-safe list of network peers. It includes
// a slice of peers and a mutex to modify it safely. Both private arguments to
// keep the control of the data isolated on this package.
type Members struct {
	peers    []*Peer
	webChans map[string]chan []byte
	mutex    *sync.Mutex
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
func (m *Members) Peers() []*Peer {
	panicIfNotInitialized(m)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	var peers []*Peer
	peers = append(peers, m.peers...)
	return peers
}

// Len function returns the number of peers that current members contains
// safely (using the current members mutex).
func (m *Members) Len() int {
	panicIfNotInitialized(m)

	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.peers)
}

// Append function checks if the provided peer is already into the current
// members. If so, it returns the current members without any change. Unless,
// it append safely the provided peer to the current list and returns the
// modified current members, using its own mutex.
func (m *Members) Append(peer *Peer) *Members {
	panicIfNotInitialized(m)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	included := false
	for _, member := range m.peers {
		if member.Equal(peer) {
			included = true
			break
		}
	}

	if !included {
		m.peers = append(m.peers, peer)
	}

	return m
}

// Delete function removes the provided peer from the current members safely.
func (m *Members) Delete(peer *Peer) *Members {
	panicIfNotInitialized(m)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, member := range m.peers {
		if member.Equal(peer) {
			m.peers = append(m.peers[:i], m.peers[i+1:]...)
			return m
		}
	}

	return m
}

// Contains function checks if the provided peer is registered into the current
// members peer list safely.
func (m *Members) Contains(peer *Peer) bool {
	panicIfNotInitialized(m)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, member := range m.peers {
		if member.Equal(peer) {
			return true
		}
	}

	return false
}

// ToJSON function encodes the current list of network members into a JSON
// format and returns it as slice of bytes. If something was wrong, returns an
// error.
func (m *Members) ToJSON() ([]byte, error) {
	panicIfNotInitialized(m)

	m.mutex.Lock()
	defer m.mutex.Unlock()
	return json.Marshal(m.peers)
}

// FromJSON function parses the provided input as peer list and sets it to the
// current members peer list safely.
func (m *Members) FromJSON(input []byte) (*Members, error) {
	panicIfNotInitialized(m)

	peers := []*Peer{}
	if err := json.Unmarshal(input, &peers); err != nil {
		return nil, err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.peers = peers

	return m, nil
}
