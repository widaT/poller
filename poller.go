package poller

import (
	"log"

	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"golang.org/x/sys/unix"
)

type Poller struct {
	waker    *Waker
	selector *Selector
	tasks    []func()
	lock     *Locker
}

func NewPoller() (p *Poller, err error) {
	p = new(Poller)
	p.selector, err = NewSelector()
	if err != nil {
		return
	}
	p.lock = &Locker{}
	p.waker, err = NewWaker(p.selector, WakerToken)
	if err != nil {
		return
	}
	return
}

func (p *Poller) AddTask(fn func()) {
	p.lock.Lock()
	p.tasks = append(p.tasks, fn)
	p.lock.Unlock()
}

func (p *Poller) Wake() error {
	return p.waker.Wake()
}

func (p *Poller) Polling(f func(*Event) error) error {
	events := MakeEvents(DefaultEventLen)
	wake := false
	for {
		n, err := p.selector.Select(events, -1)
		if err != nil && err != unix.EINTR {
			log.Println(err)
			continue
		}
		for i := 0; i < n; i++ {
			ev := events[i]
			switch ev.Token() {
			case WakerToken:
				p.waker.Reset()
				wake = true
			default:
				if err := f(&ev); err != nil {
					return err
				}
			}
		}
		if wake {
			wake = false
			p.runTask()
		}
		if n == len(events) {
			events = MakeEvents(n << 1)
		}
	}
}

func (p *Poller) runTask() {
	p.lock.Lock()
	tasks := p.tasks
	p.tasks = nil
	p.lock.Unlock()
	length := len(tasks)
	for i := 0; i < length; i++ {
		tasks[i]()
	}
}

func (p *Poller) Register(fd int, token Token, interests interest.Interest, opt pollopt.PollOpt) error {
	return p.selector.Register(fd, token, interests, opt)
}

func (p *Poller) Reregister(fd int, token Token, interests interest.Interest, opt pollopt.PollOpt) error {
	return p.selector.Reregister(fd, token, interests, opt)
}

func (p *Poller) Deregister(fd int) error {
	return p.selector.Deregister(fd)
}
