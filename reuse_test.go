package reuseport

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func testDialFromListeningPort(t *testing.T, network string) {
	lc := net.ListenConfig{
		Control: Control,
	}
	ctx := context.Background()
	ll, err := lc.Listen(ctx, network, "localhost:0")
	require.NoError(t, err)
	rl, err := lc.Listen(ctx, network, "localhost:0")
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
	testDialFromListeningPort(t, "tcp")
}

func TestDialFromListeningPortTcp6(t *testing.T) {
	testDialFromListeningPort(t, "tcp6")
}
