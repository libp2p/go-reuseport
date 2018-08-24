// +build linux darvin

package reuseport

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func boolint(b bool) int {
	if b {
		return 1
	}
	return 0
}

func setNoDelay(fd int, noDelay bool) error {
	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, boolint(noDelay)))
}

func setLinger(fd int, sec int) error {
	var l unix.Linger
	if sec >= 0 {
		l.Onoff = 1
		l.Linger = int32(sec)
	} else {
		l.Onoff = 0
		l.Linger = 0
	}
	return os.NewSyscallError("setsockopt", unix.SetsockoptLinger(fd, unix.SOL_SOCKET, unix.SO_LINGER, &l))
}

func Control(network, address string, c syscall.RawConn) error {
	var err error
	c.Control(func(fd uintptr) {
		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err != nil {
			return
		}

		err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
		if err != nil {
			return
		}
		err = setLinger(int(fd), 5)
		if err != nil {
			return
		}
		err = setNoDelay(int(fd), true)
		if err != nil {
			return
		}
	})
	return err
}
