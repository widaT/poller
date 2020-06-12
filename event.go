package poller

import (
	"golang.org/x/sys/unix"
)

type Event unix.EpollEvent

func MakeEvents(length int) []Event {
	return make([]Event, length)
}

func (e Event) Token() Token {
	return Token(e.Pad)
}

func (e Event) IsReadable() bool {
	return (e.Events&unix.POLLIN) != 0 || (e.Events&unix.EPOLLPRI) != 0
}

func (e Event) IsWritable() bool {
	return (e.Events & unix.POLLOUT) != 0
}

func (e Event) IsError() bool {
	return (e.Events & unix.POLLERR) != 0
}

func (e Event) IsReadClosed() bool {
	// Both halves of the socket have closed
	return e.Events&unix.POLLHUP != 0 ||
		// Socket has received FIN or called shutdown(SHUT_RD)
		(e.Events&unix.POLLIN != 0 &&
			e.Events&unix.POLLRDHUP != 0)
}

func (e Event) IsWriteClosed() bool {
	// Both halves of the socket have closed
	return e.Events&unix.POLLHUP != 0 ||
		// Unix pipe write end has closed
		(e.Events&unix.POLLOUT != 0 &&
			e.Events&unix.POLLERR != 0)
}

func (e Event) IsPriority(event *Event) bool {
	return (e.Events & unix.POLLPRI) != 0
}
