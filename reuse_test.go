package reuseport

import (
	"testing"
)

var message = "Error %v when attempting to listen to 127.0.0.1"

func TestReusePortFeatureAvailable(t *testing.T) {
	if ok := Available(); !ok {
		t.Error("SO_REUSEPORT is not available on this OS")
	}
}

func TestListenOnSamePort(t *testing.T) {
	l1, err := Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Errorf(message+":1234", err)
	}
	l2, err := Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Errorf(message+":1234", err)
	}
	l1.Close()
	l2.Close()

	lp1, err := ListenPacket("udp", "127.0.0.1:1235")
	if err != nil {
		t.Errorf(message+":1234", err)
	}
	lp2, err := ListenPacket("udp", "127.0.0.1:1235")
	if err != nil {
		t.Errorf(message+":1234", err)
	}
	lp1.Close()
	lp2.Close()
}

func TestDialFromSamePort(t *testing.T) {
	_, err := Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		t.Errorf(message+":1234", err)
	}
	_, err = Listen("tcp", "127.0.0.1:1235")
	if err != nil {
		t.Errorf(message+":1235", err)
	}
	c, err := Dial("tcp", "127.0.0.1:1234", "127.0.0.1:1235")
	if err != nil {
		t.Errorf("Error %v when attempting to dial from 127.0.0.1:1234 to 127.0.0.1:1235", err)
	}
	c.Close()
}

func TestErrorWhenDialUnresolvable(t *testing.T) {
	_, err := Dial("asd", "127.0.0.1:1234", "127.0.0.1:1234")
	if err == nil {
		t.Error("Expected error when trying to dial an unknown protocol")
	}

	_, err = Dial("tcp", "a.b.c.d:1234", "a.b.c.d:1235")
	if err == nil {
		t.Error("Expected error when trying to dial an unknown address")
	}
}
