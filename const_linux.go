// +build linux

package reuseport

import (
	"golang.org/x/sys/unix"
)

var soReusePort = 15 // this is not defined in unix go pkg.
var soReuseAddr = unix.SO_REUSEADDR
