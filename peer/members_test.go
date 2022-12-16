package peer

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func getExamples(n int) []*Peer {
	examples := []*Peer{}
	if n <= 0 {
		return examples
	}

	for i := 0; i < n; i++ {
		example, _ := Me(5000+i, false)
		examples = append(examples, example)
	}

	return examples
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
	c.Assert(result.peers, qt.DeepEquals, []*Peer{})
	c.Assert(result.mutex, qt.IsNotNil)
}

func TestMembersPeers(t *testing.T) {
	c := qt.New(t)

	members := NewMembers()
	c.Assert(members.Peers(), qt.HasLen, 0)

	expected := getExamples(3)
	for _, member := range expected {
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

	expected := getExamples(3)
	for _, member := range expected {
		result.Append(member)
	}
	c.Assert(expected, qt.HasLen, result.Len())

	result.Delete(expected[0])
	expected = expected[1:]
	c.Assert(expected, qt.HasLen, result.Len())
}

func TestMembersAppend(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	expected := getExamples(3)
	for i, member := range expected {
		result.Append(member)
		c.Assert(member, qt.ContentEquals, result.peers[i])
	}

	c.Assert(expected, qt.HasLen, result.Len())
}

func TestMembersDelete(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	expected := getExamples(3)
	for _, member := range expected {
		result.Append(member)
	}

	result.Delete(expected[0])
	expected = expected[1:]
	c.Assert(expected, qt.DeepEquals, result.peers)
}

func TestMembersContains(t *testing.T) {
	c := qt.New(t)

	result := NewMembers()
	expected := getExamples(3)
	for _, member := range expected {
		result.Append(member)
	}

	c.Assert(result.Contains(expected[0]), qt.IsTrue)
	needle, _ := Me(5003, false)
	c.Assert(result.Contains(needle), qt.IsFalse)
}

func TestMembersToJSON(t *testing.T) {
	c := qt.New(t)

	examples := getExamples(3)
	address := examples[0].Address
	expected := []byte("[{\"port\":5000,\"address\":\"" + address + "\"}]")
	members := NewMembers()
	members.Append(examples[0])

	result, err := members.ToJSON()
	c.Assert(err, qt.IsNil)
	c.Assert(expected, qt.DeepEquals, result)

	for _, example := range examples[1:] {
		members.Append(example)
	}

	expected = []byte("[{\"port\":5000,\"address\":\"" + address + "\"},{\"port\":5001,\"address\":\"" + address + "\"},{\"port\":5002,\"address\":\"" + address + "\"}]")
	result, err = members.ToJSON()
	c.Assert(err, qt.IsNil)
	c.Assert(expected, qt.DeepEquals, result)
}

func TestMembersFromJSON(t *testing.T) {
	c := qt.New(t)

	examples := getExamples(3)
	address := examples[0].Address
	example := []byte("[{\"port\":5000,\"address\":\"" + address + "\"}]")
	expected := NewMembers()
	expected.Append(examples[0])

	result, err := expected.FromJSON(example)
	c.Assert(err, qt.IsNil)
	c.Assert(expected.peers, qt.DeepEquals, result.peers)

	for _, p := range examples[1:] {
		expected.Append(p)
	}

	example = []byte("[{\"port\":5000,\"address\":\"" + address + "\"},{\"port\":5001,\"address\":\"" + address + "\"},{\"port\":5002,\"address\":\"" + address + "\"}]")
	result, err = expected.FromJSON(example)
	c.Assert(err, qt.IsNil)
	c.Assert(expected.peers, qt.DeepEquals, result.peers)
}
