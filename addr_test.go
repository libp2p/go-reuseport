package reuseport

import (
	"testing"
)

type netAddr struct {
	address string
	network string
}

func TestResolveAddr(t *testing.T) {
	netAddrs := []netAddr{
		netAddr{"127.0.0.1", "ip"}, netAddr{"127.0.0.1", "ip4"},
		netAddr{"::1", "ip6"}, netAddr{"127.0.0.1:1234", "tcp"},
		netAddr{"127.0.0.1:1234", "tcp4"}, netAddr{"[::1]:1234", "tcp6"},
		netAddr{"127.0.0.1:1234", "udp"}, netAddr{"127.0.0.1:1234", "udp4"},
		netAddr{"[::1]:1234", "udp6"}, netAddr{"127.0.0.1:1234", "unix"},
		netAddr{"127.0.0.1:1234", "unixgram"}, netAddr{"127.0.0.1:1234", "unixpacket"},
	}
	for _, na := range netAddrs {
		_, err := ResolveAddr(na.network, na.address)
		if err != nil {
			t.Errorf("Failed to resolve address %v", err)
		}
	}
}
