package peer

import (
	"fmt"
	"regexp"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNew(t *testing.T) {
	c := qt.New(t)

	expected := &Peer{Address: "localhost", Port: 5000}
	result, err := New("localhost", 5000)
	c.Assert(err, qt.IsNil)
	c.Assert(result, qt.ContentEquals, expected)

	_, err = New("", 5000)
	c.Assert(err, qt.IsNotNil)

	_, err = New("0.0.0.0", -1)
	c.Assert(err, qt.IsNotNil)
}

func TestMe(t *testing.T) {
	c := qt.New(t)

	_, err := Me(-1, false)
	c.Assert(err, qt.IsNotNil)
	_, err = Me(-1, true)
	c.Assert(err, qt.IsNotNil)

	addressRgx := regexp.MustCompile(`(((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}|localhost)`)
	result, _ := Me(5000, false)
	c.Assert(addressRgx.MatchString(result.Address), qt.IsTrue)

	result, err = Me(5000, true)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Address, qt.Equals, allAddresses)
}

func TestPeerEqual(t *testing.T) {
	c := qt.New(t)

	expected, _ := Me(5000, false)
	candidate := &Peer{Address: expected.Address, Port: expected.Port}
	c.Assert(expected.Equal(candidate), qt.IsTrue)
	candidate.Address = allAddresses
	c.Assert(expected.Equal(candidate), qt.IsFalse)
	candidate, _ = Me(5001, false)
	c.Assert(expected.Equal(candidate), qt.IsFalse)
}

func TestPeerString(t *testing.T) {
	me, _ := Me(5000, false)
	expected := fmt.Sprintf(baseString, me.Address, me.Port)
	if result := me.String(); expected != result {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestPeerHostname(t *testing.T) {
	me, _ := Me(5000, false)
	expected := fmt.Sprintf(baseHostname, me.Address, me.Port)
	if result := me.Hostname(); expected != result {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
