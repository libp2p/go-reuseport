package reuseport

import (
	"errors"
	"net"
)

// ErrUnsuportedProtocol signals that the protocol is not currently
// supported by this package. This package currently only supports TCP.
var ErrUnsupportedProtocol = errors.New("protocol not yet supported")

// ErrReuseFailed is returned if a reuse attempt was unsuccessful.
var ErrReuseFailed = errors.New("protocol not yet supported")

// Listen listens at the given network and address. see net.Listen
// Returns a net.Listener created from a file discriptor for a socket
// with SO_REUSEPORT and SO_REUSEADDR option set.
func Listen(network, address string) (net.Listener, error) {
	return listen(network, address)
}

// Dialer is used to specify the Dial options, much like net.Dialer.
// We simply wrap a net.Dialer.
type Dialer struct {
	D net.Dialer
}

// Dial dials the given network and address. see net.Dialer.Dial
// Returns a net.Conn created from a file discriptor for a socket
// with SO_REUSEPORT and SO_REUSEADDR option set.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return dial(d.D, network, address)
}
