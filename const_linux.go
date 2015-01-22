// +build linux

package reuseport

import (
	"syscall"
)

var soReusePort = 15 // this is not defined in unix go pkg.
var soReuseAddr = syscall.SO_REUSEADDR

func Select(nfd int, r *syscall.FdSet, w *syscall.FdSet, e *syscall.FdSet, timeout *syscall.Timeval) (n int, err error) {
	return syscall.Select(nfd, r, w, e, timeout)
}
