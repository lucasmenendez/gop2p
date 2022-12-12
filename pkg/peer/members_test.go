package peer

import (
	"reflect"
	"sync"
	"testing"
)

func TestEmptyMembers(t *testing.T) {
	var expected, result = &Members{[]*Peer{}, &sync.Mutex{}}, EmptyMembers()
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestMembersPeers(t *testing.T) {

}

func TestMembersLen(t *testing.T) {

}

func TestMembersAppend(t *testing.T) {

}

func TestMembersDelete(t *testing.T) {

}

func TestMembersContains(t *testing.T) {

}

func TestMembersToJSON(t *testing.T) {

}

func TestMembersFromJSON(t *testing.T) {

}
