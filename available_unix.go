// +build darwin freebsd dragonfly netbsd openbsd linux

package reuseport

import (
	"golang.org/x/sys/unix"
	"sync"
	"syscall"
	"time"
)

var (
	hasReusePort bool
	didReusePort sync.Once
)

// Available returns whether or not SO_REUSEPORT is available in the OS.
// It does so by attepting to open a tcp listener, setting the option, and
// checking ENOPROTOOPT on error. After checking, the decision is cached
// for the rest of the process run.
func available() bool {
	didReusePort.Do(checkReusePort)
	return hasReusePort
}

func checkReusePort() {
	// there may be fluke reasons to fail to add a listener.
	// so we give it 5 shots. if not, give up and call it not avail.
	for i := 0; i < 5; i++ {
		// try to listen at tcp port 0.
		l, err := listenStream("tcp", "127.0.0.1:0")
		if err == nil {
			l.Close() // Go back to the Shadow!
			// no error? available.
			hasReusePort = true
			return
		}

		if errno, ok := err.(syscall.Errno); ok && errno == unix.ENOPROTOOPT {
			return // :( that's all folks.
		}

		// not an errno? or not ENOPROTOOPT? retry.
		time.Sleep(20 * time.Millisecond) // wait a bit
	}
}
