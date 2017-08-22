// +build windows plan9 ios

package poll

import (
	"errors"
)

func WaitWrite(fd int) error {
	return errors.New("platform not supported")
}
