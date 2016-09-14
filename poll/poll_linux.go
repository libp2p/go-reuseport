// +build linux

package poll

import (
	"context"
	"sync"
	"syscall"

	"github.com/sahne/eventfd"
)

type Poller struct {
	epfd int

	eventMain syscall.EpollEvent
	eventWait syscall.EpollEvent
	events    []syscall.EpollEvent

	wake      *eventfd.EventFD // Use eventfd to wakeup epoll
	wakeMutex sync.Mutex
}

func New(fd int) (p *Poller, err error) {
	p = &Poller{
		events: make([]syscall.EpollEvent, 32),
	}
	if p.epfd, err = syscall.EpollCreate1(0); err != nil {
		return nil, err
	}
	wake, err := eventfd.New()
	if err != nil {
		syscall.Close(p.epfd)
		return nil, err
	}
	p.wakeMutex.Lock()
	p.wake = wake
	p.wakeMutex.Unlock()

	p.eventMain.Events = syscall.EPOLLOUT
	p.eventMain.Fd = int32(fd)
	if err = syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_ADD, fd, &p.eventMain); err != nil {
		p.Close()
		return nil, err
	}

	// poll that eventfd can be read
	p.eventWait.Events = syscall.EPOLLIN
	p.eventWait.Fd = int32(wake.Fd())
	if err = syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_ADD, wake.Fd(), &p.eventWait); err != nil {
		p.wake.Close()
		p.Close()
		return nil, err
	}

	return p, nil
}

func (p *Poller) Close() error {
	p.wakeMutex.Lock()
	err1 := p.wake.Close()
	// set wake to nil to be sure that we won't call write on closed wake
	// it should naver happen but if semeone changes something this might show a bug
	p.wake = nil
	p.wakeMutex.Unlock()

	err2 := syscall.Close(p.epfd)
	if err1 != nil {
		return err1
	} else {
		return err2
	}
}

func (p *Poller) WaitWriteCtx(ctx context.Context) error {
	doneChan := make(chan struct{})
	defer close(doneChan)

	go func() {
		select {
		case <-doneChan:
			return
		case <-ctx.Done():
			select {
			case <-doneChan:
				// if we re done with this function do not write to p.wake
				// it might be already closed and the fd could be reopened for
				// different purpose
				return
			default:
			}
			p.wakeMutex.Lock()
			p.wake.Write([]byte{0, 0, 0, 0, 0, 0, 0, 1}) // send event to wake up epoll
			p.wakeMutex.Unlock()
			return
		}

	}()

	n, err := syscall.EpollWait(p.epfd, p.events, -1)
	if err != nil {
		return err
	}
	good := false
	for i := 0; i < n; i++ {
		ev := p.events[i]
		if ev.Fd == p.eventMain.Fd {
			good = true
		}
		if ev.Fd == p.eventWait.Fd {
			p.wakeMutex.Lock()
			p.wake.Read(make([]byte, 8)) // clear eventfd
			p.wakeMutex.Unlock()
		}
	}
	if good {
		return nil
	}
	if ctx.Err() == nil {
		panic("notification but no deadline, this should be impossible")
	}
	return ctx.Err()
}
