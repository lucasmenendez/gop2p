package peer

import (
	"reflect"
	"sync"
	"testing"
)

func TestNotInitializedMembers(t *testing.T) {
	defer func() {
		if except := recover(); except == nil {
			t.Error("expected panic, got error")
		}
	}()

	// not panic
	var result = NewMembers()
	result.Len()

	// panic
	result = new(Members)
	result.Len()
}

func TestNewMembers(t *testing.T) {
	var expected, result = &Members{[]*Peer{}, &sync.Mutex{}}, NewMembers()
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestMembersPeers(t *testing.T) {
	var members = NewMembers()
	if result := members.Peers(); len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}

	var expected = []*Peer{Me(5000), Me(5001), Me(5002)}
	for _, member := range expected {
		members.Append(member)
	}

	if !reflect.DeepEqual(expected, members.Peers()) {
		t.Errorf("expected %v, got %v", expected, members.Peers())
	}
}

func TestMembersLen(t *testing.T) {
	var result = NewMembers()
	if result.Len() != 0 {
		t.Errorf("expected 0, got %d", result.Len())
	}

	var expected = []*Peer{Me(5000), Me(5001), Me(5002)}
	for _, member := range expected {
		result.Append(member)
	}

	if len(expected) != result.Len() {
		t.Errorf("expected %d, got %d", len(expected), result.Len())
	}

	var needle = expected[0]
	result.Delete(needle)
	if expected = expected[1:]; len(expected) != result.Len() {
		t.Errorf("expected %d, got %d", len(expected), result.Len())
	}
}

func TestMembersAppend(t *testing.T) {
	var result = NewMembers()
	var expected = []*Peer{Me(5000), Me(5001), Me(5002)}
	for i, member := range expected {
		result.Append(member)

		if !reflect.DeepEqual(member, result.peers[i]) {
			t.Errorf("expected %v at index %d, got %v", member, i, result.peers[i])
		}
	}

	if len(expected) != result.Len() {
		t.Errorf("expected %d, got %d", len(expected), result.Len())
	}
}

func TestMembersDelete(t *testing.T) {
	var result = NewMembers()
	var expected = []*Peer{Me(5000), Me(5001), Me(5002)}
	for _, member := range expected {
		result.Append(member)
	}

	result.Delete(expected[0])
	if expected = expected[1:]; !reflect.DeepEqual(expected, result.peers) {
		t.Errorf("expected %v, got %v", expected, result.peers)
	}
}

func TestMembersContains(t *testing.T) {
	var result = NewMembers()
	var expected = []*Peer{Me(5000), Me(5001), Me(5002)}
	for _, member := range expected {
		result.Append(member)
	}

	if needle := expected[0]; !result.Contains(needle) {
		t.Error("expected true, got false")
	} else if needle = Me(5003); result.Contains(needle) {
		t.Error("expected false, got true")
	}
}

func TestMembersToJSON(t *testing.T) {
	var examples = []*Peer{Me(5000), Me(5001), Me(5002)}
	var address = examples[0].Address
	var expected = []byte("[{\"port\":\"5000\",\"address\":\"" + address + "\"}]")
	var members = NewMembers()
	members.Append(examples[0])

	if result, err := members.ToJSON(); err != nil {
		t.Errorf("expected nil, got %v", err)
	} else if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected %s, got %s", expected, result)
	}

	for _, example := range examples[1:] {
		members.Append(example)
	}

	expected = []byte("[{\"port\":\"5000\",\"address\":\"" + address + "\"},{\"port\":\"5001\",\"address\":\"" + address + "\"},{\"port\":\"5002\",\"address\":\"" + address + "\"}]")
	if result, err := members.ToJSON(); err != nil {
		t.Errorf("expected nil, got %v", err)
	} else if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestMembersFromJSON(t *testing.T) {
	var examples = []*Peer{Me(5000), Me(5001), Me(5002)}
	var address = examples[0].Address
	var example = []byte("[{\"port\":\"5000\",\"address\":\"" + address + "\"}]")
	var expected = NewMembers()
	expected.Append(examples[0])

	if result, err := expected.FromJSON(example); err != nil {
		t.Errorf("expected nil, got %v", err)
	} else if !reflect.DeepEqual(expected.peers, result.peers) {
		t.Errorf("expected %v, got %v", expected.peers, result.peers)
	}

	for _, p := range examples[1:] {
		expected.Append(p)
	}

	example = []byte("[{\"port\":\"5000\",\"address\":\"" + address + "\"},{\"port\":\"5001\",\"address\":\"" + address + "\"},{\"port\":\"5002\",\"address\":\"" + address + "\"}]")
	if result, err := expected.FromJSON(example); err != nil {
		t.Errorf("expected nil, got %v", err)
	} else if !reflect.DeepEqual(expected.peers, result.peers) {
		t.Errorf("expected %v, got %v", expected.peers, result.peers)
	}
}
