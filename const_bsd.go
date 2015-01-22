// +build darwin freebsd dragonfly netbsd openbsd

package reuseport

import (
	"syscall"
)

var soReusePort = syscall.SO_REUSEPORT
var soReuseAddr = syscall.SO_REUSEADDR

func Select(nfd int, r *syscall.FdSet, w *syscall.FdSet, e *syscall.FdSet, timeout *syscall.Timeval) (n int, err error) {
	err = syscall.Select(nfd, r, w, e, timeout)
	return 0, err
}
