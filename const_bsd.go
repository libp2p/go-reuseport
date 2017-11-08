// +build darwin freebsd dragonfly netbsd openbsd

package reuseport

import (
	"golang.org/x/sys/unix"
)

var soReusePort = unix.SO_REUSEPORT
var soReuseAddr = unix.SO_REUSEADDR
