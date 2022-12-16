package peer

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

func TestNew(t *testing.T) {
	var expected = &Peer{Address: "localhost", Port: 5000}
	var result = New("localhost", 5000)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected %v, got %v", expected, result)
	} else if result = New("", 5000); result != nil {
		t.Errorf("expected nil, got %v", result)
	} else if result = New("localhost", 0); result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMe(t *testing.T) {
	var result = Me(0)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}

	var addressRgx = regexp.MustCompile(`((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}`)
	if result = Me(5000); !addressRgx.MatchString(result.Address) && result.Address != "localhost" {
		t.Errorf("expected a valid IP or 'localhost', got %s", result.Address)
	}
}

func TestPeerEqual(t *testing.T) {
	var expected = Me(5000)
	var candidate = &Peer{Address: expected.Address, Port: expected.Port}
	if !expected.Equal(candidate) {
		t.Errorf("expected %v == %v = true, got false", expected, candidate)
	} else if candidate.Address = "0.0.0.0"; expected.Equal(candidate) {
		t.Errorf("expected %v == %v = false, got true", expected, candidate)
	} else if candidate = Me(5001); expected.Equal(candidate) {
		t.Errorf("expected %v == %v = false, got true", expected, candidate)
	}
}

func TestPeerString(t *testing.T) {
	var me = Me(5000)
	var expected = fmt.Sprintf(baseString, me.Address, me.Port)
	if result := me.String(); expected != result {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestPeerHostname(t *testing.T) {
	var me = Me(5000)
	var expected = fmt.Sprintf(baseHostname, me.Address, me.Port)
	if result := me.Hostname(); expected != result {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
