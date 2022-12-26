package peer

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func getExamples(n int) (map[*Peer]chan []byte, *Peer) {
	examples := map[*Peer]chan []byte{}
	if n <= 0 {
		return examples, nil
	}

	first, _ := Me(5000, false)
	examples[first] = nil
	for i := 1; i < n; i++ {
		example, _ := Me(5000+i, false)
		examples[example] = nil
	}

	return examples, first
}

func TestNotInitializedMembers(t *testing.T) {
	c := qt.New(t)

	defer func() {
		c.Assert(recover(), qt.IsNotNil)
	}()

	// not panic
	result := NewMembers()
	result.Len()

	// panic
	result = new(Members)
	result.Len()
}

func TestNewMembers(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	c.Assert(result.peers, qt.DeepEquals, map[*Peer]chan []byte{})
	c.Assert(result.mutex, qt.IsNotNil)
}

func TestMembersPeers(t *testing.T) {
	c := qt.New(t)

	members := NewMembers()
	c.Assert(members.Peers(), qt.HasLen, 0)

	expected, _ := getExamples(3)
	for member := range expected {
		members.Append(member)
	}

	c.Assert(expected, qt.ContentEquals, members.Peers())
}

func TestMembersLen(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	if result.Len() != 0 {
		t.Errorf("expected 0, got %d", result.Len())
	}

	expected, first := getExamples(3)
	for member := range expected {
		result.Append(member)
	}
	c.Assert(expected, qt.HasLen, result.Len())

	result.Delete(first)
	delete(expected, first)
	c.Assert(expected, qt.HasLen, result.Len())
}

func TestMembersAppend(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	expected, _ := getExamples(3)
	for member := range expected {
		result.Append(member)
		_, included := result.peers[member]
		c.Assert(included, qt.IsTrue)
	}

	c.Assert(expected, qt.HasLen, result.Len())
}

func TestMembersDelete(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	expected, first := getExamples(3)
	for member := range expected {
		result.Append(member)
	}

	result.Delete(first)
	delete(expected, first)
	c.Assert(expected, qt.DeepEquals, result.peers)
}

func TestMembersContains(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	expected, first := getExamples(3)
	for member := range expected {
		result.Append(member)
	}

	c.Assert(result.Contains(first), qt.IsTrue)
	needle, _ := Me(5003, false)
	c.Assert(result.Contains(needle), qt.IsFalse)
}

func TestMembersToJSON(t *testing.T) {
	c := qt.New(t)

	examples, first := getExamples(3)
	address := first.Address
	expected := []byte("[{\"port\":5000,\"address\":\"" + address + "\",\"type\":\"FULL\"}]")
	members := NewMembers()
	members.Append(first)

	result, err := members.ToJSON()
	c.Assert(err, qt.IsNil)
	c.Assert(expected, qt.DeepEquals, result)

	members = NewMembers()
	for example := range examples {
		members.Append(example)
	}

	strExpected := []string{
		"{\"port\":5000,\"address\":\"" + address + "\",\"type\":\"FULL\"}",
		"{\"port\":5001,\"address\":\"" + address + "\",\"type\":\"FULL\"}",
		"{\"port\":5002,\"address\":\"" + address + "\",\"type\":\"FULL\"}",
	}
	result, err = members.ToJSON()
	c.Assert(err, qt.IsNil)
	for _, str := range strExpected {
		c.Assert(string(result), qt.Contains, str)
	}
}

func TestMembersFromJSON(t *testing.T) {
	c := qt.New(t)

	examples, first := getExamples(3)
	address := first.Address
	example := []byte("[{\"port\":5000,\"address\":\"" + address + "\"}]")
	expected := NewMembers()
	expected.Append(first)

	result, err := expected.FromJSON(example)
	c.Assert(err, qt.IsNil)
	c.Assert(expected.peers, qt.DeepEquals, result.peers)

	expected = NewMembers()
	for p := range examples {
		expected.Append(p)
	}

	example = []byte("[{\"port\":5000,\"address\":\"" + address + "\"},{\"port\":5001,\"address\":\"" + address + "\"},{\"port\":5002,\"address\":\"" + address + "\"}]")
	result, err = expected.FromJSON(example)
	c.Assert(err, qt.IsNil)
	c.Assert(expected.peers, qt.DeepEquals, result.peers)
}
