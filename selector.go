package poller

import (
	"log"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"golang.org/x/sys/unix"
)

const POLLET uint32 = 1 << 31
const POLLONESHOT uint32 = 1 << 30
const DefaultEventLen = 128

type Token int32

const WakerToken = Token(0)

var token = Token(1)

func NextToken() (tk Token) {
	tk = token
	token++
	return
}

type Selector struct {
	id uint32
	fd int
}

var NextId uint32 = 1

func NewSelector() (selector *Selector, err error) {
	selector = new(Selector)
	selector.fd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	selector.id = atomic.AddUint32(&NextId, 1)
	return
}

func (s *Selector) Id() uint32 {
	return s.id
}

func (s *Selector) Fd() int {
	return s.fd
}

func (s *Selector) Select(events []Event, timeout int) (int, error) {
	n, err := epollWait(s.fd, events, timeout)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Selector) Register(fd int, token Token, interests interest.Interest, opt pollopt.PollOpt) error {
	return unix.EpollCtl(s.fd, syscall.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Events: interestsToEpoll(interests, opt), Fd: int32(fd), Pad: int32(token)})
}

func (s *Selector) Reregister(fd int, token Token, interests interest.Interest, opt pollopt.PollOpt) error {
	return unix.EpollCtl(s.fd, syscall.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Events: interestsToEpoll(interests, opt), Fd: int32(fd), Pad: int32(token)})
}

func (s *Selector) Deregister(fd int) error {
	return unix.EpollCtl(s.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func Polling(s *Selector, f func(*Event) error) error {
	events := MakeEvents(DefaultEventLen)
	for {
		n, err := s.Select(events, -1)
		if err != nil && err != unix.EINTR {
			log.Println(err)
			continue
		}
		for i := 0; i < n; i++ {
			ev := events[i]
			if err := f(&ev); err != nil {
				return err
			}
		}
		if n == len(events) {
			events = MakeEvents(n << 1)
		}
	}
}

func interestsToEpoll(interests interest.Interest, opts pollopt.PollOpt) uint32 {
	var kind uint32 = 0
	if interests.IsReadable() {
		kind = kind | unix.POLLIN | unix.POLLHUP
	}
	if interests.IsWritable() {
		kind |= unix.POLLOUT
	}

	if opts.IsEdge() {
		kind |= POLLET
	}

	if opts.IsOneshot() {
		kind |= POLLONESHOT
	}

	if opts.IsLevel() {
		kind &= ^POLLET
	}
	return kind
}

var _zero uintptr

func epollWait(epfd int, events []Event, msec int) (n int, err error) {
	var _p0 unsafe.Pointer
	if len(events) > 0 {
		_p0 = unsafe.Pointer(&events[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	r0, _, e1 := unix.Syscall6(unix.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(_p0), uintptr(len(events)), uintptr(msec), 0, 0)
	n = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case unix.EAGAIN:
		return syscall.EAGAIN
	case unix.EINVAL:
		return syscall.EINVAL
	case unix.ENOENT:
		return syscall.ENOENT
	}
	return e
}
