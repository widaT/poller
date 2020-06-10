package main

import (
	"fmt"
	"log"
	"net"

	"github.com/widaT/poller"
	"github.com/widaT/poller/interest"
	"github.com/widaT/poller/pollopt"
	"golang.org/x/sys/unix"
)

const SERVER poller.Token = poller.Token(0)

// run it
// then nc localhost 9999
func main() {
	poll, err := poller.New()
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal(err)
	}

	fd, err := poller.Listener2Fd(ln, true)
	if err != nil {
		log.Fatal(err)
	}
	err = poll.Register(fd, SERVER, interest.READABLE, pollopt.Edge)
	if err != nil {
		log.Fatal(err)
	}

	uniqueToken := 1
	connections := make(map[poller.Token]int)

	fn := func(ev *poller.Event) error {
		switch ev.Token() {
		case SERVER:
			for {
				cfd, _, err := unix.Accept(fd)
				if err != nil {
					//WouldBlock
					if err == unix.EAGAIN {
						//	fmt.Println(err)
						break
					}
					return err
				}
				if err := poller.Nonblock(cfd); err != nil {
					return err
				}

				uniqueToken++
				err = poll.Register(cfd, poller.Token(uniqueToken), interest.READABLE.Add(interest.WRITABLE), pollopt.Edge)
				if err != nil {
					log.Fatal(err)
				}
				connections[poller.Token(uniqueToken)] = cfd
			}

		default:
			if fd, found := connections[ev.Token()]; found {
				err := handle(poll, fd, ev)
				if err != nil {
					delete(connections, poller.Token(uniqueToken))
				}
			}
		}
		return nil
	}
	poller.Polling(poll, fn)
}

func handle(s *poller.Selector, fd int, event *poller.Event) error {
	switch {
	case event.IsReadable():
		connectionClosed := false
		receivedData := make([]byte, 4096)
		for {
			buf := make([]byte, 256)
			n, err := unix.Read(fd, buf)
			if n == 0 {
				connectionClosed = true
			}
			if err != nil {
				//WouldBlock
				if err == unix.EAGAIN {
					break
				}
				//Interrupted
				if err == unix.EINTR {
					continue
				}
				return err
			}
			receivedData = append(receivedData, buf[:n]...)
		}

		fmt.Println(string(receivedData))
		if connectionClosed {
			fmt.Println("Connection closed")
			return nil
		}
		unix.Write(fd, receivedData)
	}
	return nil
}
