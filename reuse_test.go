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
	l1, err := lc.Listen(ctx, network, "localhost:0")
	require.NoError(t, err)
	l2, err := lc.Listen(ctx, network, "localhost:0")
	require.NoError(t, err)
	d := net.Dialer{
		LocalAddr: l1.Addr(),
		Control:   Control,
	}
	c, err := d.Dial(network, l2.Addr().String())
	require.NoError(t, err)
	c.Close()
}

func TestDialFromListeningPort(t *testing.T) {
	testDialFromListeningPort(t, "tcp")
}
