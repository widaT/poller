package poller

import (
	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"golang.org/x/sys/unix"
)

var _1u64 = []byte{1, 0, 0, 0, 0, 0, 0, 0}
var _0u64 = make([]byte, 8)

type Waker struct {
	fd int
}

func New(s *poller.Selector, token poller.Token) (waker *Waker, err error) {
	waker = new(Waker)
	if waker.fd, err = unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC); err != nil {
		return
	}
	if err = s.Register(waker.fd, token, interest.READABLE, pollopt.Edge); err != nil {
		return
	}
	return
}

func (w *Waker) Wake() error {
	_, err := unix.Write(w.fd, _1u64)
	if err != nil {
		if err == unix.EAGAIN {
			w.Reset()
			return w.Wake()
		}
		return err
	}
	return nil
}

func (w *Waker) Reset() error {
	_, err := unix.Read(w.fd, _0u64)
	if err != nil && err != unix.EAGAIN {
		return err
	}
	return nil
}
