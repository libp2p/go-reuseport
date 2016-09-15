// +build linux

package singlepoll

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"syscall"

	"github.com/sahne/eventfd"
)

var (
	ErrUnsupportedMode error = errors.New("only 'w' and 'r' modes are supported on this arch")
)

var (
	initOnce sync.Once
	workChan chan interface{}
)

type addPoll struct {
	fd     int
	events uint32
	ctx    context.Context
	wakeUp chan<- error
}

type ctxDone struct {
	fd int
}

func PollPark(ctx context.Context, fd int, mode string) error {
	initOnce.Do(func() {
		workChan = make(chan interface{}, 128)
		go worker()
	})

	events := uint32(0)
	for _, c := range mode {
		switch c {
		case 'w':
			events |= syscall.EPOLLOUT
		case 'r':
			events |= syscall.EPOLLIN
		default:
			return ErrUnsupportedMode
		}

	}

	wakeUp := make(chan error)
	workChan <- addPoll{
		fd:     fd,
		events: events,
		ctx:    ctx,
		wakeUp: wakeUp,
	}

	return <-wakeUp
}

func worker() {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	epfd, err := syscall.EpollCreate1(0)
	evfd, err := eventfd.New()
	_ = err

	pool := make(map[int]addPoll)

	{
		event := syscall.EpollEvent{
			Events: syscall.EPOLLIN,
			Fd:     int32(evfd.Fd()),
		}
		syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, evfd.Fd(), &event)
		runtime.KeepAlive(event) // won't hurt but might fix some possible issues with GC
	}
	go poller(epfd, evfd)

	remove := func(fd int) *addPoll {
		unit, ok := pool[fd]
		if ok {
			syscall.EpollCtl(epfd, syscall.EPOLL_CTL_DEL, unit.fd, nil)
			delete(pool, fd)
			close(unit.wakeUp)
		}
		return &unit
	}
	for {
		select {
		case <-ctx.Done():
			evfd.WriteEvents(1)
			return
		case unit := <-workChan:
			switch u := unit.(type) {
			case addPoll:
				event := syscall.EpollEvent{
					Events: u.events,
					Fd:     int32(u.fd),
				}

				// Make copies for *I* before we add it to Epoll group
				wrapWakeUp := make(chan error)
				wakeUp := u.wakeUp
				u.wakeUp = wrapWakeUp

				pool[u.fd] = u

				err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, u.fd, &event)
				if err != nil {
					delete(pool, u.fd)
					u.wakeUp <- err
				}

				// *I*
				reqCtx := u.ctx
				fd := u.fd

				go func() {
					select {
					case err := <-wrapWakeUp:
						wakeUp <- err
					case <-reqCtx.Done():
						workChan <- ctxDone{
							fd: fd,
						}
						wakeUp <- reqCtx.Err()
					}
				}()

			case []syscall.EpollEvent:
				for _, event := range u {
					remove(int(event.Fd))
				}
			case ctxDone:
				remove(u.fd)
			}
		}
	}

}

func poller(epfd int, evfd *eventfd.EventFD) {
	for {
		// do not reuse the array as we will be passing it over channel
		events := make([]syscall.EpollEvent, 128)
		n, _ := syscall.EpollWait(epfd, events, -1)

		for i := 0; i < n; i++ {
			if int(events[n].Fd) == evfd.Fd() {
				syscall.Close(epfd)
				evfd.Close()
				return
			}
		}
		workChan <- events[:n]
	}
}
