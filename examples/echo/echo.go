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
	events := poller.MakeEvents(128)
	connections := make(map[poller.Token]int)
	for {
		n, err := poll.Select(events, -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			log.Fatal(err)
		}
		for i := 0; i < n; i++ {
			ev := events[i]
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
						log.Fatal(err)
					}
					if err := poller.Nonblock(cfd); err != nil {
						log.Fatal(err)
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
					err := handleEvent(poll, fd, ev)
					if err != nil {
						delete(connections, poller.Token(uniqueToken))
					}
				}
			}
		}
	}
}

func handleEvent(s *poller.Selector, fd int, event poller.Event) error {
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
