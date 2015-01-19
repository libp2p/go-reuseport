package reuseport

import (
	"errors"
	"net"

	resolve "github.com/jbenet/go-net-resolve-addr"
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

// Dial dials the given network and address. see net.Dialer.Dial
// Returns a net.Conn created from a file discriptor for a socket
// with SO_REUSEPORT and SO_REUSEADDR option set.
func Dial(network, laddr, raddr string) (net.Conn, error) {

	var d Dialer
	if laddr != "" {
		netladdr, err := resolve.ResolveAddr("dial", network, laddr)
		if err != nil {
			return nil, err
		}
		d.D.LocalAddr = netladdr
	}

	return dial(d.D, network, raddr)
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
