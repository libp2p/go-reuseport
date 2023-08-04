//go:build !plan9 && !windows && !wasm

package reuseport

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// This value has been taken from https://reviews.freebsd.org/D11003 since this is not yet provided in golang.
const FREEBSD_SO_REUSEPORT_LB = 0x00010000

func Control(network, address string, c syscall.RawConn) (err error) {
	controlErr := c.Control(func(fd uintptr) {
		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err != nil {
			return
		}
		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		if err != nil {
			return
		}
		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, FREEBSD_SO_REUSEPORT_LB, 1)
	})
	if controlErr != nil {
		err = controlErr
	}
	return
}
