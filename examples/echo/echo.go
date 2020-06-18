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

var SERVER = poller.NextToken()
var CLIENT = poller.NextToken()

// run it
// then nc localhost 9999
func main() {
	poll, err := poller.NewPoller()
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
				err = poll.Register(cfd, poller.Token(CLIENT), interest.READABLE.Add(interest.WRITABLE), pollopt.Edge)
				if err != nil {
					log.Fatal(err)
				}
			}
		case CLIENT:
			err := handle(poll, int(ev.Fd), ev)
			if err != nil {
			}
		}
		return nil
	}
	poll.Polling(fn)
}

func handle(s *poller.Poller, fd int, event *poller.Event) error {
	switch {
	case event.IsReadable():
		connectionClosed := false
		receivedData := make([]byte, 4096)
		for {
			buf := make([]byte, 256)
			n, err := unix.Read(fd, buf)
			if n == 0 {
				connectionClosed = true
				break
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
				//`connection reset by peer` n is not 0 ,err is not nil but we can't break the work,so can't use return
				log.Printf("%s", err)
				connectionClosed = true
				break
				//return err
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
