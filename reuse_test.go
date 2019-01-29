package reuseport

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func testDialFromListeningPort(t *testing.T, network, host string) {
	lc := net.ListenConfig{
		Control: Control,
	}
	ctx := context.Background()
	ll, err := lc.Listen(ctx, network, host+":0")
	if err != nil && strings.Contains(err.Error(), "cannot assign requested address") {
		t.Skip(err)
	}
	require.NoError(t, err)
	rl, err := lc.Listen(ctx, network, host+":0")
	require.NoError(t, err)
	d := net.Dialer{
		LocalAddr: ll.Addr(),
		Control:   Control,
	}
	c, err := d.Dial(network, rl.Addr().String())
	require.NoError(t, err)
	c.Close()
}

func TestDialFromListeningPort(t *testing.T) {
	testDialFromListeningPort(t, "tcp", "localhost")
}

func TestDialFromListeningPortTcp6(t *testing.T) {
	testDialFromListeningPort(t, "tcp6", "[::1]")
}

func TestListenPacketWildcardAddress(t *testing.T) {
	pc, err := ListenPacket("udp", ":0")
	require.NoError(t, err)
	pc.Close()
}

func TestErrorWhenDialUnresolvable(t *testing.T) {
	_, err := Dial("asd", "127.0.0.1:1234", "127.0.0.1:1234")
	assert.IsType(t, net.UnknownNetworkError(""), errors.Cause(err))
	_, err = Dial("tcp", "a.b.c.d:1234", "a.b.c.d:1235")
	assert.Error(t, err)
}
